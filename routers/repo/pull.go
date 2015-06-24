// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-gitea/gitea/models"
	"github.com/go-gitea/gitea/modules/auth"
	"github.com/go-gitea/gitea/modules/base"
	"github.com/go-gitea/gitea/modules/git"
	"github.com/go-gitea/gitea/modules/middleware"
	"github.com/go-gitea/gitea/modules/setting"
)

const (
	PULLS base.TplName = "repo/pull/list"
	PULL  base.TplName = "repo/pull/pull"
	//PULLS    base.TplName = "repo/pulls"
	//NEW_PULL base.TplName = "repo/pull_new"
)

func Pulls(ctx *middleware.Context) {
	issues(ctx, PULLS, true)
}

// view pull request /repopath/pull/:id
func Pull(ctx *middleware.Context) {
	ctx.Data["IsRepoToolbarPulls"] = true

	repo := ctx.Repo.Repository
	repoLink, _ := repo.RepoLink()
	ctx.Data["RepoLink"] = repoLink

	issueIndex := ctx.ParamsInt64(":id")
	issue, err := models.GetIssueByIndex(repo.ID, issueIndex)
	if err != nil {
		ctx.Handle(500, "GetIssueById", err)
		return
	}
	issueID := issue.ID

	err = issue.GetPoster()
	if err != nil {
		ctx.Handle(500, "GetIssueById", err)
		return
	}

	ctx.Data["Issue"] = issue

	pull, err := models.GetRepoPullByIssueID(issueID)
	if err != nil {
		ctx.Handle(500, "GetRepoPullByIssueId", err)
		return
	}
	ctx.Data["Pull"] = pull

	comments, err := models.GetIssueComments(issueID)
	if err != nil {
		ctx.Handle(500, "GetIssueComments", err)
		return
	}
	ctx.Data["Comments"] = comments
	ctx.Data["CountComments"] = len(comments) + 1

	fromRepo, err := models.GetRepositoryById(pull.FromRepoID)
	if err != nil {
		ctx.Handle(500, "GetRepositoryById", err)
		return
	}
	fromRepo.GetOwner()

	beforeRepoPath, err := repo.RepoPath()
	if err != nil {
		ctx.Handle(500, "GetRepositoryById", err)
		return
	}

	beforeRepo, err := git.OpenRepository(beforeRepoPath)
	if err != nil {
		ctx.Handle(404, "OpenRepository", err)
		return
	}

	afterRepoPath, err := fromRepo.RepoPath()
	if err != nil {
		ctx.Handle(500, "GetRepositoryById", err)
		return
	}

	afterRepo, err := git.OpenRepository(afterRepoPath)
	if err != nil {
		ctx.Handle(404, "OpenRepository", err)
		return
	}

	commit, err := beforeRepo.GetCommitOfBranch(pull.ToBranch)
	if err != nil {
		ctx.Handle(404, "GetCommit", err)
		return
	}

	diff, err := models.GetDiffForkedRange(beforeRepoPath, afterRepoPath,
		pull.ToBranch, pull.FromBranch, setting.Git.MaxGitDiffLines)
	if err != nil {
		ctx.Handle(404, "GetDiffRange", err)
		return
	}

	isImageFile := func(name string) bool {
		blob, err := commit.GetBlobByPath(name)
		if err != nil {
			return false
		}

		dataRc, err := blob.Data()
		if err != nil {
			return false
		}
		buf := make([]byte, 1024)
		n, _ := dataRc.Read(buf)
		if n > 0 {
			buf = buf[:n]
		}
		_, isImage := base.IsImageFile(buf)
		return isImage
	}

	commit, err = afterRepo.GetCommitOfBranch(pull.FromBranch)
	if err != nil {
		ctx.Handle(500, "CommitsBeforeUntil", err)
		return
	}

	commits, err := commit.CommitsBeforeUntil(commit.Id.String())
	if err != nil {
		ctx.Handle(500, "CommitsBeforeUntil", err)
		return
	}
	commits = models.ValidateCommitsWithEmails(commits)

	ctx.Data["Commits"] = commits
	ctx.Data["CommitCount"] = commits.Len()
	ctx.Data["Commit"] = commit
	ctx.Data["Diff"] = diff
	ctx.Data["IsImageFile"] = isImageFile
	ctx.Data["FromRepo"] = fromRepo
	ctx.Data["User"] = ctx.User

	ctx.HTML(200, PULL)
}

func hasPullRequested(ctx *middleware.Context, repoID int64, forkRepo *models.Repository) bool {
	pr, err := models.GetPullRequest(repoID)
	if err != nil {
		if err != models.ErrPullRequestNotExist {
			ctx.Handle(500, "GetPullRequest", err)
			return true
		}
	} else {
		repoLink, err := forkRepo.RepoLink()
		if err != nil {
			ctx.Handle(500, "RepoLink", err)
		} else {
			ctx.Redirect(fmt.Sprintf("%s/pull/%d", repoLink, pr.Index))
		}
		return true
	}
	return false
}

// FIXME: check if branch exists
func NewPullRequest(ctx *middleware.Context, form auth.NewPullRequestForm) {
	repo := ctx.Repo.Repository

	froms := strings.Split(form.FromBranch, ":")
	tos := strings.Split(form.ToBranch, ":")
	if len(froms) != 2 && len(tos) != 2 {
		ctx.Handle(500, "Params error", errors.New("should have repo and branch names"))
		return
	}

	fromRepo, err := models.GetRepositoryByFork(repo.ID, froms[0])
	if err != nil {
		ctx.Handle(500, "GetForkRepo", err)
		return
	}
	fromRepoID := fromRepo.ID

	var toRepoID int64
	if repo.Owner.Name == tos[0] {
		toRepoID = repo.ID
	} else {
		toRepo, err := models.GetRepositoryByFork(repo.ID, tos[0])
		if err != nil {
			ctx.Handle(500, "GetForkRepo", err)
			return
		}
		toRepoID = toRepo.ID
	}

	pullRepo := &models.PullRepo{
		FromRepoID: fromRepoID,
		FromBranch: froms[1],
		ToRepoID:   toRepoID,
		ToBranch:   tos[1],
	}

	has, err := models.IsPullRequestOpened(pullRepo)
	if err != nil {
		ctx.Handle(500, "GetByExample", err)
		return
	}

	repoLink, err := repo.RepoLink()
	if err != nil {
		ctx.Handle(500, "RepoLink", err)
		return
	}

	if has {
		ctx.Redirect(fmt.Sprintf("%s/pull/%d", repoLink, pullRepo.Index))
		return
	}

	pr := &models.Issue{
		RepoID:   repo.ID,
		Index:    int64(repo.NumPulls) + 1,
		Name:     form.Title,
		PosterID: ctx.User.Id,
		IsPull:   true,
		Content:  form.Description,
	}

	if err := models.NewPullRequest(pr, pullRepo); err != nil {
		ctx.Handle(500, "NewPullRequest", err)
		return
	} else if err := models.NewIssueUserPairs(repo, pr.ID, repo.OwnerID,
		ctx.User.Id, 0); err != nil {
		ctx.Handle(500, "NewIssueUserPairs", err)
		return
	}

	// FIXME: add action
	ctx.Redirect(fmt.Sprintf("%s/pull/%d", repoLink, pr.Index))
}

func PullComment(ctx *middleware.Context) {
	repoID := ctx.Repo.Repository.ID
	userID := ctx.User.Id
	issueID := ctx.QueryInt64("issueID")
	issueIndex := ctx.QueryInt("issueIndex")
	content := ctx.Query("content")
	submit := ctx.Query("submit")

	var err error
	if submit == "comment" {
		_, err = models.CreateComment(userID, repoID, issueID, "", "",
			models.COMMENT_TYPE_COMMENT, content, nil)
	} else if submit == "close" {
		_, err = models.CreateComment(userID, repoID, issueID, "", "",
			models.COMMENT_TYPE_CLOSE, content, nil)
	}
	if err != nil {
		ctx.Handle(500, "CreateComment", err)
		return
	}

	repoLink, err := ctx.Repo.Repository.RepoLink()
	if err != nil {
		ctx.Handle(500, "RepoLink", err)
		return
	}

	ctx.Redirect(fmt.Sprintf("%s/pull/%d", repoLink, issueIndex))
}

// merge pulls
func PullMerge(ctx *middleware.Context) {
	//issueID := ctx.ParamsInt64(":id")
	//repo := ctx.Repo.Repository
	/*pull, err := models.NewRepoPull(repo.Id, issueID)
	if err != nil {
		ctx.Handle(500, "GetPullRepoById", err)
		return
	}*/

	ctx.Redirect(ctx.Repo.RepoLink)
}
