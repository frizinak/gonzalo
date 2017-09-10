package project

import (
	"path/filepath"

	"github.com/frizinak/gonzalo/git"
	"github.com/frizinak/gonzalo/ssh/sshmanager"
)

type Project struct {
	repo *git.Repo
	fn   string
	ssh  *sshmanager.Pool
}

func New(
	repo *git.Repo,
	config string,
	ssh *sshmanager.Pool,
) *Project {
	return &Project{repo, config, ssh}
}

func (p *Project) Config(commitish string) (*Config, error) {
	if err := p.repo.Update(); err != nil {
		return nil, err
	}

	if err := p.repo.Reset(commitish); err != nil {
		return nil, err
	}

	return decodeFile(filepath.Join(p.repo.Path(), p.fn))
}

func (p *Project) ConfigEnv(commitish, env string) (Env, error) {
	c, err := p.Config(commitish)
	if err != nil {
		return Env{}, err
	}

	return c.GetEnv(env)
}
