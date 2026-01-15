package setup

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/bakito/gws/internal/types"
)

func Run(cfg *types.Config, context string) error {
	cfg.CurrentContextName = context
	p := tea.NewProgram(InitialModel(cfg), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return err
	}

	if model, ok := m.(Model); ok {
		if model.Aborted {
			fmt.Println("Setup aborted.")
			return nil
		}
		return SaveConfig(model)
	}
	return errors.New("could not assert model type")
}
