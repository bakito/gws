package spinner

import (
	"fmt"
	"math/rand"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var spinners = []spinner.Spinner{
	spinner.Line,
	spinner.Dot,
	spinner.MiniDot,
	spinner.Jump,
	spinner.Pulse,
	spinner.Points,
	spinner.Globe,
	spinner.Moon,
	spinner.Monkey,
	spinner.Meter,
}

// Spinner is a wrapper around a bubbletea program.
type Spinner struct {
	p    *tea.Program
	done chan struct{}
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	if s.p != nil {
		s.p.Quit()
		<-s.done
	}
}

// Start starts a new spinner with the given title.
func Start(title string, sp ...spinner.Spinner) *Spinner {
	s := spinner.New()
	if len(sp) > 0 {
		s.Spinner = sp[0]
	} else {
		s.Spinner = spinners[rand.Intn(len(spinners))]
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m := model{spinner: s, title: " " + title}
	p := tea.NewProgram(m)
	done := make(chan struct{})
	go func() {
		_, _ = p.Run()
		close(done)
	}()
	return &Spinner{p: p, done: done}
}

type model struct {
	spinner  spinner.Model
	title    string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.QuitMsg:
		m.quitting = true
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	return tea.NewView(fmt.Sprintf("%s%s", m.spinner.View(), m.title))
}
