package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/frizinak/gonzalo/git"
	"github.com/frizinak/gonzalo/ssh/sshconn"
	"github.com/frizinak/gonzalo/ssh/sshmanager"
	"golang.org/x/crypto/ssh"
)

type Gonzalo struct {
	sshkey    ssh.Signer
	ssh       *sshmanager.Pool
	sshHStore sshmanager.KeyStorage
	sshPStore sshmanager.KeyStorage

	git *git.Pool
}

func NewGonzalo(
	storage string,
	sshkey ssh.Signer,
	gitAuth map[string]git.Auth,
) (*Gonzalo, error) {
	storages := [2]string{
		filepath.Join(storage, "ssh", "known_hosts"),
		filepath.Join(storage, "ssh", "private"),
	}

	for _, p := range storages {
		os.MkdirAll(p, 0700)
	}

	fshost, err := sshmanager.NewFSKeyStorage(storages[0], 0644)
	if err != nil {
		return nil, err
	}
	fspriv, err := sshmanager.NewFSKeyStorage(storages[1], 0600)
	if err != nil {
		return nil, err
	}

	gitpool := git.NewPool(filepath.Join(storage, "git"))
	for provider := range gitAuth {
		gitpool.SetProviderAuth(provider, gitAuth[provider])
	}

	gonzalo := &Gonzalo{
		sshkey,
		sshmanager.NewPool(fshost, fspriv, 2048),
		fshost,
		fspriv,
		gitpool,
	}

	return gonzalo, nil
}

func (g *Gonzalo) sshClient(
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

func (g *Gonzalo) repo(provider, vendor, project string) (*git.Repo, error) {
	return g.git.Add(provider, vendor, project)
}

func main() {
	gitkey := sshconn.MustPKey(sshconn.ParsePrivateKeyFile("resources/git.key"))
	sshkey := sshconn.MustPKey(sshconn.ParsePrivateKeyFile("resources/key"))
	gonzalo, err := NewGonzalo(
		"storage",
		sshkey,
		map[string]git.Auth{
			"wieni.githost.io": git.NewSSHAuth(gitkey, ""),
			"github.com":       git.NewNoAuth(),
		},
	)
	if err != nil {
		panic(err)
	}

	pubrepo, err := gonzalo.repo("github.com", "frizinak", "ym")
	if err != nil {
		panic(err)
	}
	if err := pubrepo.Open(); err != nil {
		panic(err)
	}

	privaterepo, err := gonzalo.repo("wieni.githost.io", "wieni", "sbstv")
	if err != nil {
		panic(err)
	}
	if err := privaterepo.Open(); err != nil {
		log.Println("Failed to open private repo")
	}

	c, err := gonzalo.sshClient("dako.friz.pro", "22", "asdf")
	if err != nil {
		panic(err)
	}

	cmds := []string{
		"whoami",
		"hostname",
		"ls",
	}

	for _, cmd := range cmds {
		stdout, stderr, err := c.Output(cmd, nil)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(string(stdout), string(stderr), err)
	}
}
