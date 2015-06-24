package repo

import (
	"path"
	"strings"

	"github.com/go-gitea/gitea/models"
	"github.com/go-gitea/gitea/modules/base"
	"github.com/go-gitea/gitea/modules/git"
	"github.com/go-gitea/gitea/modules/middleware"
	"github.com/go-gitea/gitea/modules/setting"
)

var (
	PULL_COMPARE base.TplName = "repo/pull/compare"
)

func ForkDiff(ctx *middleware.Context, beforeCommitId, afterCommitId string) {
	ctx.Data["IsRepoToolbarCommits"] = true
	ctx.Data["IsDiffCompare"] = true
	userName := ctx.Repo.Owner.Name
	repoName := ctx.Repo.Repository.Name

	err := ctx.Repo.Repository.GetForks()
	if err != nil {
		ctx.Handle(500, "GetForks ForkId", err)
		return
	}

	var beforeRepo, afterRepo *models.Repository
	var beforeBranch, afterBranch string
	if strings.Contains(beforeCommitId, ":") {
		s := strings.Split(beforeCommitId, ":")
		for _, fork := range ctx.Repo.Repository.Forks {
			if s[0] == fork.Owner.Name {
				beforeRepo = fork
				break
			}
		}
		if beforeRepo == nil {
			ctx.Handle(500, "GetRepositoryByRef", err)
			return
		}
		beforeBranch = s[1]
	} else {
		beforeRepo = ctx.Repo.Repository
		beforeBranch = beforeCommitId
	}

	repoLink, _ := beforeRepo.RepoLink()

	beforeRepoPath, err := beforeRepo.RepoPath()
	if err != nil {
		ctx.Handle(500, "GetRepositoryByRef", err)
		return
	}
	beforeRep, err := git.OpenRepository(beforeRepoPath)
	if err != nil {
		ctx.Handle(500, "GetRepositoryByRef", err)
		return
	}

	beforeCommit, err := beforeRep.GetCommitOfBranch(beforeBranch)
	if err != nil {
		ctx.Handle(404, "GetCommit", err)
		return
	}
	beforeCommitId = beforeCommit.Id.String()

	beforeBranches, _ := beforeRep.GetBranches()

	if strings.Contains(afterCommitId, ":") {
		s := strings.Split(afterCommitId, ":")
		for _, fork := range ctx.Repo.Repository.Forks {
			if s[0] == fork.Owner.Name {
				afterRepo = fork
				break
			}
		}
		if afterRepo == nil {
			ctx.Handle(500, "GetRepositoryByRef", err)
			return
		}
		afterBranch = s[1]
	} else {
		afterRepo = ctx.Repo.Repository
		afterBranch = afterCommitId
	}

	afterRepoPath, err := afterRepo.RepoPath()
	if err != nil {
		ctx.Handle(500, "GetRepositoryByRef", err)
		return
	}
	afterRep, err := git.OpenRepository(afterRepoPath)
	if err != nil {
		ctx.Handle(500, "GetRepositoryByRef", err)
		return
	}
	afterBranches, _ := afterRep.GetBranches()

	ctx.Data["BeforeCommitId"] = beforeCommitId
	ctx.Data["AfterCommitId"] = afterCommitId
	ctx.Data["BeforeRepo"] = beforeRepo
	ctx.Data["AfterRepo"] = afterRepo
	ctx.Data["BeforeBranches"] = beforeBranches
	ctx.Data["BeforeBranch"] = beforeBranch
	ctx.Data["BeforeRepoPath"] = beforeRepoPath
	ctx.Data["RepoLink"] = repoLink
	ctx.Data["AfterBranches"] = afterBranches
	ctx.Data["AfterBranch"] = afterBranch
	ctx.Data["AfterRepoPath"] = afterRepo.Owner.Name + "/" + afterRepo.Name
	ctx.Data["Username"] = userName
	ctx.Data["Reponame"] = repoName

	ctx.Data["Forks"] = append([]*models.Repository{ctx.Repo.Repository}, ctx.Repo.Repository.Forks...)
	ctx.Data["Title"] = "Comparing " + base.ShortSha(beforeCommitId) + "..." + base.ShortSha(afterCommitId) + " · " + userName + "/" + repoName

	repo, err := git.OpenRepository(afterRepoPath)
	if err != nil {
		ctx.Handle(404, "OpenRepository", err)
		return
	}

	err = models.CheckUpstream(beforeRepoPath, afterRepoPath, beforeBranch)
	if err != nil {
		ctx.Handle(404, "CheckUpstream", err)
		return
	}

	diff, err := models.GetDiffForkedRange(beforeRepoPath, afterRepoPath,
		beforeBranch, afterBranch, setting.Git.MaxGitDiffLines)
	if err != nil {
		ctx.Handle(404, "GetDiffForkedRange", err)
		return
	}

	afterCommit, err := repo.GetCommitIdOfRef("refs/remotes/upstream/" + afterBranch)
	if err != nil {
		ctx.Handle(404, "GetCommitIdOfRef", err)
		return
	}
	afterCommitId = afterCommit

	commit, err := repo.GetCommit(afterCommitId)
	if err != nil {
		ctx.Handle(404, "GetCommit", err)
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

	commits, err := repo.CommitsBetweenBranch("upstream/"+beforeBranch, afterBranch, 1)
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
	ctx.Data["DiffNotAvailable"] = diff.NumFiles() == 0
	ctx.Data["SourcePath"] = setting.AppSubUrl + "/" + path.Join(afterRepo.Owner.Name, afterRepo.Name, "src", afterBranch)
	ctx.Data["BeforeSourcePath"] = setting.AppSubUrl + "/" + path.Join(beforeRepo.Owner.Name, beforeRepo.Name, "src", beforeBranch)
	ctx.Data["RawPath"] = setting.AppSubUrl + "/" + path.Join(afterRepo.Owner.Name, afterRepo.Name, "raw", afterBranch)

	ctx.HTML(200, PULL_COMPARE)
}
