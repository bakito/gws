package types

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".gws.yaml"

type Config struct {
	Contexts           map[string]*Context `yaml:"contexts"`
	CurrentContextName string              `yaml:"currentContext"`
	currentContext     *Context            `yaml:"-"`
	FilePath           string              `yaml:"-"`
}

func (c *Config) CurrentContext() *Context {
	return c.currentContext
}

func (c *Config) Load(fileName string) error {
	var file string
	if fileName != "" {
		if _, err := os.Stat(fileName); err == nil {
			file = fileName
		}
	}

	if file == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		homePath := filepath.Join(userHomeDir, ConfigFileName)
		if _, err := os.Stat(homePath); err == nil {
			file = homePath
		} else {
			return errors.New("config file not found")
		}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}

	c.FilePath = file

	if c.CurrentContextName == "" {
		if len(c.Contexts) == 1 {
			for k := range maps.Keys(c.Contexts) {
				if err = c.SwitchContext(k); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Config) SwitchContext(newContext string) error {
	if _, ok := c.Contexts[newContext]; !ok {
		return fmt.Errorf("context with name %q not defined", newContext)
	}

	if c.CurrentContextName != newContext {
		c.CurrentContextName = newContext
		c.currentContext = c.Contexts[newContext]

		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)

		err := encoder.Encode(c)
		if err != nil {
			return err
		}
		if err = os.WriteFile(c.FilePath, buf.Bytes(), 0o600); err != nil {
			return err
		}
	}

	return nil
}

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
