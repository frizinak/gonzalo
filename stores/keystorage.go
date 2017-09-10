package stores

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
)

var pathRE *regexp.Regexp

func init() {
	pathRE = regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)
}

type KeyStorage interface {
	Has(net.Addr, string) bool
	Get(net.Addr, string) []byte
	Set(net.Addr, string, []byte) error
	Del(net.Addr, string) error
}

type FSKeyStorage struct {
	dir string
	fm  os.FileMode
}

func NewFSKeyStorage(dir string, filemode os.FileMode) (*FSKeyStorage, error) {
	stat, err := os.Stat(dir)
	if err != nil || !stat.IsDir() {
		return nil, err
	}

	return &FSKeyStorage{dir, filemode}, nil
}

func (fs *FSKeyStorage) Has(a net.Addr, user string) bool {
	_, err := os.Stat(fs.path(a, user))
	return !os.IsNotExist(err)
}

func (fs *FSKeyStorage) Get(a net.Addr, user string) []byte {
	c, err := ioutil.ReadFile(fs.path(a, user))
	if err != nil {
		return nil
	}

	return c
}

func (fs *FSKeyStorage) Set(a net.Addr, user string, k []byte) error {
	return ioutil.WriteFile(fs.path(a, user), k, fs.fm)
}

func (fs *FSKeyStorage) Del(a net.Addr, user string) error {
	return os.Remove(fs.path(a, user))
}

func (fs *FSKeyStorage) path(a net.Addr, u string) string {
	return filepath.Join(fs.dir, HashAddr(a, u))
}

func HashAddr(a net.Addr, u string) string {
	key := a.String() + "-" + u
	human := pathRE.ReplaceAllString(key, "-")
	sum := sha1.Sum([]byte(key))

	return human + "." + hex.EncodeToString(sum[:])
}
