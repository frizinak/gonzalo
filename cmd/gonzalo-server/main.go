package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/frizinak/gonzalo/git"
	"github.com/frizinak/gonzalo/server"
	"github.com/frizinak/gonzalo/ssh/sshconn"
	"github.com/frizinak/gonzalo/stores"
)

func main() {
	gitkey := sshconn.MustPKey(sshconn.ParsePrivateKeyFile("resources/git.key"))
	sshkey := sshconn.MustPKey(sshconn.ParsePrivateKeyFile("resources/key"))

	storage := "storage"
	storages := [2]string{
		filepath.Join(storage, "ssh", "known_hosts"),
		filepath.Join(storage, "ssh", "private"),
	}

	for _, p := range storages {
		os.MkdirAll(p, 0700)
	}

	hostKeyStore, err := stores.NewFSKeyStorage(storages[0], 0644)
	if err != nil {
		panic(err)
	}

	privateKeyStore, err := stores.NewFSKeyStorage(storages[1], 0600)
	if err != nil {
		panic(err)
	}

	gonzalo, err := server.New(
		sshkey,
		map[string]git.Auth{
			"wieni.githost.io": git.NewSSHAuth(gitkey, hostKeyStore, ""),
			"github.com":       git.NewNoAuth(),
		},
		hostKeyStore,
		privateKeyStore,
		filepath.Join(storage, "git"),
	)
	if err != nil {
		panic(err)
	}

	pubrepo, err := gonzalo.Repo("github.com", "frizinak", "ym")
	if err != nil {
		panic(err)
	}

	if err := pubrepo.Open(); err != nil {
		panic(err)
	}

	if err := pubrepo.Update(); err != nil {
		panic(err)
	}

	prj, err := gonzalo.Project("wieni.githost.io", "wieni", "sbstv")
	if err != nil {
		panic(err)
	}

	conf, err := prj.ConfigEnv("9.0.0", "dev-backend")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", conf)

	// privaterepo, err := gonzalo.Repo("wieni.githost.io", "wieni", "sbstv")
	// if err != nil {
	// 	panic(err)
	// }
	// if err := privaterepo.Open(); err != nil {
	// 	log.Println("Failed to open private repo")
	// }

	// if err := privaterepo.Update(); err != nil {
	// 	log.Println("Failed to update private repo")
	// }

	c, err := gonzalo.SSHClient("dako.friz.pro", "22", "asdf")
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
