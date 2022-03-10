package main

import (
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
)

func CloneRepo(ansibleRepo string) string {
	ColorPrint(INFO, "Attempting to clone the provided repo for ansible.")
	repo, err := git.PlainClone(REPO_LOCATION, false, &git.CloneOptions{
		URL:      ansibleRepo,
		Progress: os.Stdout,
	})
	if err != nil && err == git.ErrRepositoryAlreadyExists {
		ColorPrint(INFO, "Repo already cloned. Attempting to pull the latest from origin.")
		worktree, err := repo.Worktree()
		FailError(err)
		err = worktree.Pull(&git.PullOptions{})
		FailError(err)
	}
	FailError(err)
	return strings.ReplaceAll(ansibleRepo[strings.LastIndex(ansibleRepo, "/")+1:], ".git", "")
}
