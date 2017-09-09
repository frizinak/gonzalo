package sshconn

import (
	"io/ioutil"
	"net"

	"golang.org/x/crypto/ssh"
)

func HostInfo(pkey ssh.Signer, host, user string) (
	ssh.PublicKey,
	net.Addr,
	error,
) {
	var hkey ssh.PublicKey
	var addr net.Addr

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(pkey)},
		HostKeyCallback: func(
			hostname string,
			remote net.Addr,
			key ssh.PublicKey,
		) error {
			hkey = key
			addr = remote
			return nil
		},
	}

	conn, err := ssh.Dial("tcp", host, config)
	if conn != nil {
		conn.Close()
	}

	return hkey, addr, err
}

func ParsePrivateKey(raw []byte) (ssh.Signer, error) {
	return ssh.ParsePrivateKey(raw)
}

func ParsePrivateKeyFile(file string) (ssh.Signer, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return ParsePrivateKey(raw)
}

func MustPKey(s ssh.Signer, err error) ssh.Signer {
	if err != nil {
		panic(err)
	}

	return s
}
