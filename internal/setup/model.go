package setup

import (
	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/bakito/gws/internal/types"
)

const (
	ContextName focusable = iota
	Port
	User
	PrivateKeyFile
	KnownHostsFile
	GcloudProject
	GcloudAccount
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
