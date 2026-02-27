package types

import (
	"fmt"
)

type Context struct {
	Host           string `validate:"required"       yaml:"host"`
	Port           int    `validate:"required"       yaml:"port"`
	User           string `                          yaml:"user"`
	PrivateKeyFile string `validate:"omitempty,file" yaml:"privateKeyFile"`
	KnownHostsFile string `validate:"omitempty,file" yaml:"knownHostsFile"`

	GCloud *GCloud `yaml:"gcloud"`

	Dirs  []Dir  `yaml:"dirs,omitempty"`
	Files []File `yaml:"files,omitempty" validate:"dive,required"`
}

type GCloud struct {
	Project string `yaml:"project"`
	Region  string `yaml:"region"`
	Cluster string `yaml:"cluster"`
	Config  string `yaml:"config"`
	Name    string `yaml:"name"`
}

func (c Context) HostAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type Dir struct {
	Path        string `yaml:"path"`
	Permissions string `yaml:"permissions,omitempty"`
}

type File struct {
	SourcePath  string `validate:"required,file"          yaml:"sourcePath"`
	Path        string `validate:"required"              yaml:"path"`
	Permissions string `                                  yaml:"permissions"`
	Direction   string `validate:"omitempty,oneof=up down" yaml:"direction"`
}

type FilePatch struct {
	File     string `yaml:"file"`
	Indent   string `yaml:"indent,omitempty"`
	OldBlock string `yaml:"oldBlock,omitempty"`
	NewBlock string `yaml:"newBlock,omitempty"`
}
