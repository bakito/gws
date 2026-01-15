package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

func (Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.FilePickerActive {
		var cmd tea.Cmd
		m.Fp, cmd = m.Fp.Update(msg)

		if didSelect, path := m.Fp.DidSelectFile(msg); didSelect {
			m.Inputs[m.FilePickerField].SetValue(path)
			m.FilePickerActive = false
			return m, nil
		}
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
			m.FilePickerActive = false
			return m, nil
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if (m.Focused == PrivateKeyFile || m.Focused == KnownHostsFile) && msg.String() == "ctrl+f" {
			m.FilePickerActive = true
			m.FilePickerField = m.Focused
			m.StatusMessage = ""
			currentFile := m.Inputs[m.Focused].Value()
			if currentFile != "" {
				if info, err := os.Stat(currentFile); err == nil && !info.IsDir() {
					m.Fp.CurrentDirectory = filepath.Dir(currentFile)
				} else if err == nil && info.IsDir() {
					m.Fp.CurrentDirectory = currentFile
				}
			}
			return m, m.Fp.Init()
		}
		switch msg.String() {
		case "ctrl+c", "esc":
			m.Aborted = true
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// check if the user wants to submit
			if s == "enter" && m.Focused == Submit {
				return m.validateAndSubmit()
			}

			// Clear status message on navigation
			m.StatusMessage = ""

			if s == "up" || s == "shift+tab" {
				m.Focused--
			} else {
				m.Focused++
			}

			if m.Focused > Submit {
				m.Focused = 0
			} else if m.Focused < 0 {
				m.Focused = Submit
			}

			for i := range m.Inputs {
				if i == int(m.Focused) {
					m.Inputs[i].Focus()
					m.Inputs[i].PromptStyle = m.Styles.InputFocused
					m.Inputs[i].TextStyle = m.Styles.InputFocused
				} else {
					m.Inputs[i].Blur()
					m.Inputs[i].PromptStyle = m.Styles.InputUnfocused
					m.Inputs[i].TextStyle = m.Styles.InputUnfocused
				}
			}

			return m, nil
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m Model) validateAndSubmit() (tea.Model, tea.Cmd) {
	ctxName := m.Inputs[ContextName].Value()
	portVal, err := strconv.Atoi(m.Inputs[Port].Value())
	if err != nil || portVal < 1000 || portVal > 65535 {
		m.StatusMessage = fmt.Sprintf(
			"Error: Port must be a number between 1001 and 65535. (Current value: %s)",
			m.Inputs[Port].Value(),
		)
		return m, nil
	}

	if m.Config.Contexts != nil {
		for name, ctx := range m.Config.Contexts {
			if ctx.Port == portVal && name != ctxName {
				m.StatusMessage = fmt.Sprintf("Error: Port %d is already used by context %q.", portVal, name)
				return m, nil
			}
		}
	}

	for i := range m.Inputs {
		if m.Inputs[i].Value() == "" && focusable(i) != KnownHostsFile {
			m.StatusMessage = fmt.Sprintf("Error: %s is a required field.", m.Inputs[i].Label)
			return m, nil
		}
	}

	privateKeyFile := m.Inputs[PrivateKeyFile].Value()
	if st, err := os.Stat(privateKeyFile); os.IsNotExist(err) {
		m.StatusMessage = "Error: private key file does not exist: " + privateKeyFile
		return m, nil
	} else if st.IsDir() {
		m.StatusMessage = "Error: private key must not be directory: " + privateKeyFile
		return m, nil
	}

	knownHostsFile := m.Inputs[KnownHostsFile].Value()
	if knownHostsFile != "" {
		if st, err := os.Stat(knownHostsFile); os.IsNotExist(err) {
			m.StatusMessage = "Error: known hosts file does not exist: " + knownHostsFile
			return m, nil
		} else if st.IsDir() {
			m.StatusMessage = "Error: known hosts must not be directory: " + knownHostsFile
			return m, nil
		}
	}

	return m, tea.Quit
}

func (m Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Inputs))
	for i := range m.Inputs {
		m.Inputs[i].Model, cmds[i] = m.Inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}
