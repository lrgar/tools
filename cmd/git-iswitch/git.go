package main

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type Repo struct {
	storage *filesystem.Storage
	rawRepo *git.Repository
}

func OpenRepo() (*Repo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	fs := osfs.New(cwd)
	if _, err := fs.Stat(git.GitDirName); err != nil {
		return nil, errors.New("the current directory is not a Git repository")
	}

	fs, err = fs.Chroot(git.GitDirName)
	if err != nil {
		return nil, fmt.Errorf("failed to chroot to .git directory: %w", err)
	}

	var r Repo
	r.storage = filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
	r.rawRepo, err = git.Open(r.storage, fs)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &r, nil
}

func (r *Repo) Close() error {
	return r.storage.Close()
}

func (r *Repo) Branches() ([]*plumbing.Reference, error) {
	detached := true

	headRef, err := r.rawRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get current HEAD: %w", err)
	}

	branchRefs, err := r.rawRepo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}
	defer branchRefs.Close()

	branches := []*plumbing.Reference{}
	if err = branchRefs.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, ref)
		if ref.Hash().String() == headRef.Hash().String() {
			detached = false
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to iterate through branches: %w", err)
	}

	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name().Short() < branches[j].Name().Short()
	})

	if !detached {
		return branches, nil
	}

	return append([]*plumbing.Reference{headRef}, branches...), nil
}

func (r *Repo) SwitchBranch(branch *plumbing.Reference) error {
	headRef, err := r.rawRepo.Head()
	if err != nil {
		return fmt.Errorf("failed to get current HEAD: %w", err)
	}

	if headRef.Hash().String() == branch.Hash().String() {
		return nil
	}

	wt, err := r.rawRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: branch.Name(),
		Keep:   true,
	}); err != nil {
		return fmt.Errorf("failed to get checkout in worktree: %w", err)
	}

	return nil
}
