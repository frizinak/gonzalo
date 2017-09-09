package sshmanager

import (
	"net"
	"sync"

	"github.com/frizinak/gonzalo/ssh/sshconn"
	"golang.org/x/crypto/ssh"
)

type Pool struct {
	pool   map[string]*Manager
	m      sync.RWMutex
	hstore KeyStorage
	pstore KeyStorage
	bits   int
}

func NewPool(hostKeyStorage, privateKeyStorage KeyStorage, keyBits int) *Pool {
	return &Pool{
		pool:   map[string]*Manager{},
		hstore: hostKeyStorage,
		pstore: privateKeyStorage,
		bits:   keyBits,
	}
}

func (p *Pool) Get(addr net.Addr, user string) *Manager {
	p.m.RLock()
	m := p.pool[key(addr, user)]
	p.m.RUnlock()
	return m
}

func (p *Pool) Add(
	log sshconn.Logger,
	pkey ssh.Signer,
	addr net.Addr,
	user string,
	replaceKey bool,
) (*Manager, error) {
	if m := p.Get(addr, user); m != nil {
		return m, nil
	}

	p.m.Lock()
	defer p.m.Unlock()
	m, err := New(log, pkey, addr, user, p.hstore, p.pstore)
	if err != nil {
		return nil, err
	}

	if replaceKey {
		if err := m.ReplaceKey(p.bits); err != nil {
			return nil, err
		}
	}

	p.pool[key(addr, user)] = m
	return m, nil
}

func key(a net.Addr, u string) string {
	return a.String() + ":" + u
}
