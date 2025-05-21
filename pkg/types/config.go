package types

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".gws.yaml"

type Config struct {
	Contexts           map[string]*Context  `yaml:"contexts"`
	CurrentContextName string               `yaml:"currentContext"`
	FilePath           string               `yaml:"-"`
	FilePatches        map[string]FilePatch `yaml:"filePatches,omitempty"`
	currentContext     *Context
	Token              oauth2.Token `yaml:"token"`
}

func (c *Config) CurrentContext() *Context {
	return c.currentContext
}

func (c *Config) Load(fileName string) error {
	var file string
	file, data, err := ReadGWSFile(fileName)
	if err != nil {
		return err
	}
	_, _ = fmt.Printf("Using config: %s\n", file)

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

func ReadGWSFile(fileName string) (absoluteFile string, data []byte, err error) {
	var file string
	if fileName != "" {
		if _, err := os.Stat(fileName); err == nil {
			file = fileName
		}
	}

	if file == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return "", nil, err
		}

		homePath := filepath.Join(userHomeDir, ConfigFileName)
		if _, err := os.Stat(homePath); err != nil {
			return "", nil, errors.New("config file not found")
		}
		file = homePath
	}

	abs, err := filepath.Abs(file)
	if err != nil {
		return "", nil, err
	}
	data, err = os.ReadFile(file)

	return abs, data, err
}

func (c *Config) SwitchContext(newContext string) error {
	if _, ok := c.Contexts[newContext]; !ok {
		return fmt.Errorf("context with name %q not defined", newContext)
	}

	if c.CurrentContextName != newContext {
		c.CurrentContextName = newContext

		if err := c.save(); err != nil {
			return err
		}
	}
	c.currentContext = c.Contexts[newContext]

	return nil
}

func (c *Config) save() error {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	err := encoder.Encode(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.FilePath, buf.Bytes(), 0o600)
}

func (c *Config) SetToken(token oauth2.Token) error {
	if c.Token.AccessToken != token.AccessToken {
		_, _ = fmt.Printf("🎟️ Got new Google Access Token (expires: %s)\n", token.Expiry.Format(time.RFC822))
		c.Token = token
		return c.save()
	}
	return nil
}
