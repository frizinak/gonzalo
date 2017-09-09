package git

import (
	"fmt"
	"strings"
	"sync"
)

type Pool struct {
	pool         map[string]*Repo
	providerAuth map[string]*Auth
	m            sync.RWMutex
	dir          string
}

func NewPool(dir string) *Pool {
	return &Pool{
		pool:         map[string]*Repo{},
		providerAuth: map[string]*Auth{},
		dir:          dir,
	}
}

func (p *Pool) SetProviderAuth(provider string, auth Auth) {
	p.m.Lock()
	p.providerAuth[provider] = &auth
	p.m.Unlock()
}

func (p *Pool) Get(provider, vendor, project string) *Repo {
	p.m.RLock()
	r := p.pool[key(provider, vendor, project)]
	p.m.RUnlock()
	return r
}

func (p *Pool) Add(provider, vendor, project string) (*Repo, error) {
	if r := p.Get(provider, vendor, project); r != nil {
		return r, nil
	}

	p.m.RLock()
	auth := p.providerAuth[provider]
	p.m.RUnlock()
	if auth == nil {
		return nil, fmt.Errorf("No auth found for %s", provider)
	}

	return p.AddCustomAuth(provider, vendor, project, *auth)
}

func (p *Pool) AddCustomAuth(
	provider, vendor, project string,
	auth Auth,
) (*Repo, error) {
	if r := p.Get(provider, vendor, project); r != nil {
		return r, nil
	}

	p.m.Lock()
	defer p.m.Unlock()
	r, err := New(p.dir, provider, vendor, project, auth)
	if err != nil {
		return nil, err
	}

	p.pool[key(provider, vendor, project)] = r
	return r, nil
}

func key(provider, vendor, project string) string {
	return strings.Join([]string{provider, vendor, project}, ":")
}
