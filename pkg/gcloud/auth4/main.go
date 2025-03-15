package main

import (
	"context"
	"log"

	workstations "cloud.google.com/go/workstations/apiv1"
	"github.com/bakito/gws/pkg/gcloud"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

func main() {
	token, _ := gcloud.Login()
	// Create an OAuth2 token source
	tokenSource := oauth2.StaticTokenSource(token)

	client, err := workstations.NewClient(context.TODO(), option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatalf("Failed to create Workstations client: %v", err)
	}
	defer client.Close()
}
