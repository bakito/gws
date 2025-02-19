package cmd

import (
	"fmt"
	"github.com/bakito/gws/pkg/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"maps"
	"slices"
)

// ctxCmd represents the ctx command
var ctxCmd = &cobra.Command{
	Use:   "ctx",
	Short: "Set the active context",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}
		if len(args) == 1 {
			return cfg.SwitchContext(args[0])
		}
		selected, err := selectContext(cfg)
		if err != nil {
			return err
		}

		if selected != "" {
			return cfg.SwitchContext(selected)
		}

		return nil
	},
}

func selectContext(cfg *types.Config) (string, error) {
	m := ctxModel{choices: slices.Sorted(maps.Keys(cfg.Contexts))}
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return "", err
	}
	return m.selected, nil
}

func init() {
	rootCmd.AddCommand(ctxCmd)
}

type ctxModel struct {
	cursor   int
	choices  []string
	selected string
}

func (m ctxModel) Init() tea.Cmd {
	return nil
}

func (m ctxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ctxModel) View() string {
	s := "Choose an option:\n\n"
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	s += "\nPress ↑/↓ to move, Enter to select, Q to quit.\n"
	return s
}
