package easygit

import (
	"os"
	"sync"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/coronon/artemisbot/internal/git"
)

func NewRepository(config *git.GitConfig, credentials *git.GitCredentials, path string) (git.Repository, error) {
	auth := &http.BasicAuth{
		Username: credentials.Username,
		Password: credentials.Password,
	}

	repo, err := gogit.PlainClone(path, false, &gogit.CloneOptions{
		URL:           config.URL,
		Auth:          auth,
		ReferenceName: plumbing.ReferenceName(config.Branch),
		SingleBranch:  true,
		Progress:      nil,
	})
	if err != nil {
		return nil, err
	}

	return &GoGitRepository{
		mux:    sync.Mutex{},
		path:   path,
		repo:   repo,
		auth:   auth,
		config: config,

		isClosed: false,
	}, nil
}

type GoGitRepository struct {
	mux    sync.Mutex
	path   string
	repo   *gogit.Repository
	auth   transport.AuthMethod
	config *git.GitConfig

	isClosed bool
}

func (r *GoGitRepository) Close() error {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.isClosed = true

	return os.RemoveAll(r.path)
}

func (r *GoGitRepository) PushEmptyCommit() (string, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	wt, err := r.repo.Worktree()
	if err != nil {
		return "", err
	}

	// Build commit
	sig := &object.Signature{
		Name:  r.config.Name,
		Email: r.config.Email,
		When:  time.Now(),
	}

	hash, err := wt.Commit("retrigger", &gogit.CommitOptions{
		AllowEmptyCommits: true,
		Author:            sig,
		Committer:         sig,
	})
	if err != nil {
		return "", err
	}

	// Push commit
	err = r.repo.Push(&gogit.PushOptions{
		Auth:     r.auth,
		Progress: nil,
		Atomic:   true,
	})
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}
