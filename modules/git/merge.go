package git

import (
	"errors"
	"strings"

	"github.com/Unknwon/com"
)


// check if merge is fast-forward
func (repo *Repository) MergeCheck(fromBranch, toBranch string) (bool, error) {
	stdout, stderr, err := com.ExecCmdDir(repo.Path, "git", "merge-base", fromBranch, toBranch)
	if err != nil {
		return false, errors.New(stderr)
	}

	stdout2, stderr2, err := com.ExecCmdDir(repo.Path, "git", "merge-tree", strings.TrimSpace(stdout), toBranch, fromBranch)
	if err != nil {
		return false, errors.New(stderr2)
	}

	if strings.Contains(stdout2, "===") {
		return false, nil
	}
	return true, nil
}
