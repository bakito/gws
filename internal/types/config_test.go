package types

import (
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestConfig_Validate(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with multiple contexts",
			config: &Config{
				Contexts: map[string]*Context{
					"context1": {Host: "localhost", Port: 8080, Files: []File{
						{SourcePath: tmpFile.Name(), Path: "/path/to/dest"},
					}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config with missing contexts",
			config: &Config{
				Contexts: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid context fields",
			config: &Config{
				Contexts: map[string]*Context{
					"context1": {Host: "", Port: 0, Files: nil},
				},
			},
			wantErr: true,
		},
		{
			name: "missing required file fields",
			config: &Config{
				Contexts: map[string]*Context{
					"context1": {Host: "localhost", Port: 8080, Files: []File{
						{SourcePath: "", Path: "/path/to/dest"},
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid config with one context",
			config: &Config{
				Contexts: map[string]*Context{
					"context1": {Host: "example.com", Port: 22, User: "user", Files: []File{
						{SourcePath: tmpFile.Name(), Path: "/valid/dest"},
					}},
				},
				CurrentContextName: "context1",
			},
			wantErr: false,
		},
		{
			name: "chrome executable missing",
			config: &Config{
				Contexts: map[string]*Context{
					"context1": {Host: "localhost", Port: 8080, Files: []File{
						{SourcePath: tmpFile.Name(), Path: "/path/to/dest"},
					}},
				},
				ChromeBrowser: &ChromeBrowserConfig{ProfileDirectory: "foo"},
			},
			wantErr: true,
		},
		{
			name: "chrome profile missing",
			config: &Config{
				Contexts: map[string]*Context{
					"context1": {Host: "localhost", Port: 8080, Files: []File{
						{SourcePath: tmpFile.Name(), Path: "/path/to/dest"},
					}},
				},
				ChromeBrowser: &ChromeBrowserConfig{ExecutablePath: tmpFile.Name()},
			},
			wantErr: true,
		},
		{
			name: "valid chrome config",
			config: &Config{
				Contexts: map[string]*Context{
					"context1": {Host: "localhost", Port: 8080, Files: []File{
						{SourcePath: tmpFile.Name(), Path: "/path/to/dest"},
					}},
				},
				ChromeBrowser: &ChromeBrowserConfig{
					ExecutablePath:   tmpFile.Name(),
					ProfileDirectory: "foo",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			err := validate.Struct(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
