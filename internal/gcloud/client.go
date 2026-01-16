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
	"google.golang.org/api/option"

	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/spinner"
	"github.com/bakito/gws/internal/types"
)

const (
	pollInterval    = 10 * time.Second
	maxPollAttempts = 10
	defaultTimeout  = pollInterval * maxPollAttempts
)

func StartWorkstation(ctx context.Context, cfg *types.Config) error {
	sshContext, c, ws, err := setup(ctx, cfg)
	if err != nil {
		return err
	}
	defer c.Close()

	start := time.Now()

	switch ws.GetState() {
	case workstationspb.Workstation_STATE_STOPPED:
		op, err := c.StartWorkstation(ctx, &workstationspb.StartWorkstationRequest{Name: ws.GetName()})
		if err != nil {
			log.Logf("Error starting workstation: %v\n", err)
			return err
		}
		spinny := spinner.Start(fmt.Sprintf(" Waiting for workstation %s to start...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic
		_, err = op.Wait(ctx)
		spinny.Stop()
		if err != nil {
			log.Logf("Error waiting for workstation to start: %v\n", err)
			return err
		}
		log.Logf("Workstation started in %s %q\n", time.Since(start).String(), sshContext.GCloud.Name)
	case workstationspb.Workstation_STATE_RUNNING:
		log.Logf("Workstation running %q\n", sshContext.GCloud.Name)
	case workstationspb.Workstation_STATE_STARTING:
		spinny := spinner.Start(fmt.Sprintf(" Workstation %s is already starting ...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic

		err = waitForWorkstationRunning(ctx, c, ws, defaultTimeout)
		spinny.Stop()

		if err != nil {
			return err
		}

		if ws.GetState() == workstationspb.Workstation_STATE_RUNNING {
			log.Logf("Workstation started in %s %q\n", time.Since(start).String(), sshContext.GCloud.Name)
		} else {
			log.Logf("Workstation is in unexpected state: %s\n", ws.GetState())
		}
	default:
	}
	return nil
}

// waitForWorkstationRunning polls the workstation status until it's running or timeout occurs.
// Returns error if the workstation fails to reach in running state within the specified timeout.
func waitForWorkstationRunning(
	ctx context.Context,
	c *workstations.Client,
	ws *workstationspb.Workstation,
	timeout time.Duration,
) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-timeoutCh:
			return fmt.Errorf("timeout waiting for workstation %s to start", ws.GetName())
		case <-ticker.C:
			updatedWs, err := c.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{Name: ws.GetName()})
			if err != nil {
				return fmt.Errorf("failed to get workstation status: %w", err)
			}

			if updatedWs.GetState() == workstationspb.Workstation_STATE_RUNNING {
				*ws = workstationspb.Workstation{
					Name:  updatedWs.GetName(),
					State: updatedWs.GetState(),
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func setup(ctx context.Context, cfg *types.Config) (*types.Context, *workstations.Client, *workstationspb.Workstation, error) {
	sshContext := cfg.CurrentContext()
	if sshContext.GCloud == nil {
		log.Log("No gcloud config found")
		return nil, nil, nil, nil
	}
	// gcloud auth application-default login
	// Default credentials: ${HOME}/.config/gcloud/application_default_credentials.json
	tokenSource, err := Login(ctx, cfg)
	if err != nil {
		log.Logf("Error getting OAUTH token: %v\n", err)
		return nil, nil, nil, err
	}

	c, err := workstations.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Logf("Error creating workstations client: %v\n", err)
		return nil, nil, nil, err
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
		log.Logf("Error getting workstation: %v\n", err)
		return nil, nil, nil, err
	}
	return sshContext, c, ws, err
}

func StopWorkstation(ctx context.Context, cfg *types.Config) error {
	sshContext, c, ws, err := setup(ctx, cfg)
	if err != nil {
		return err
	}

	defer c.Close()

	if ws.GetState() != workstationspb.Workstation_STATE_STOPPED {
		op, err := c.StopWorkstation(ctx, &workstationspb.StopWorkstationRequest{Name: ws.GetName()})
		if err != nil {
			log.Logf("Error stopping workstation: %v\n", err)
			return err
		}
		spinny := spinner.Start(fmt.Sprintf(" Waiting for workstation %s to stop...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic

		_, err = op.Wait(ctx)
		if err != nil {
			log.Logf("Error waiting for workstation to stop: %v\n", err)
			return err
		}
		spinny.Stop()
	}
	log.Logf("Workstation stopped %q\n", sshContext.GCloud.Name)
	return nil
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

func DeleteWorkstation(ctx context.Context, cfg *types.Config) error {
	name := stringPrompt(
		fmt.Sprintf(
			"Please confirm the deletion of workstation %q by entering the name again:",
			cfg.CurrentContext().GCloud.Name,
		),
	)
	if name != cfg.CurrentContext().GCloud.Name {
		log.Log("Aborting ...")
		return nil
	}

	sshContext, c, ws, err := setup(ctx, cfg)
	if err != nil {
		return err
	}

	defer c.Close()

	op, err := c.DeleteWorkstation(ctx, &workstationspb.DeleteWorkstationRequest{Name: ws.GetName()})
	if err != nil {
		log.Logf("Error deleting workstation: %v\n", err)
		return err
	}
	spinny := spinner.Start(fmt.Sprintf(" Deleting workstation %s to stop...", sshContext.GCloud.Name))
	defer spinny.Stop() // reset the terminal in case of a panic

	_, err = op.Wait(ctx)
	if err != nil {
		log.Logf("Error waiting for workstation to be deleted: %v\n", err)
		return err
	}
	spinny.Stop()
	return nil
}
