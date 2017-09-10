package git

import (
	"bytes"
	"errors"
	"net"

	"github.com/frizinak/gonzalo/stores"
	"golang.org/x/crypto/ssh"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	protoGit proto = iota
	protoHttps
)

type proto int

type Auth struct {
	proto    proto
	user     string
	password string
	key      *gitssh.PublicKeys
}

func NewSSHAuth(
	privateKey ssh.Signer,
	hostKeyStorage stores.KeyStorage,
	user string,
) Auth {
	if user == "" {
		user = "git"
	}

	key := &gitssh.PublicKeys{
		User:   user,
		Signer: privateKey,
	}

	key.HostKeyCallback = func(
		hostname string,
		remote net.Addr,
		key ssh.PublicKey,
	) error {
		if hostKeyStorage.Has(remote, user) {
			raw := hostKeyStorage.Get(remote, user)
			if len(raw) == 0 {
				return errors.New("Could not get contents of known hostkey")
			}

			if !bytes.Equal(key.Marshal(), raw) {
				return errors.New("Known hostkey does not match received hostkey")
			}

			return nil
		}

		return hostKeyStorage.Set(remote, user, key.Marshal())
	}

	return Auth{
		proto: protoGit,
		user:  user,
		key:   key,
	}
}

func NewNoAuth() Auth {
	return Auth{proto: protoHttps}
}

func NewHTTPSAuth(user, password string) Auth {
	return Auth{proto: protoHttps, user: user, password: password}
}
