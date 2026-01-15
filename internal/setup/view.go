package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const title = `██████╗ ██╗    ██╗███████╗
██╔════╝ ██║    ██║██╔════╝
██║  ███╗██║ █╗ ██║███████╗
██║   ██║██║███╗██║╚════██║
╚██████╔╝╚███╔███╔╝███████║
 ╚═════╝  ╚══╝╚══╝ ╚══════╝`

func (i Input) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, i.Style.Render(i.Label), i.Model.View())
}

func (m Model) View() string {
	if m.FilePickerActive {
		help := m.Styles.Help.Render("enter: select / backspace: directory up / esc: quit")
		fpView := lipgloss.JoinVertical(lipgloss.Left, m.Fp.View(), help)
		return m.Styles.Border.Width(m.Width - 4).Render(fpView)
	}
	var b strings.Builder

	b.WriteString(m.Styles.Title.Render(title))
	b.WriteString("\n\n")

	for i := range m.Inputs {
		b.WriteString(m.Inputs[i].View())
		b.WriteString("\n")
	}

	var button string
	if m.Focused == Submit {
		button = m.Styles.Button.Render("[ Submit ]")
	} else {
		button = fmt.Sprintf("[ %s ]", m.Styles.Blurred.Render("Submit"))
	}
	b.WriteString(fmt.Sprintf("\n%s\n\n", button))

	if m.StatusMessage != "" {
		b.WriteString(m.Styles.ErrText.Render(m.StatusMessage))
		b.WriteString("\n\n")
	}

	help := m.Help
	if m.Focused == PrivateKeyFile || m.Focused == KnownHostsFile {
		help += " / ctrl+f: file picker"
	}
	b.WriteString(m.Styles.Help.Render(help))

	return m.Styles.Border.Width(m.Width - 4).Render(b.String())
}
