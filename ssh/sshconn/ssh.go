package sshconn

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Logger interface {
	Println(v ...interface{})
}

type Connection struct {
	log  Logger
	hkey ssh.PublicKey
	pkey ssh.Signer
	addr net.Addr
	user string
	c    *ssh.Client
	mu   sync.Mutex
}

func New(
	log Logger,
	hkey ssh.PublicKey,
	pkey ssh.Signer,
	addr net.Addr,
	user string,
) *Connection {
	return &Connection{
		log:  log,
		hkey: hkey,
		pkey: pkey,
		addr: addr,
		user: user,
	}
}

func (c *Connection) Connect() error {
	if c.hkey == nil {
		return errors.New("Hostkey cannot be nil")
	}

	c.Close()
	config := &ssh.ClientConfig{
		User:            c.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(c.pkey)},
		HostKeyCallback: ssh.FixedHostKey(c.hkey),
	}

	conn, err := ssh.Dial(c.addr.Network(), c.addr.String(), config)
	if err != nil {
		return err
	}

	c.c = conn
	return nil
}

func (c *Connection) Close() (err error) {
	if c.c != nil {
		err = c.c.Close()
		c.c = nil
	}
	return
}

func (c *Connection) client() (*ssh.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.c == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	return c.c, nil
}

func (c *Connection) Session() (*ssh.Session, error) {
	client, err := c.client()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	var sess *ssh.Session
	defer func() {
		if sess != nil {
			sess.Wait()
		}
		c.mu.Unlock()
	}()

	sess, err = client.NewSession()
	if err != nil {
		return nil, err
	}

	return sess, err
}

func (c *Connection) Output(cmd string, stdin io.Reader) (
	stdout,
	stderr []byte,
	err error,
) {
	var stdoutB bytes.Buffer
	var stderrB bytes.Buffer
	var session *ssh.Session

	if session, err = c.Session(); err != nil {
		return
	}

	defer session.Close()

	session.Stdout = &stdoutB
	session.Stderr = &stderrB
	session.Stdin = stdin

	err = session.Run(cmd)
	stdout = stdoutB.Bytes()
	stderr = stderrB.Bytes()

	return
}

func (c *Connection) SetPrivateKey(pkey ssh.Signer) {
	c.pkey = pkey
	c.Close()
}

func (c *Connection) PrivateKey() ssh.Signer {
	return c.pkey
}

func (c *Connection) Addr() net.Addr {
	return c.addr
}

func (c *Connection) User() string {
	return c.user
}
