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
	submit
	maxFocusable
)

var (
	indigo    = lipgloss.Color("63")
	hotPink   = lipgloss.Color("205")
	darkGray  = lipgloss.Color("240")
	lightGray = lipgloss.Color("244")
)

type styles struct {
	Border         lipgloss.Style
	Label          lipgloss.Style
	Help           lipgloss.Style
	Err            lipgloss.Style
	ErrText        lipgloss.Style
	Focused        lipgloss.Style
	Blurred        lipgloss.Style
	NoStyle        lipgloss.Style
	button         lipgloss.Style
	inputFocused   lipgloss.Style
	inputUnfocused lipgloss.Style
}

func defaultStyles() *styles {
	s := new(styles)
	s.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		Padding(1, 2)
	s.Label = lipgloss.NewStyle().
		Bold(true).
		Foreground(darkGray).
		Padding(0, 2)
	s.Help = lipgloss.NewStyle().
		Foreground(lightGray)
	s.Err = lipgloss.NewStyle().
		Foreground(hotPink)
	s.ErrText = s.Err.Bold(true)
	s.Focused = lipgloss.NewStyle().
		Foreground(hotPink)
	s.Blurred = lipgloss.NewStyle().
		Foreground(lightGray)
	s.NoStyle = lipgloss.NewStyle()
	s.button = lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(hotPink).
		Padding(0, 3).
		MarginTop(1)
	s.inputFocused = lipgloss.NewStyle().
		Foreground(hotPink)
	s.inputUnfocused = lipgloss.NewStyle()
	return s
}

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
	inputs        []input
	focused       focusable
	aborted       bool
	statusMessage string
	config        *types.Config
	styles        *styles
	help          string
}

type input struct {
	textinput.Model
	label string
}

func initialModel() model {
	m := model{
		inputs: make([]input, maxFocusable-1),
		config: &types.Config{
			Contexts: make(map[string]*types.Context),
		},
		styles: defaultStyles(),
		help:   "tab: next field / up: prev field / esc: quit / enter: confirm",
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
		t.Cursor.Style = m.styles.inputFocused
		t.CharLimit = 32
		t.Width = 50

		switch focusable(i) {
		case contextName:
			m.inputs[i].label = "Context Name"
			t.Focus()
			t.PromptStyle = m.styles.inputFocused
			t.TextStyle = m.styles.inputFocused
		case host:
			m.inputs[i].label = "Host"
			t.SetValue("localhost")
		case port:
			m.inputs[i].label = "Port"
			t.CharLimit = 5
		case user:
			m.inputs[i].label = "User"
			t.SetValue("user")
		case privateKeyFile:
			m.inputs[i].label = "Private Key File"
			t.CharLimit = 128
			if userHomeDir != "" {
				t.SetValue(filepath.Join(userHomeDir, ".ssh", "id_rsa"))
			}
		case knownHostsFile:
			m.inputs[i].label = "Known Hosts File (optional)"
			t.CharLimit = 128
			if userHomeDir != "" {
				t.SetValue(filepath.Join(userHomeDir, ".ssh", "known_hosts"))
			}
		case gcloudProject:
			m.inputs[i].label = "gcloud: Project"
		case gcloudRegion:
			m.inputs[i].label = "gcloud: Region"
		case gcloudCluster:
			m.inputs[i].label = "gcloud: Cluster"
		case gcloudConfig:
			m.inputs[i].label = "gcloud: Config"
		case gcloudName:
			m.inputs[i].label = "gcloud: Name"
		default:
			// This should not be reached as maxFocusable defines the number of inputs.
		}
		t.Placeholder = m.inputs[i].label
		m.inputs[i].Model = t
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

			// check if the user wants to submit
			if s == "enter" && m.focused == submit {
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

				for i := range m.inputs {
					if m.inputs[i].Value() == "" && focusable(i) != knownHostsFile {
						m.statusMessage = fmt.Sprintf("Error: %s is a required field.", m.inputs[i].label)
						return m, nil
					}
				}

				privateKeyFile := m.inputs[privateKeyFile].Value()
				if _, err := os.Stat(privateKeyFile); os.IsNotExist(err) {
					m.statusMessage = "Error: private key file does not exist: " + privateKeyFile
					return m, nil
				}

				knownHostsFile := m.inputs[knownHostsFile].Value()
				if knownHostsFile != "" {
					if _, err := os.Stat(knownHostsFile); os.IsNotExist(err) {
						m.statusMessage = "Error: known hosts file does not exist: " + knownHostsFile
						return m, nil
					}
				}

				return m, tea.Quit
			}

			// Clear status message on navigation
			m.statusMessage = ""

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused > submit {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = submit
			}

			for i := range len(m.inputs) {
				if i == int(m.focused) {
					m.inputs[i].Focus()
					m.inputs[i].PromptStyle = m.styles.inputFocused
					m.inputs[i].TextStyle = m.styles.inputFocused
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = m.styles.inputUnfocused
					m.inputs[i].TextStyle = m.styles.inputUnfocused
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
		m.inputs[i].Model, _ = m.inputs[i].Update(msg)
		cmds = append(cmds, nil) // no-op to satisfy the interface
	}
	return tea.Batch(cmds...)
}

func (i input) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, i.label, i.Model.View())
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(m.styles.Label.Render("Create a new gws context"))
	b.WriteString("\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteString("\n")
	}

	var button string
	if m.focused == submit {
		button = m.styles.button.Render("[ Submit ]")
	} else {
		button = fmt.Sprintf("[ %s ]", m.styles.Blurred.Render("Submit"))
	}
	b.WriteString(fmt.Sprintf("\n%s\n\n", button))

	if m.statusMessage != "" {
		b.WriteString(m.styles.ErrText.Render(m.statusMessage))
		b.WriteString("\n\n")
	}

	b.WriteString(m.styles.Help.Render(m.help))

	return m.styles.Border.Render(b.String())
}

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
