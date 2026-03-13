package script

import (
	"bytes"
	"text/template"

	"github.com/bakito/gws/internal/types"

	_ "embed"
)

//go:embed reconnect-win.cmd
var windowsReconnectScript string

func WindowsReconnectSSH(cfg *types.Config) (string, error) {
	tmpl, err := template.New("reconnect").Parse(windowsReconnectScript)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf,
		&NamedContext{Name: cfg.CurrentContextName, Context: *cfg.CurrentContext()},
	); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type NamedContext struct {
	types.Context
	Name string
}
