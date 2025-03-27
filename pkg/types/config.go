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
	Contexts           map[string]*Context  `yaml:"contexts"`
	CurrentContextName string               `yaml:"currentContext"`
	FilePath           string               `yaml:"-"`
	FilePatches        map[string]FilePatch `yaml:"filePatches,omitempty"`
	currentContext     *Context
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
		if _, err := os.Stat(homePath); err != nil {
			return errors.New("config file not found")
		}
		file = homePath
	}

	p, _ := filepath.Abs(file)
	_, _ = fmt.Printf("Reading config: %s\n", p)

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
				c.CurrentContextName = k
			}
		}
	}
	return c.SwitchContext(c.CurrentContextName)
}

func (c *Config) SwitchContext(newContext string) error {
	if _, ok := c.Contexts[newContext]; !ok {
		return fmt.Errorf("context with name %q not defined", newContext)
	}

	if c.CurrentContextName != newContext {
		c.CurrentContextName = newContext

		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)

		err := encoder.Encode(c)
		if err != nil {
			return err
		}
		if err := os.WriteFile(c.FilePath, buf.Bytes(), 0o600); err != nil {
			return err
		}
	}
	c.currentContext = c.Contexts[newContext]

	return nil
}
