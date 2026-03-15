package script

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/bakito/gws/internal/types"

	_ "embed"
)

//go:embed reconnect-win.cmd
var windowsReconnectScript string

func WindowsReconnectSSH(cfg *types.Config, fileName string) (string, error) {
	tmpl, err := template.New("reconnect").Parse(windowsReconnectScript)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf,
		&NamedContext{
			Name:     cfg.CurrentContextName,
			FileName: outFileName(fileName),
			Context:  *cfg.CurrentContext(),
		},
	); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func outFileName(name string) string {
	if name == "" {
		return ""
	}
	return filepath.Base(name)
}

type NamedContext struct {
	types.Context
	Name     string
	FileName string
}
