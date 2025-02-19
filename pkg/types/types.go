package types

import (
	"fmt"
)

type Context struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	PrivateKeyFile string `yaml:"privateKeyFile"`

	GCloud *GCloud `yaml:"gcloud"`

	Dirs  []Dir  `yaml:"dirs,omitempty"`
	Files []File `yaml:"files,omitempty"`
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
	Permissions string `yaml:"permissions"`
}

type File struct {
	SourcePath  string `yaml:"sourcePath"`
	Path        string `yaml:"path"`
	Permissions string `yaml:"permissions"`
}

type FilePatch struct {
	File     string   `yaml:"file"`
	Append   string   `yaml:"append,omitempty"`
	Indent   string   `yaml:"indent,omitempty"`
	OldBlock []string `yaml:"oldBlock,omitempty"`
	NewBlock []string `yaml:"newBlock,omitempty"`
}
