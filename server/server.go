package server

import (
	"log"
	"net"
	"os"

	"github.com/frizinak/gonzalo/git"
	"github.com/frizinak/gonzalo/project"
	"github.com/frizinak/gonzalo/ssh/sshconn"
	"github.com/frizinak/gonzalo/ssh/sshmanager"
	"github.com/frizinak/gonzalo/stores"
	"golang.org/x/crypto/ssh"
)

const (
	DeployFile = ".deploy"
)

type Gonzalo struct {
	sshkey ssh.Signer

	ssh *sshmanager.Pool
	git *git.Pool
}

func New(
	sshkey ssh.Signer,
	gitAuth map[string]git.Auth,
	hostKeyStore stores.KeyStorage,
	privateKeyStore stores.KeyStorage,
	gitdir string,
) (*Gonzalo, error) {
	gitpool := git.NewPool(gitdir)
	for provider := range gitAuth {
		gitpool.SetProviderAuth(provider, gitAuth[provider])
	}

	gonzalo := &Gonzalo{
		sshkey,
		sshmanager.NewPool(hostKeyStore, privateKeyStore, 2048),
		gitpool,
	}

	return gonzalo, nil
}

func (g *Gonzalo) SSHClient(
	host, port, user string,
) (*sshconn.Connection, error) {
	logger := log.New(os.Stdout, "ssh-"+host, log.LstdFlags)
	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}

	m, err := g.ssh.Add(logger, g.sshkey, addr, user, true)
	if err != nil {
		return nil, err
	}

	return m.Connection(), nil
}

func (g *Gonzalo) Repo(provider, vendor, proj string) (*git.Repo, error) {
	return g.git.Add(provider, vendor, proj)
}

func (g *Gonzalo) Project(provider, vendor, proj string) (
	*project.Project,
	error,
) {
	repo, err := g.Repo(provider, vendor, proj)
	if err != nil {
		return nil, err
	}

	return project.New(repo, DeployFile, g.ssh), nil
}
