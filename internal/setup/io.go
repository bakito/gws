package setup

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/types"
)

func SaveConfig(m Model) error {
	portVal, _ := strconv.Atoi(m.Inputs[Port].Value())
	ctxName := m.Inputs[ContextName].Value()
	config := m.Config
	if config.Contexts == nil {
		config.Contexts = make(map[string]*types.Context)
	}
	newCtx := config.Contexts[ctxName]
	if newCtx == nil {
		newCtx = &types.Context{
			Host: "localhost",
		}
	}
	newCtx.Port = portVal
	newCtx.User = m.Inputs[User].Value()
	newCtx.PrivateKeyFile = m.Inputs[PrivateKeyFile].Value()
	newCtx.KnownHostsFile = m.Inputs[KnownHostsFile].Value()

	if newCtx.GCloud == nil {
		newCtx.GCloud = &types.GCloud{}
	}

	newCtx.GCloud.Project = m.Inputs[GcloudProject].Value()
	newCtx.GCloud.Region = m.Inputs[GcloudRegion].Value()
	newCtx.GCloud.Cluster = m.Inputs[GcloudCluster].Value()
	newCtx.GCloud.Config = m.Inputs[GcloudConfig].Value()
	newCtx.GCloud.Name = m.Inputs[GcloudName].Value()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(userHomeDir, types.ConfigDir)
	configPath := filepath.Join(configDir, types.ConfigFileName)

	config.Contexts[ctxName] = newCtx
	config.CurrentContextName = ctxName

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	log.Logf("\n💾 Writing config to %s\n", configPath)
	return config.SwitchContext(ctxName, true)
}
