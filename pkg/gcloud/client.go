package gcloud

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/bakito/gws/pkg/types"
)

func StartWorkstation(sshContext types.Context) {
	if sshContext.GCloud == nil {
		slog.Info("No gcloud config found")
		return
	}
	// https://cloud.google.com/go/docs/reference/cloud.google.com/go/workstations/latest/apiv1

	// gcloud auth application-default login
	ctx := context.TODO()

	c, err := workstations.NewClient(ctx)
	if err != nil {
		slog.Error("Error creating workstations client", "error", err)
	}
	defer c.Close()

	wsName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstations/%s",
		sshContext.GCloud.Project,
		sshContext.GCloud.Region,
		sshContext.GCloud.Cluster,
		sshContext.GCloud.Config,
		sshContext.GCloud.Name,
	)

	ws, err := c.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{Name: wsName})
	if err != nil {
		slog.Error("Error getting workstation", "error", err)
		os.Exit(1)
	}

	if ws.State == workstationspb.Workstation_STATE_STOPPED {
		op, err := c.StartWorkstation(ctx, &workstationspb.StartWorkstationRequest{Name: wsName})
		if err != nil {
			slog.Error("Error starting workstation", "error", err)
			os.Exit(1)
		}
		slog.Info("Waiting for workstation to start", "name", sshContext.GCloud.Name)
		_, err = op.Wait(ctx)
		if err != nil {
			slog.Error("Error waiting for workstation to start", "error", err)
			os.Exit(1)
		}
		slog.Info("Workstation started", "name", sshContext.GCloud.Name)
	} else if ws.State == workstationspb.Workstation_STATE_RUNNING {
		slog.Info("Workstation running", "name", sshContext.GCloud.Name)
	} else if ws.State == workstationspb.Workstation_STATE_STARTING {
		slog.Info("Workstation is already starting", "name", sshContext.GCloud.Name)

		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Second)
			ws, err = c.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{Name: wsName})
			if err != nil {
				slog.Error("Error getting workstation", "error", err)
				os.Exit(1)
			}
			if ws.State == workstationspb.Workstation_STATE_RUNNING {
				break
			}
		}
		if ws.State == workstationspb.Workstation_STATE_RUNNING {
			slog.Info("Workstation started", "name", sshContext.GCloud.Name)
		} else {
			slog.Error("Workstation is in unexpected state", "state", ws.State)
		}
	}
}

func StopWorkstation(sshContext types.Context) {
	if sshContext.GCloud == nil {
		slog.Info("No gcloud config found")
		return
	}
	// https://cloud.google.com/go/docs/reference/cloud.google.com/go/workstations/latest/apiv1

	// gcloud auth application-default login
	ctx := context.TODO()

	c, err := workstations.NewClient(ctx)
	if err != nil {
		slog.Error("Error creating workstations client", "error", err)
	}
	defer c.Close()

	wsName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstations/%s",
		sshContext.GCloud.Project,
		sshContext.GCloud.Region,
		sshContext.GCloud.Cluster,
		sshContext.GCloud.Config,
		sshContext.GCloud.Name,
	)

	ws, err := c.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{Name: wsName})
	if err != nil {
		slog.Error("Error getting workstation", "error", err)
		os.Exit(1)
	}

	if ws.State != workstationspb.Workstation_STATE_STOPPED {
		op, err := c.StopWorkstation(ctx, &workstationspb.StopWorkstationRequest{Name: wsName})
		if err != nil {
			slog.Error("Error stopping workstation", "error", err)
			os.Exit(1)
		}
		slog.Info("Waiting for workstation to stop", "name", sshContext.GCloud.Name)
		_, err = op.Wait(ctx)
		if err != nil {
			slog.Error("Error waiting for workstation to stop", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Workstation stopped", "name", sshContext.GCloud.Name)
}
