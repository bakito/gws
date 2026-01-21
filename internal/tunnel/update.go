package tunnel

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/bakito/gws/internal/gcloud"
	"github.com/bakito/gws/internal/log"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return startTunnelMsg{}
		},
		m.waitForLog(),
	)
}

func (m Model) startTunnel() tea.Cmd {
	log.SetLogger(func(log string) {
		select {
		case m.LogChan <- log:
		default:
		}
	})

	return func() tea.Msg {
		err := gcloud.TCPTunnelWithPassphrase(m.ctx, m.Config, m.Port)
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.Quitting = true
			m.cancel()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case startTunnelMsg:
		return m, m.startTunnel()
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
	startTunnelMsg struct{}
)
