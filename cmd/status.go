package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/spf13/cobra"

	"github.com/bakito/gws/internal/gcloud"
	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/spinner"
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List all configured workstations with their state",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := readConfig()
		if err != nil {
			return err
		}

		spinner.Disable()
		log.SetLogger(log.Null)
		states, err := gcloud.GetWorkstationStates(cmd.Context(), cfg)
		log.SetLogger(log.Stdout)
		if err != nil {
			return err
		}

		slices.SortFunc(states, func(a, b gcloud.WorkstationState) int {
			return strings.Compare(a.Context, b.Context)
		})

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CONTEXT\tNAME\tSTATE\tUPTIME")
		for _, s := range states {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Context, s.Name, formatState(s.State), formatUptime(s.Uptime))
		}
		return w.Flush()
	},
}

func formatState(s workstationspb.Workstation_State) string {
	state := s.String()
	state = strings.ReplaceAll(state, "Workstation_STATE_", "")
	state = strings.ReplaceAll(state, "STATE_", "")
	switch s {
	case workstationspb.Workstation_STATE_RUNNING:
		return "🟢 " + state
	case workstationspb.Workstation_STATE_STOPPED:
		return "🔴 " + state
	case workstationspb.Workstation_STATE_STARTING, workstationspb.Workstation_STATE_STOPPING:
		return "🟡 " + state
	default:
		return "⚪ " + state
	}
}

func formatUptime(u *time.Duration) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
