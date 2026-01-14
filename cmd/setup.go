package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/bakito/gws/internal/types"
)

const (
	contextName focusable = iota
	host
	port
	user
	privateKeyFile
	knownHostsFile
	maxFocusable
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create a new or update the config.yaml and create a context configuration",
	Long:  `Create a new or update the config.yaml and create a context configuration using an interactive terminal setup wizard.`,
	RunE:  setup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func setup(cmd *cobra.Command, args []string) error {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return err
	}

	if m.(model).aborted {
		fmt.Println("Setup aborted.")
		return nil
	}

	return saveConfig(m.(model))
}

type focusable int

type model struct {
	inputs  []textinput.Model
	focused focusable
	aborted bool
}

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, maxFocusable),
	}

	for i := range m.inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		t.CharLimit = 32
		t.Width = 50

		switch focusable(i) {
		case contextName:
			t.Placeholder = "Context Name"
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		case host:
			t.Placeholder = "Host"
		case port:
			t.Placeholder = "Port"
			t.CharLimit = 5
		case user:
			t.Placeholder = "User"
		case privateKeyFile:
			t.Placeholder = "Private Key File"
			t.CharLimit = 128
		case knownHostsFile:
			t.Placeholder = "Known Hosts File"
			t.CharLimit = 128
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focused == maxFocusable-1 {
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > maxFocusable-1 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = maxFocusable - 1
			}

			for i := range len(m.inputs) {
				if i == int(m.focused) {
					m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = lipgloss.NewStyle()
					m.inputs[i].TextStyle = lipgloss.NewStyle()
				}
			}

			return m, nil
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.inputs {
		m.inputs[i], _ = m.inputs[i].Update(msg)
		cmds = append(cmds, nil) // no-op to satisfy the interface
	}
	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("Create a new gws context\n\n")
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	button := &blurredButton
	if m.focused == maxFocusable-1 {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

var (
	focusedButton = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Submit"))
)

func saveConfig(m model) error {
	portVal, err := strconv.Atoi(m.inputs[port].Value())
	if err != nil {
		portVal = 22
	}
	newCtx := &types.Context{
		Host:           m.inputs[host].Value(),
		Port:           portVal,
		User:           m.inputs[user].Value(),
		PrivateKeyFile: m.inputs[privateKeyFile].Value(),
		KnownHostsFile: m.inputs[knownHostsFile].Value(),
	}

	ctxName := m.inputs[contextName].Value()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(userHomeDir, types.ConfigDir)
	configPath := filepath.Join(configDir, types.ConfigFileName)

	var config types.Config
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(data, &config); err != nil {
			return err
		}
	} else {
		config.Contexts = make(map[string]*types.Context)
	}

	config.Contexts[ctxName] = newCtx
	config.CurrentContextName = ctxName
	config.FilePath = configPath

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	fmt.Printf("Writing new config to %s\n", configPath)
	return os.WriteFile(configPath, data, 0o600)
}
