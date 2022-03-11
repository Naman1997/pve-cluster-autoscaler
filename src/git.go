package main

import (
	"os"

	git "github.com/go-git/go-git/v5"
)

func CloneRepo(ansibleRepo string) error {
	ColorPrint(INFO, "Attempting to clone the provided repo for ansible.")
	repo, err := git.PlainClone(REPO_LOCATION, false, &git.CloneOptions{
		URL:      ansibleRepo,
		Progress: os.Stdout,
	})
	if err != nil && err == git.ErrRepositoryAlreadyExists {
		ColorPrint(INFO, "Repo already cloned. Attempting to pull the latest from origin.")
		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}
		err = worktree.Pull(&git.PullOptions{})
		return err
	}
	return err
}
