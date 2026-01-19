package types

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	"github.com/bakito/gws/internal/log"
)

const TokenFileName = "token.yaml"

type TokenStorage struct {
	Token oauth2.Token `yaml:"token"`
}

func GetTokenFilePath() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	tokenDir := filepath.Join(userHomeDir, ConfigDir)
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		return "", err
	}

	return filepath.Join(tokenDir, TokenFileName), nil
}

func LoadToken() (*oauth2.Token, error) {
	tokenPath, err := GetTokenFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // No token yet
		}
		return nil, err
	}

	var storage TokenStorage
	err = yaml.Unmarshal(data, &storage)
	if err != nil {
		return nil, err
	}

	return &storage.Token, nil
}

func SaveToken(token oauth2.Token) error {
	tokenPath, err := GetTokenFilePath()
	if err != nil {
		return err
	}

	storage := TokenStorage{Token: token}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	err = encoder.Encode(storage)
	if err != nil {
		return err
	}

	log.Logf("🎟️ Got new Google Access Token (expires: %s)", token.Expiry.Format(time.RFC822))
	return os.WriteFile(tokenPath, buf.Bytes(), 0o600)
}

func SetToken(token oauth2.Token) error {
	// Check if token has changed before saving
	existingToken, err := LoadToken()
	if err != nil {
		return err
	}

	if existingToken != nil && existingToken.AccessToken == token.AccessToken {
		return nil // No change
	}

	return SaveToken(token)
}
