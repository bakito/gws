package types

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	"github.com/bakito/gws/internal/log"
)

const (
	ConfigFileName = "config.yaml"
	ConfigDir      = ".config/gws"
)

type Config struct {
	Contexts           map[string]*Context  `yaml:"contexts"`
	CurrentContextName string               `yaml:"currentContext"`
	FilePath           string               `yaml:"-"`
	TokenCheck         bool                 `yaml:"-"`
	FilePatches        map[string]FilePatch `yaml:"filePatches,omitempty"`
	SSHTimeoutSeconds  int                  `yaml:"sshTimeoutSeconds,omitempty"`
	currentContext     *Context
	Token              *TokenStorage `yaml:"-"`
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
	fmt.Printf("Using config: %s\n", file)

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

	tk, err := LoadToken()
	if err != nil {
		return err
	}
	c.Token = &TokenStorage{}
	if tk != nil {
		c.Token.Token = *tk
	}
	return c.SwitchContext(c.CurrentContextName, false)
}

func (c *Config) SSHTimeout() time.Duration {
	return time.Duration(c.SSHTimeoutSeconds) * time.Second
}

func ReadGWSFile(fileName string) (absoluteFile string, data []byte, err error) {
	var file string
	if fileName != "" {
		if _, err := os.Stat(fileName); err == nil {
			file = fileName
		}
	}

	if file == "" {
		// Try new location first
		newConfigPath, userHomeDir := DefaultConfigDir()
		if _, err := os.Stat(newConfigPath); err == nil {
			file = newConfigPath
		} else {
			// Fallback to legacy location for backward compatibility
			legacyPath := filepath.Join(userHomeDir, ".gws.yaml")
			if _, err := os.Stat(legacyPath); err != nil {
				return "", nil, fmt.Errorf("%w: config file not found", os.ErrNotExist)
			}
			file = legacyPath
			fmt.Printf("⚠️  Using legacy config location. Consider moving to: %s\n", newConfigPath)
		}
	}

	abs, err := filepath.Abs(file)
	if err != nil {
		return "", nil, err
	}
	data, err = os.ReadFile(file)

	return abs, data, err
}

func DefaultConfigDir() (newConfigPath, userHomeDir string) {
	userHomeDir, _ = os.UserHomeDir()

	newConfigPath = filepath.Join(userHomeDir, ConfigDir, ConfigFileName)
	return newConfigPath, userHomeDir
}

func (c *Config) SwitchContext(newContext string, force bool) error {
	if _, ok := c.Contexts[newContext]; !ok {
		return fmt.Errorf("context with name %q not defined", newContext)
	}

	if force || c.CurrentContextName != newContext {
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
	if c.Token == nil {
		c.Token = &TokenStorage{Token: token}
	}

	if c.Token.Token.AccessToken != token.AccessToken {
		log.Logf("🎟️ Got new Google Access Token (expires: %s)\n", token.Expiry.Format(time.RFC822))
		c.Token.Token = token
		return SaveToken(c.Token.Token)
	}
	return nil
}
