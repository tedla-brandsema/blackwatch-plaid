package repo

import (
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"log/slog"
	"os"
	"sync"
)

var (
	mux = sync.RWMutex{}
	fs  billy.Filesystem
)

func openRepo(path string) (*git.Repository, error) {
	mux.Lock()
	defer mux.Unlock()

	fs = osfs.New(path)
	storage := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
	r, err := git.Open(storage, fs)
	if err != nil && errors.Is(err, git.ErrRepositoryNotExists) {
		slog.Info("repository does not exit: initializing repository")
		r, err = git.Init(storage, fs)
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

func deleteRepo(path string) error {
	mux.Lock()
	defer mux.Unlock()

	return os.RemoveAll(path)
}

func getWorktree(repo *git.Repository) (*git.Worktree, error) {
	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("worktree error %w", err)
	}
	return w, nil
}
