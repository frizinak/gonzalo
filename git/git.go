package git

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const remote = "origin"

type Repo struct {
	auth     Auth
	provider string
	vendor   string
	project  string
	path     string
	repo     *git.Repository
}

func New(
	dir,
	provider, vendor, project string,
	auth Auth,
) (*Repo, error) {
	path := filepath.Join(provider, vendor, project)
	if strings.Count(path, string(filepath.Separator)) != 2 {
		return nil, fmt.Errorf(
			"provider, vendor or project are invalid: the resulting path would be %s",
			path,
		)
	}

	return &Repo{
		auth:     auth,
		provider: provider,
		vendor:   vendor,
		project:  project,
		path:     filepath.Join(dir, path),
	}, nil
}

// Open opens the repo if it exists, clones it otherwise.
func (r *Repo) Open() error {
	if r.repo != nil {
		return nil
	}

	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return r.Update()
	}

	r.repo = repo
	return nil
}

// Ensure opens the repo if it exists, clones it if not and runs git fetch.
func (r *Repo) Update() (err error) {
	defer func() {
		if err != nil {
			r.Delete()
		}
	}()

	clone := func() {
		if err = r.Delete(); err != nil {
			return
		}

		err = r.clone()
	}

	fetch := func() {
		err = r.repo.Fetch(
			&git.FetchOptions{
				RemoteName: remote,
				Auth:       r.getAuth(),
			},
		)

		if err == git.NoErrAlreadyUpToDate {
			err = nil
		}

		if err != nil {
			clone()
		}
	}

	if r.repo != nil {
		fetch()
		return
	}

	var repo *git.Repository
	repo, err = git.PlainOpen(r.path)
	if err != nil {
		clone()
		return
	}

	r.repo = repo
	if fetch(); err != nil {
		r.repo = nil
	}

	return
}

// Reset resets the repo (hard) to the given commitish
func (r *Repo) Reset(commitish string) error {
	if err := r.Open(); err != nil {
		return err
	}

	head, err := r.repo.Head()
	if err != nil {
		return err
	}
	current := head.Hash()
	if current.String() == commitish {
		return reset(r.repo, current)
	}

	commits, err := lookup(r.repo, commitish)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		return fmt.Errorf("No such commitish: %s", commitish)
	}

	if len(commits) > 1 {
		refs := make([]string, len(commits))
		for i := range commits {
			refs[i] = commits[i].ref
		}

		return fmt.Errorf("Ambiguous commitish: %s", strings.Join(refs, ", "))
	}

	return reset(r.repo, commits[0].commit.Hash)
}

func (r *Repo) Delete() error {
	return os.RemoveAll(r.path)
}

func (r *Repo) clone() error {
	repo, err := git.PlainClone(
		r.path,
		false,
		&git.CloneOptions{
			URL:               r.uri(),
			Auth:              r.getAuth(),
			RemoteName:        remote,
			SingleBranch:      false,
			Depth:             0,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Progress:          nil,
			Tags:              git.AllTags,
		},
	)

	if err != nil {
		return err
	}

	r.repo = repo

	return nil
}

func (r *Repo) uri() string {
	if r.auth.proto == protoGit {
		return fmt.Sprintf("%s:%s/%s", r.provider, r.vendor, r.project)
	}

	var prefix string
	if r.auth.user != "" && r.auth.password != "" {
		prefix = fmt.Sprintf(
			"%s:%s@",
			url.QueryEscape(r.auth.user),
			url.QueryEscape(r.auth.password),
		)
	} else if r.auth.user != "" {
		prefix = fmt.Sprintf("%s@", url.QueryEscape(r.auth.user))
	}

	return fmt.Sprintf(
		"https://%s%s/%s/%s.git",
		prefix,
		r.provider,
		r.vendor,
		r.project,
	)
}

func (r *Repo) getAuth() transport.AuthMethod {
	if r.auth.proto == protoGit {
		return r.auth.key
	}

	if r.auth.user == "" && r.auth.password == "" {
		return nil
	}

	return http.NewBasicAuth(r.auth.user, r.auth.password)
}

func reset(repo *git.Repository, commit plumbing.Hash) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = wt.Checkout(
		&git.CheckoutOptions{Hash: commit, Force: true},
	)

	if err != nil {
		return err
	}

	return wt.Reset(
		&git.ResetOptions{commit, git.HardReset},
	)
}
