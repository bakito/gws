package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/types"
)

const (
	contextName focusable = iota
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
	cfg, err := loadConfig()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	cfg.CurrentContextName = flagContext

	p := tea.NewProgram(initialModel(cfg))
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
	inputs           []input
	focused          focusable
	aborted          bool
	statusMessage    string
	config           *types.Config
	styles           *styles
	help             string
	fp               filepicker.Model
	filePickerActive bool
	filePickerField  focusable
	width            int
	height           int
}

type input struct {
	textinput.Model
	label string
}

func initialModel(cfg *types.Config) model {
	m := model{
		inputs: make([]input, maxFocusable-1),
		config: cfg,
		styles: defaultStyles(),
		help:   "tab: next field / up: prev field / esc: quit / enter: confirm",
	}

	configDir, userHomeDir := types.DefaultConfigDir()

	if m.config.FilePath == "" {
		m.config.FilePath = configDir
	}

	context := &types.Context{GCloud: &types.GCloud{}}
	if m.config.CurrentContextName != "" {
		if m.config.Contexts != nil {
			c := m.config.Contexts[m.config.CurrentContextName]
			if c != nil {
				context = c
			}
		}
	}

	fp := filepicker.New()
	fp.SetHeight(20)
	fp.ShowHidden = true            // Show hidden files (e.g., .ssh)	fp.KeyMap.Back.SetKeys("backspace") // Explicitly set key for going up a directory
	fp.KeyMap.Open.SetKeys("enter") // Explicitly set key for going down into a directory
	startDir := ""
	if context.PrivateKeyFile != "" {
		startDir = filepath.Dir(context.PrivateKeyFile)
	} else if userHomeDir != "" {
		startDir = filepath.Join(userHomeDir, ".ssh")
	}

	if startDir != "" {
		if _, err := os.Stat(startDir); !os.IsNotExist(err) {
			fp.CurrentDirectory = startDir
		}
	}
	m.fp = fp

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
			t.SetValue(m.config.CurrentContextName)
		case port:
			m.inputs[i].label = "Local SSH Port"
			t.CharLimit = 5
			if context.Port > 0 {
				t.SetValue(strconv.Itoa(context.Port))
			}
		case user:
			m.inputs[i].label = "Cloud Workstation User"
			if context.User != "" {
				t.SetValue(context.User)
			} else {
				t.SetValue("user")
			}
		case privateKeyFile:
			m.inputs[i].label = "Private Key File"
			t.CharLimit = 128
			if context.PrivateKeyFile != "" {
				t.SetValue(context.PrivateKeyFile)
			} else if userHomeDir != "" {
				t.SetValue(filepath.Join(userHomeDir, ".ssh"))
			}
		case knownHostsFile:
			m.inputs[i].label = "Known Hosts File (optional)"
			t.CharLimit = 128
			if context.KnownHostsFile != "" {
				t.SetValue(context.KnownHostsFile)
			} else if userHomeDir != "" {
				t.SetValue(filepath.Join(userHomeDir, ".ssh"))
			}
		case gcloudProject:
			m.inputs[i].label = "gcloud: Project ID"
			t.SetValue(context.GCloud.Project)
		case gcloudRegion:
			m.inputs[i].label = "gcloud: Region"
			t.SetValue(context.GCloud.Region)
		case gcloudCluster:
			m.inputs[i].label = "gcloud: Cluster"
			t.SetValue(context.GCloud.Cluster)
		case gcloudConfig:
			m.inputs[i].label = "gcloud: Config"
			t.SetValue(context.GCloud.Config)
		case gcloudName:
			m.inputs[i].label = "gcloud: Workstation ID"
			t.SetValue(context.GCloud.Name)
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
	if m.filePickerActive {
		var cmd tea.Cmd
		m.fp, cmd = m.fp.Update(msg)

		if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
			m.inputs[m.filePickerField].SetValue(path)
			m.filePickerActive = false
			return m, nil
		}
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
			m.filePickerActive = false
			return m, nil
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if (m.focused == privateKeyFile || m.focused == knownHostsFile) && msg.String() == "ctrl+f" {
			m.filePickerActive = true
			m.filePickerField = m.focused
			m.statusMessage = ""
			currentFile := m.inputs[m.focused].Value()
			if currentFile != "" {
				if info, err := os.Stat(currentFile); err == nil && !info.IsDir() {
					m.fp.CurrentDirectory = filepath.Dir(currentFile)
				} else if err == nil && info.IsDir() {
					m.fp.CurrentDirectory = currentFile
				}
			}
			return m, m.fp.Init()
		}
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
				portVal, err := strconv.Atoi(m.inputs[port].Value())
				if err != nil || portVal < 1000 || portVal > 65535 {
					m.statusMessage = fmt.Sprintf(
						"Error: Port must be a number between 1001 and 65535. (Current value: %s)",
						m.inputs[port].Value(),
					)
					return m, nil
				}

				if m.config.Contexts != nil {
					for name, ctx := range m.config.Contexts {
						if ctx.Port == portVal && name != ctxName {
							m.statusMessage = fmt.Sprintf("Error: Port %d is already used by context %q.", portVal, name)
							return m, nil
						}
					}
				}

				for i := range m.inputs {
					if m.inputs[i].Value() == "" && focusable(i) != knownHostsFile {
						m.statusMessage = fmt.Sprintf("Error: %s is a required field.", m.inputs[i].label)
						return m, nil
					}
				}

				privateKeyFile := m.inputs[privateKeyFile].Value()
				if st, err := os.Stat(privateKeyFile); os.IsNotExist(err) {
					m.statusMessage = "Error: private key file does not exist: " + privateKeyFile
					return m, nil
				} else if st.IsDir() {
					m.statusMessage = "Error: private key must not be directory: " + privateKeyFile
					return m, nil
				}

				knownHostsFile := m.inputs[knownHostsFile].Value()
				if knownHostsFile != "" {
					if st, err := os.Stat(knownHostsFile); os.IsNotExist(err) {
						m.statusMessage = "Error: known hosts file does not exist: " + knownHostsFile
						return m, nil
					} else if st.IsDir() {
						m.statusMessage = "Error: known hosts must not be directory: " + knownHostsFile
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

			for i := range m.inputs {
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i].Model, cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (i input) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, i.label, i.Model.View())
}

func (m model) View() string {
	if m.filePickerActive {
		help := m.styles.Help.Render("enter: select / backspace: directory up / esc: quit")
		return lipgloss.JoinVertical(lipgloss.Left, m.styles.Border.Render(m.fp.View()), help)
	}
	var b strings.Builder

	b.WriteString(m.styles.Label.Render(">_ Modify the gws context"))
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

	help := m.help
	if m.focused == privateKeyFile || m.focused == knownHostsFile {
		help = help + " / ctrl+f: file picker"
	}
	b.WriteString(m.styles.Help.Render(help))

	return m.styles.Border.Width(m.width - 4).Render(b.String())
}

func saveConfig(m model) error {
	portVal, _ := strconv.Atoi(m.inputs[port].Value())
	ctxName := m.inputs[contextName].Value()
	config := m.config
	if config.Contexts == nil {
		config.Contexts = make(map[string]*types.Context)
	}
	newCtx := config.Contexts[ctxName]
	if newCtx == nil {
		newCtx = &types.Context{
			Host: "localhost",
		}
	}
	newCtx.Port = portVal
	newCtx.User = m.inputs[user].Value()
	newCtx.PrivateKeyFile = m.inputs[privateKeyFile].Value()
	newCtx.KnownHostsFile = m.inputs[knownHostsFile].Value()

	if newCtx.GCloud == nil {
		newCtx.GCloud = &types.GCloud{}
	}

	newCtx.GCloud.Project = m.inputs[gcloudProject].Value()
	newCtx.GCloud.Region = m.inputs[gcloudRegion].Value()
	newCtx.GCloud.Cluster = m.inputs[gcloudCluster].Value()
	newCtx.GCloud.Config = m.inputs[gcloudConfig].Value()
	newCtx.GCloud.Name = m.inputs[gcloudName].Value()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(userHomeDir, types.ConfigDir)
	configPath := filepath.Join(configDir, types.ConfigFileName)

	config.Contexts[ctxName] = newCtx
	config.CurrentContextName = ctxName

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	fmt.Printf("\n💾 Writing config to %s\n", configPath)
	return config.SwitchContext(ctxName, true)
}
