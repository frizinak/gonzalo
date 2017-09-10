package project

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Command string
type Role uint

type Env struct {
	// Share deploy artifacts by overriding the env with this value.
	BuildKey string `yaml:"buildkey"`

	// Amount of deployment backups to keep.
	Backups int `yaml:"backups"`

	// Deprecated
	Server string `yaml:"server"`

	// The host to deploy to.
	Host string `yaml:"host"`
	// The user on the remote server.
	User string `yaml:"user"`

	// Path inside your repo that will be deployed.
	Root string `yaml:"root"`
	// Path your repo will be deployed to.
	Dest string `yaml:"dest"`

	// The minimum role that is allowed to deploy.
	Role Role `yaml:"role"`

	// The chat channel that will receive deployment pings.
	Chatroom string `yaml:"chatroom"`

	// List of paths inside the repo that are uploaded before deployment starts.
	Required []string `yaml:"required"`

	// List of commands whose output is backed up using the key as filename.
	Backup map[string]Command `yaml:"backup"`

	Build             []Command `yaml:"build"`
	PreUpload         []Command `yaml:"pre-upload"`
	DuringUpload      []Command `yaml:"during-upload"`
	PostUploadCurrent []Command `yaml:"post-upload-current"`
	PostUploadNext    []Command `yaml:"post-upload-next"`
	PostDeploy        []Command `yaml:"post-deploy"`
}

type Config map[string]Env

func (c Config) GetEnv(name string) (Env, error) {
	// TODO https://github.com/imdario/mergo
	env, ok := c[name]
	//all := c["all"]
	if !ok {
		return env, fmt.Errorf("Env %s is not defined", name)
	}

	return env, nil
}

func decodeFile(f string) (*Config, error) {
	d, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	return c, yaml.Unmarshal(d, c)
}
