package tunnel

import (
	"context"

	"github.com/bakito/gws/internal/types"
)

type Model struct {
	Config   *types.Config
	Port     int
	Styles   *Styles
	Width    int
	Height   int
	Logs     []string
	Err      error
	Quitting bool

	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc

	LogChan chan string
}

func NewModel(ctx context.Context, cfg *types.Config, port int) Model {
	c, cancel := context.WithCancel(ctx)
	return Model{
		Config:  cfg,
		Port:    port,
		Styles:  DefaultStyles(),
		ctx:     c,
		cancel:  cancel,
		LogChan: make(chan string, 10),
	}
}

type (
	logMsg string
	errMsg struct{ err error }
)
