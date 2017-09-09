package sshmanager

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/frizinak/gonzalo/ssh/sshconn"
	"golang.org/x/crypto/ssh"
)

// Manager manages an ssh connection.
type Manager struct {
	log      sshconn.Logger
	conn     *sshconn.Connection
	pstorage KeyStorage
	addr     net.Addr
}

// New returns an ssh connection manager with hostkey verification.
func New(
	log sshconn.Logger,
	pkey ssh.Signer,
	addr net.Addr,
	user string,
	hostKeyStorage KeyStorage,
	privateKeyStorage KeyStorage,
) (*Manager, error) {
	if privateKeyStorage.Has(addr, user) {
		raw := privateKeyStorage.Get(addr, user)
		if newPKey, err := ssh.ParsePrivateKey(raw); err == nil {
			pkey = newPKey
		}
	}

	getFresh := func() (ssh.PublicKey, error) {
		hkey, _, err := sshconn.HostInfo(pkey, addr.String(), user)
		if err != nil {
			return nil, err
		}
		return hkey, nil

	}

	hkey, err := checkHostKey(addr, hostKeyStorage, getFresh)
	if err != nil {
		return nil, err
	}

	return &Manager{
		log,
		sshconn.New(log, hkey, pkey, addr, user),
		privateKeyStorage,
		addr,
	}, nil
}

func (m *Manager) Addr() net.Addr {
	return m.addr
}

// Connection returns the underlying ssh connection.
func (m *Manager) Connection() *sshconn.Connection {
	return m.conn
}

func (m *Manager) Close() error {
	return m.conn.Close()
}

// ReplaceKey replaces the current publicKey used in the connection with a
// newly generated one and resets the connection.
func (m *Manager) ReplaceKey(bits int) error {
	current, replaced, err := m.pkey()
	if err != nil || replaced {
		return err
	}

	currentKey := base64.RawStdEncoding.EncodeToString(
		current.PublicKey().Marshal(),
	)

	rawPKey, err := GenerateRSA(bits)
	if err != nil {
		return err
	}

	pkey, err := ssh.ParsePrivateKey(rawPKey)
	if err != nil {
		panic(err)
	}

	newKey := base64.RawStdEncoding.EncodeToString(
		pkey.PublicKey().Marshal(),
	)

	rnd := time.Now().UnixNano()
	_, _, err = m.conn.Output(
		fmt.Sprintf(
			`tmp="$HOME/.ssh/authorized_keys.%d" && \
			bu="$HOME/.ssh/authorized_keys.gonzalo.backup" && \
			cp "$HOME/.ssh/authorized_keys" "$bu" && \
			cp "$HOME/.ssh/authorized_keys" "$tmp" && \
			echo 'ssh-rsa %s' >> "$tmp" && \
			line=$(cat -n "$tmp" | grep '%s' | cut -f1 | xargs) && \
			sed -i "${line}d" "$tmp" && \
			mv "$tmp" "$HOME/.ssh/authorized_keys"`,
			rnd,
			newKey,
			currentKey,
		),
		nil,
	)

	if err != nil {
		return err
	}

	if err := m.setPKey(rawPKey); err != nil {
		m.conn.Output(
			`cp "$HOME/.ssh/authorized_keys.gonzalo.backup" \
			"$HOME/.ssh/authorized_keys"`,
			nil,
		)

		return err
	}

	m.conn.SetPrivateKey(pkey)
	return nil
}

func (m *Manager) pkey() (ssh.Signer, bool, error) {
	user := m.conn.User()
	if m.pstorage.Has(m.addr, user) {
		raw := m.pstorage.Get(m.addr, user)
		pkey, err := ssh.ParsePrivateKey(raw)
		return pkey, true, err
	}

	return m.conn.PrivateKey(), false, nil
}

func (m *Manager) setPKey(pkey []byte) error {
	return m.pstorage.Set(m.addr, m.conn.User(), pkey)
}

func checkHostKey(
	addr net.Addr,
	storage KeyStorage,
	getFresh func() (ssh.PublicKey, error),
) (ssh.PublicKey, error) {
	user := "host"
	if storage.Has(addr, user) {
		raw := storage.Get(addr, user)
		if raw == nil {
			return nil, errors.New("Could not get contents of known hostkey")
		}

		stored, err := ssh.ParsePublicKey(raw)
		if err != nil {
			return nil, err
		}

		return stored, nil
	}

	fresh, err := getFresh()
	if err != nil {
		return nil, err
	}

	if fresh == nil {
		return nil, errors.New("Fresh hostkey cannot be nil")
	}

	return fresh, storage.Set(addr, user, fresh.Marshal())
}
