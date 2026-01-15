package setup

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/bakito/gws/internal/types"
)

const (
	ContextName focusable = iota
	Port
	User
	PrivateKeyFile
	KnownHostsFile
	GcloudProject
	GcloudRegion
	GcloudCluster
	GcloudConfig
	GcloudName
	Submit
	maxFocusable
)

type focusable int

type Model struct {
	Inputs           []Input
	Focused          focusable
	Aborted          bool
	StatusMessage    string
	Config           *types.Config
	Styles           *Styles
	Help             string
	Fp               filepicker.Model
	FilePickerActive bool
	FilePickerField  focusable
	Width            int
	Height           int
}

type Input struct {
	textinput.Model
	Label string
	Style lipgloss.Style
}
