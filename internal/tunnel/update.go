package tunnel

import (
	"github.com/bakito/gws/internal/gcloud"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			err := gcloud.TCPTunnel(m.ctx, m.Config, m.Port, func(log string) {
				select {
				case m.LogChan <- log:
				default:
				}
			})
			if err != nil {
				return errMsg{err}
			}
			return nil
		},
		m.waitForLog(),
	)
}

func (m Model) waitForLog() tea.Cmd {
	return func() tea.Msg {
		return logMsg(<-m.LogChan)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.Quitting = true
			m.cancel()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case logMsg:
		m.Logs = append(m.Logs, string(msg))
		return m, m.waitForLog()
	case errMsg:
		m.Err = msg.err
		return m, nil
	}
	return m, nil
}
