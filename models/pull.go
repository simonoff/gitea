package models

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Unknwon/com"
	"github.com/go-gitea/gitea/modules/process"
)

// __________      .__  .__ __________                                     __
// \______   \__ __|  | |  |\______   \ ____  ________ __   ____   _______/  |_
//  |     ___/  |  \  | |  | |       _// __ \/ ____/  |  \_/ __ \ /  ___/\   __\
//  |    |   |  |  /  |_|  |_|    |   \  ___< <_|  |  |  /\  ___/ \___ \  |  |
//  |____|   |____/|____/____/____|_  /\___  >__   |____/  \___  >____  > |__|
//                                  \/     \/   |__|           \/     \/

var ErrPullRequestNotExist = errors.New("Pull request does not exist")

// PullRepo represents relation between pull request and repositories.
type PullRepo struct {
	ID           int64 `xorm:"pk autoincr"`
	PullID       int64 `xorm:"unique"` // issueID
	FromRepoID   int64
	*Issue       `xorm:"-"`
	FromBranch   string
	ToBranch     string
	CanAutoMerge bool
	IsMerged     bool
	BeforeCommit string
	AfterCommit  string
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
}

func GetRepoPullByIssueID(issueID int64) (*PullRepo, error) {
	var pull = PullRepo{
		PullID: issueID,
	}
	has, err := x.Get(&pull)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrIssueNotExist
	}
	return &pull, nil
}

// NewPullRequest creates new pull request for repository.
func NewPullRequest(pr *Issue, pullRepo *PullRepo) error {
	toRepo, err := GetRepositoryById(pr.RepoID)
	if err != nil {
		return err
	}
	fromRepo, err := GetRepositoryById(pullRepo.FromRepoID)
	if err != nil {
		return err
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.InsertOne(pr); err != nil {
		return err
	} else if _, err = sess.Exec("UPDATE `repository` SET num_pulls=num_pulls+1 WHERE id=?", pr.RepoID); err != nil {
		return err
	}

	pullRepo.PullID = pr.ID

	// Clone target repository.
	toRepoPath, err := toRepo.RepoPath()
	if err != nil {
		return err
	}
	tmpRepoPath := path.Join(toRepoPath, "pulls", com.ToStr(pr.ID)+".git")
	_, stderr, err := process.ExecTimeout(10*time.Minute,
		fmt.Sprintf("NewPullRequest(clone target repository): %d", fromRepo.ID),
		"git", "clone", toRepoPath, tmpRepoPath)
	if err != nil {
		return fmt.Errorf("git clone: ", stderr)
	}
	defer os.RemoveAll(tmpRepoPath)

	// Checkout a temporary branch.
	tmpBranch := com.ToStr(time.Now().UnixNano())
	_, stderr, err = process.ExecDir(-1, tmpRepoPath,
		fmt.Sprintf("NewPullRequest(checkout temporary branch): %d", fromRepo.ID),
		"git", "checkout", "-b", tmpBranch)
	if err != nil {
		return fmt.Errorf("git checkout -b: ", stderr)
	}

	// Pull downstream code.
	fromRepoPath, err := fromRepo.RepoPath()
	if err != nil {
		return err
	}
	_, stderr, err = process.ExecDir(10*time.Minute, tmpRepoPath,
		fmt.Sprintf("NewPullRequest(pull downstream code): %d", fromRepo.ID),
		"git", "pull", fromRepoPath, pullRepo.FromBranch)
	if err != nil {
		if strings.Contains(stderr, "fatal:") {
			return fmt.Errorf("git pull: %s", stderr)
		}
	} else {
		pullRepo.CanAutoMerge = true
	}

	if pullRepo.CanAutoMerge {
		// Generate patch.
		var patch string
		patch, stderr, err = process.ExecDir(10*time.Minute, tmpRepoPath,
			fmt.Sprintf("NewPullRequest(generate patch): %d", fromRepo.ID),
			"git", "diff", "-p", pullRepo.ToBranch, tmpBranch)
		if err != nil {
			return fmt.Errorf("git diff -p: ", stderr)
		}

		patchPath := path.Join(toRepoPath, "pulls", com.ToStr(pr.ID)+".patch")
		if err := ioutil.WriteFile(patchPath, []byte(patch), os.ModePerm); err != nil {
			return fmt.Errorf("write patch: %v", err)
		}
	}

	if _, err = sess.InsertOne(pullRepo); err != nil {
		return err
	}

	return sess.Commit()
}

func GetPullRequest(fromRepoID int64) (*Issue, error) {
	pullRepo := &PullRepo{FromRepoID: fromRepoID}
	has, err := x.Get(pullRepo)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrPullRequestNotExist
	}

	pr := new(Issue)
	has, err = x.Id(pullRepo.PullID).Get(pr)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrIssueNotExist
	}
	return pr, nil
}
