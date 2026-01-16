package tunnel

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/bakito/gws/internal/gcloud"
	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/ssh"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			needs, err := ssh.NeedsPassphrase(m.Config.CurrentContext().PrivateKeyFile)
			if err != nil {
				return errMsg{err}
			}
			if needs {
				return passphraseNeededMsg{}
			}
			return startTunnelMsg{}
		},
		m.waitForLog(),
	)
}

func (m Model) startTunnel(passphrase string) tea.Cmd {
	log.SetLogger(func(log string) {
		select {
		case m.LogChan <- log:
		default:
		}
	})

	return func() tea.Msg {
		err := gcloud.TCPTunnelWithPassphrase(m.ctx, m.Config, m.Port, passphrase)
		if err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m Model) waitForLog() tea.Cmd {
	return func() tea.Msg {
		l, ok := <-m.LogChan
		if !ok {
			return nil
		}
		return logMsg(l)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.AskingPassphrase {
			switch msg.String() {
			case "enter":
				m.AskingPassphrase = false
				return m, m.startTunnel(m.PassphraseInput.Value())
			case "ctrl+c", "esc":
				m.Quitting = true
				m.cancel()
				return m, tea.Quit
			}
			m.PassphraseInput, cmd = m.PassphraseInput.Update(msg)
			return m, cmd
		}

		if msg.String() == "ctrl+c" {
			m.Quitting = true
			m.cancel()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case passphraseNeededMsg:
		m.AskingPassphrase = true
		return m, textinput.Blink
	case startTunnelMsg:
		return m, m.startTunnel("")
	case logMsg:
		m.Logs = append(m.Logs, string(msg))
		return m, m.waitForLog()
	case errMsg:
		m.Err = msg.err
		return m, nil
	}
	return m, nil
}

type (
	passphraseNeededMsg struct{}
	startTunnelMsg      struct{}
)
