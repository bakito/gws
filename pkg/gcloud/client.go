package gcloud

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/bakito/gws/pkg/spinner"
	"github.com/bakito/gws/pkg/types"
)

func StartWorkstation(cfg *types.Config) {
	sshContext, ctx, c, err, ws := setup(cfg)
	if err != nil {
		return
	}
	defer c.Close()

	if ws.State == workstationspb.Workstation_STATE_STOPPED {
		op, err := c.StartWorkstation(ctx, &workstationspb.StartWorkstationRequest{Name: ws.GetName()})
		if err != nil {
			fmt.Printf("Error starting workstation: %v\n", err)
			os.Exit(1)
		}
		spinny := spinner.Start(fmt.Sprintf(" Waiting for workstation %s to start...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic
		_, err = op.Wait(ctx)
		spinny.Stop()
		if err != nil {
			fmt.Printf("Error waiting for workstation to start: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Workstation started %q\n", sshContext.GCloud.Name)
	} else if ws.State == workstationspb.Workstation_STATE_RUNNING {
		fmt.Printf("Workstation running %q\n", sshContext.GCloud.Name)
	} else if ws.State == workstationspb.Workstation_STATE_STARTING {
		spinny := spinner.Start(fmt.Sprintf(" Workstation %s is already starting ...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Second)
			ws, err = c.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{Name: ws.GetName()})
			if err != nil {
				fmt.Printf("Error getting workstation: %v\n", err)
				os.Exit(1)
			}
			if ws.State == workstationspb.Workstation_STATE_RUNNING {
				break
			}
		}
		spinny.Stop()
		if ws.State == workstationspb.Workstation_STATE_RUNNING {
			fmt.Printf("Workstation started %q\n", sshContext.GCloud.Name)
		} else {
			fmt.Printf("Workstation is in unexpected state: %s\n", ws.State)
		}
	}
}

func setup(
	cfg *types.Config,
) (*types.Context, context.Context, *workstations.Client, error, *workstationspb.Workstation) {
	sshContext := cfg.CurrentContext()
	if sshContext.GCloud == nil {
		fmt.Println("No gcloud config found")
		return nil, nil, nil, nil, nil
	}
	// https://cloud.google.com/go/docs/reference/cloud.google.com/go/workstations/latest/apiv1

	// gcloud auth application-default login
	ctx := context.TODO()

	c, err := workstations.NewClient(ctx)
	if err != nil {
		fmt.Printf("Error creating workstations client: %v\n", err)
		os.Exit(1)
	}
	wsName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstations/%s",
		sshContext.GCloud.Project,
		sshContext.GCloud.Region,
		sshContext.GCloud.Cluster,
		sshContext.GCloud.Config,
		sshContext.GCloud.Name,
	)

	ws, err := c.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{Name: wsName})
	if err != nil {
		fmt.Printf("Error getting workstation: %v\n", err)
		os.Exit(1)
	}
	return sshContext, ctx, c, err, ws
}

func StopWorkstation(cfg *types.Config) {
	sshContext, ctx, c, err, ws := setup(cfg)
	if err != nil {
		return
	}

	defer c.Close()

	if ws.State != workstationspb.Workstation_STATE_STOPPED {
		op, err := c.StopWorkstation(ctx, &workstationspb.StopWorkstationRequest{Name: ws.GetName()})
		if err != nil {
			fmt.Printf("Error stopping workstation: %v\n", err)
			os.Exit(1)
		}
		spinny := spinner.Start(fmt.Sprintf(" Waiting for workstation %s to stop...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic

		_, err = op.Wait(ctx)
		if err != nil {
			fmt.Printf("Error waiting for workstation to stop: %v\n", err)
			os.Exit(1)
		}
		spinny.Stop()
	}
	fmt.Printf("Workstation stopped %q\n", sshContext.GCloud.Name)
}

func stringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "%s ", label)
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func DeleteWorkstation(cfg *types.Config) {
	name := stringPrompt(
		fmt.Sprintf(
			"Please confirm the deletion of workstation %q by entering the name again:",
			cfg.CurrentContext().GCloud.Name,
		),
	)
	if name != cfg.CurrentContext().GCloud.Name {
		fmt.Println("Aborting ...")
		return
	}

	sshContext, ctx, c, err, ws := setup(cfg)

	if err != nil {
		return
	}

	defer c.Close()

	op, err := c.DeleteWorkstation(ctx, &workstationspb.DeleteWorkstationRequest{Name: ws.GetName()})
	if err != nil {
		fmt.Printf("Error deleting workstation: %v\n", err)
		os.Exit(1)
	}
	spinny := spinner.Start(fmt.Sprintf(" Deleting workstation %s to stop...", sshContext.GCloud.Name))
	defer spinny.Stop() // reset the terminal in case of a panic

	_, err = op.Wait(ctx)
	if err != nil {
		fmt.Printf("Error waiting for workstation to be deleted: %v\n", err)
		os.Exit(1)
	}
	spinny.Stop()
}
