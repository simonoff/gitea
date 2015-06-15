package models

import (
	"fmt"
	"testing"
)

func TestGetDiffForkedRange(t *testing.T) {
	repoPath := "/Users/lunny/gogs-repos/lunny/sss.git"
	forkedRepoPath := "/Users/lunny/gogs-repos/ttes/sss.git"
	diffs, err := GetDiffForkedRange(repoPath, forkedRepoPath,
		"develop", "master", 10000)
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Println("total addition", diffs.TotalAddition)
	fmt.Println("total deletion", diffs.TotalDeletion)
	for _, file := range diffs.Files {
		fmt.Println(file.Name)
		fmt.Println("-------------")
		for _, section := range file.Sections {
			for _, line := range section.Lines {
				fmt.Println(line.LeftIdx, line.RightIdx, line.Content)
			}
			fmt.Println()
		}
	}
}

func TestForkedMerge(t *testing.T) {
	repoPath := "/Users/lunny/gogs-repos/lunny/sss.git"
	forkedRepoPath := "/Users/lunny/gogs-repos/ttes/sss.git"
	beforeBranch, afterBranch := "develop", "master"
	before, after, err := ForkedMerge(repoPath, forkedRepoPath, beforeBranch, afterBranch)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Printf("%s..%s %s -> %s", before, after, afterBranch, beforeBranch)
}
