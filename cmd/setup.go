package cmd

import (
	"errors"
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
	gcloudProject
	gcloudRegion
	gcloudCluster
	gcloudConfig
	gcloudName
	maxFocusable
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create a new or update the config.yaml and create a context configuration",
	Long: `Create a new or update the config.yaml and create a context configuration using an
interactive terminal setup wizard.`,
	RunE: setup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func setup(_ *cobra.Command, _ []string) error {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return err
	}

	if model, ok := m.(model); ok {
		if model.aborted {
			fmt.Println("Setup aborted.")
			return nil
		}
		return saveConfig(model)
	}
	return errors.New("could not assert model type")
}

type focusable int

type model struct {
	inputs        []textinput.Model
	focused       focusable
	aborted       bool
	statusMessage string
	config        *types.Config
}

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, maxFocusable),
		config: &types.Config{
			Contexts: make(map[string]*types.Context),
		},
	}

	userHomeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(userHomeDir, types.ConfigDir, types.ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err == nil {
				_ = yaml.Unmarshal(data, m.config)
			}
		}
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
			t.SetValue("localhost")
		case port:
			t.Placeholder = "Port"
			t.CharLimit = 5
		case user:
			t.Placeholder = "User"
			t.SetValue("user")
		case privateKeyFile:
			t.Placeholder = "Private Key File"
			t.CharLimit = 128
		case knownHostsFile:
			t.Placeholder = "Known Hosts File (optional)"
			t.CharLimit = 128
		case gcloudProject:
			t.Placeholder = "gcloud: Project"
		case gcloudRegion:
			t.Placeholder = "gcloud: Region"
		case gcloudCluster:
			t.Placeholder = "gcloud: Cluster"
		case gcloudConfig:
			t.Placeholder = "gcloud: Config"
		case gcloudName:
			t.Placeholder = "gcloud: Name"
		default:
			// This case should not be reached as maxFocusable defines the number of inputs.
			// Adding a default to satisfy the linter.
		}

		m.inputs[i] = t
	}

	return m
}

func (model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Clear status message on navigation
			m.statusMessage = ""

			// Check current field for validity before moving
			if m.focused == contextName {
				ctxName := m.inputs[contextName].Value()
				if _, ok := m.config.Contexts[ctxName]; ok {
					m.statusMessage = fmt.Sprintf("Error: Context name %q already exists.", ctxName)
					return m, nil
				}
			}

			if m.focused == port {
				portVal, err := strconv.Atoi(m.inputs[port].Value())
				if err != nil || portVal < 1000 || portVal > 65535 {
					m.statusMessage = fmt.Sprintf(
						"Error: Port must be a number between 1000 and 65535. (Current value: %s)",
						m.inputs[port].Value(),
					)
					return m, nil
				}
				for name, ctx := range m.config.Contexts {
					if ctx.Port == portVal {
						m.statusMessage = fmt.Sprintf("Error: Port %d is already used by context %q.", portVal, name)
						return m, nil
					}
				}
			} else if m.focused != knownHostsFile && m.inputs[m.focused].Value() == "" {
				m.statusMessage = fmt.Sprintf("Error: %s is a required field.", m.inputs[m.focused].Placeholder)
				return m, nil
			}

			if s == "enter" && m.focused == maxFocusable-1 {
				// Final validation before quitting
				ctxName := m.inputs[contextName].Value()
				if _, ok := m.config.Contexts[ctxName]; ok {
					m.statusMessage = fmt.Sprintf("Error: Context name %q already exists.", ctxName)
					return m, nil
				}

				portVal, err := strconv.Atoi(m.inputs[port].Value())
				if err != nil || portVal < 1000 || portVal > 65535 {
					m.statusMessage = fmt.Sprintf(
						"Error: Port must be a number between 1001 and 65535. (Current value: %s)",
						m.inputs[port].Value(),
					)
					return m, nil
				}

				for name, ctx := range m.config.Contexts {
					if ctx.Port == portVal {
						m.statusMessage = fmt.Sprintf("Error: Port %d is already used by context %q.", portVal, name)
						return m, nil
					}
				}

				if m.inputs[m.focused].Value() == "" && m.focused != knownHostsFile {
					m.statusMessage = fmt.Sprintf("Error: %s is a required field.", m.inputs[m.focused].Placeholder)
					return m, nil
				}
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

			for i := range m.inputs {
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

	if m.statusMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.statusMessage))
		b.WriteString("\n\n")
	}

	button := &blurredButton
	if m.focused == maxFocusable-1 {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "%s\n\n", *button)

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

	gcloudProject := m.inputs[gcloudProject].Value()
	gcloudRegion := m.inputs[gcloudRegion].Value()
	gcloudCluster := m.inputs[gcloudCluster].Value()
	gcloudConfig := m.inputs[gcloudConfig].Value()
	gcloudName := m.inputs[gcloudName].Value()

	if gcloudProject != "" || gcloudRegion != "" || gcloudCluster != "" || gcloudConfig != "" || gcloudName != "" {
		newCtx.GCloud = &types.GCloud{
			Project: gcloudProject,
			Region:  gcloudRegion,
			Cluster: gcloudCluster,
			Config:  gcloudConfig,
			Name:    gcloudName,
		}
	}

	ctxName := m.inputs[contextName].Value()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(userHomeDir, types.ConfigDir)
	configPath := filepath.Join(configDir, types.ConfigFileName)

	config := m.config
	config.Contexts[ctxName] = newCtx
	config.CurrentContextName = ctxName
	config.FilePath = configPath

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	fmt.Printf("Writing new config to %s\n", configPath)
	return os.WriteFile(configPath, data, 0o600)
}
