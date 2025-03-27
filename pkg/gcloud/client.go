package gcloud

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"os"
	"strings"
	"time"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"

	"github.com/bakito/gws/pkg/spinner"
	"github.com/bakito/gws/pkg/types"
)

func StartWorkstation(cfg *types.Config) error {
	sshContext, ctx, c, ws, err := setup(cfg)
	if err != nil {
		return err
	}
	defer c.Close()

	switch ws.GetState() {
	case workstationspb.Workstation_STATE_STOPPED:
		op, err := c.StartWorkstation(ctx, &workstationspb.StartWorkstationRequest{Name: ws.GetName()})
		if err != nil {
			_, _ = fmt.Printf("Error starting workstation: %v\n", err)
			return err
		}
		spinny := spinner.Start(fmt.Sprintf(" Waiting for workstation %s to start...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic
		_, err = op.Wait(ctx)
		spinny.Stop()
		if err != nil {
			_, _ = fmt.Printf("Error waiting for workstation to start: %v\n", err)
			return err
		}
		_, _ = fmt.Printf("Workstation started %q\n", sshContext.GCloud.Name)
	case workstationspb.Workstation_STATE_RUNNING:
		_, _ = fmt.Printf("Workstation running %q\n", sshContext.GCloud.Name)
	case workstationspb.Workstation_STATE_STARTING:
		spinny := spinner.Start(fmt.Sprintf(" Workstation %s is already starting ...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic
		for range 10 {
			time.Sleep(10 * time.Second)
			ws, err = c.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{Name: ws.GetName()})
			if err != nil {
				_, _ = fmt.Printf("Error getting workstation: %v\n", err)
				return err
			}
			if ws.GetState() == workstationspb.Workstation_STATE_RUNNING {
				break
			}
		}
		spinny.Stop()
		if ws.GetState() == workstationspb.Workstation_STATE_RUNNING {
			_, _ = fmt.Printf("Workstation started %q\n", sshContext.GCloud.Name)
		} else {
			_, _ = fmt.Printf("Workstation is in unexpected state: %s\n", ws.GetState())
		}
	}
	return nil
}

func setup(cfg *types.Config) (*types.Context, context.Context, *workstations.Client, *workstationspb.Workstation, error) {
	sshContext := cfg.CurrentContext()
	if sshContext.GCloud == nil {
		_, _ = fmt.Println("No gcloud config found")
		return nil, nil, nil, nil, nil
	}
	// gcloud auth application-default login
	ctx := context.TODO()

	// Default credentials: ${HOME}/.config/gcloud/application_default_credentials.json
	tokenSource, err := getToken()
	if err != nil {
		_, _ = fmt.Printf("Error getting OAUTH token: %v\n", err)
		return nil, nil, nil, nil, err
	}

	c, err := workstations.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		_, _ = fmt.Printf("Error creating workstations client: %v\n", err)
		return nil, nil, nil, nil, err
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
		_, _ = fmt.Printf("Error getting workstation: %v\n", err)
		return nil, nil, nil, nil, err
	}
	return sshContext, ctx, c, ws, err
}

func getToken() (tsrc oauth2.TokenSource, err error) {

	token := &oauth2.Token{}

	loadExistingToken(token)

	if token.ExpiresIn < 10*60 {
		token, err = Login()
		if err != nil {
			return nil, err
		}
	}
	// Create an OAuth2 token source
	tokenSource := oauth2.StaticTokenSource(token)
	return tokenSource, nil
}

func loadExistingToken(token *oauth2.Token) {
	_, data, err := types.ReadGWSFile(tokenFileName)
	if err != nil {
		// re-login
		return
	}

	if err := json.Unmarshal(data, token); err != nil {
		// re-login
		return
	}

	token.ExpiresIn = int64(token.Expiry.Sub(time.Now()).Seconds())
}

func StopWorkstation(cfg *types.Config) error {
	sshContext, ctx, c, ws, err := setup(cfg)
	if err != nil {
		return err
	}

	defer c.Close()

	if ws.GetState() != workstationspb.Workstation_STATE_STOPPED {
		op, err := c.StopWorkstation(ctx, &workstationspb.StopWorkstationRequest{Name: ws.GetName()})
		if err != nil {
			_, _ = fmt.Printf("Error stopping workstation: %v\n", err)
			return err
		}
		spinny := spinner.Start(fmt.Sprintf(" Waiting for workstation %s to stop...", sshContext.GCloud.Name))
		defer spinny.Stop() // reset the terminal in case of a panic

		_, err = op.Wait(ctx)
		if err != nil {
			_, _ = fmt.Printf("Error waiting for workstation to stop: %v\n", err)
			return err
		}
		spinny.Stop()
	}
	_, _ = fmt.Printf("Workstation stopped %q\n", sshContext.GCloud.Name)
	return nil
}

func stringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		_, _ = fmt.Fprintf(os.Stderr, "%s ", label)
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func DeleteWorkstation(cfg *types.Config) error {
	name := stringPrompt(
		fmt.Sprintf(
			"Please confirm the deletion of workstation %q by entering the name again:",
			cfg.CurrentContext().GCloud.Name,
		),
	)
	if name != cfg.CurrentContext().GCloud.Name {
		_, _ = fmt.Println("Aborting ...")
		return nil
	}

	sshContext, ctx, c, ws, err := setup(cfg)
	if err != nil {
		return err
	}

	defer c.Close()

	op, err := c.DeleteWorkstation(ctx, &workstationspb.DeleteWorkstationRequest{Name: ws.GetName()})
	if err != nil {
		_, _ = fmt.Printf("Error deleting workstation: %v\n", err)
		return err
	}
	spinny := spinner.Start(fmt.Sprintf(" Deleting workstation %s to stop...", sshContext.GCloud.Name))
	defer spinny.Stop() // reset the terminal in case of a panic

	_, err = op.Wait(ctx)
	if err != nil {
		_, _ = fmt.Printf("Error waiting for workstation to be deleted: %v\n", err)
		return err
	}
	spinny.Stop()
	return nil
}
