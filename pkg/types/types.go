package types

import (
	"fmt"
	"maps"
)

type Config struct {
	Contexts map[string]Context
}

func (c Config) Get(name string) (string, *Context, error) {
	if len(c.Contexts) == 1 {
		for k := range maps.Keys(c.Contexts) {
			ctx := c.Contexts[k]
			return k, &ctx, nil
		}
	}
	if name == "" {
		return "", nil, fmt.Errorf("context name must be defined")
	}
	if ctx, ok := c.Contexts[name]; ok {
		return name, &ctx, nil
	}

	return "", nil, fmt.Errorf("context with name %q not defined", name)
}

type Context struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	PrivateKeyFile string `yaml:"privateKeyFile"`

	GCloud *GCloud `yaml:"gcloud"`

	Dirs  []Dir  `yaml:"dirs"`
	Files []File `yaml:"files"`
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
