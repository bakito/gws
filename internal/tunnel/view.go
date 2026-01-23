package tunnel

import (
	"fmt"
	"strings"

	"github.com/bakito/gws/version"
)

func (m Model) View() string {
	if m.Quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.Styles.Title.Render(">_ GWS Tunnel"))
	b.WriteString("\n\n")

	currCtx := m.Config.CurrentContext()
	port := m.Port
	if port == 0 {
		port = currCtx.Port
	}
	b.WriteString(fmt.Sprintf("  Version: %s\n", version.Version))
	b.WriteString(fmt.Sprintf("  Context: %s\n", m.Styles.Success.Render(m.Config.CurrentContextName)))
	b.WriteString(fmt.Sprintf("  Workstation: %s\n", m.Styles.Success.Render(currCtx.GCloud.Name)))
	b.WriteString(fmt.Sprintf("  Local Port: %d\n", port))
	b.WriteString("\n")

	if m.Err != nil {
		b.WriteString(m.Styles.ErrText.Render(fmt.Sprintf("Error: %v", m.Err)))
		b.WriteString("\n\n")
	}

	b.WriteString(m.Styles.Info.Render("Logs:"))
	b.WriteString("\n")

	var logView string
	if len(m.Logs) > 0 {
		start := 0
		if len(m.Logs) > 10 {
			start = len(m.Logs) - 10
		}
		logView = strings.Join(m.Logs[start:], "\n")
	} else {
		logView = "Waiting for connections..."
	}
	b.WriteString(logView)
	b.WriteString("\n\n")

	help := m.Styles.Help.Render("ctrl+c: quit")
	b.WriteString(help)

	return m.Styles.Border.Width(m.Width - 4).Render(b.String())
}
