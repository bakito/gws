package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/workstations/apiv1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var oauth2Config = &oauth2.Config{
	ClientID:     "",
	ClientSecret: "",
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8080/callback", // Local web server for OAuth
	Scopes:       []string{"https://www.googleapis.com/auth/cloud-platform"},
}

var authCodeCh = make(chan string)

func main() {
	ctx := context.Background()

	// Start a temporary web server to receive the auth code
	go startAuthServer()

	// Step 1: Print auth URL and prompt user to open it
	authURL := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Open the following URL in your browser:")
	fmt.Println(authURL)

	// Step 2: Wait for auth code from local web server
	authCode := <-authCodeCh

	// Step 3: Exchange auth code for access token
	token, err := oauth2Config.Exchange(ctx, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token: %v", err)
	}

	// Step 4: Create an authenticated HTTP client
	client := oauth2Config.Client(ctx, token)

	// Step 5: Use the authenticated client with the Workstations API
	workstationClient, err := workstations.NewClient(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Failed to create Workstations client: %v", err)
	}
	defer workstationClient.Close()

}

// startAuthServer runs a local HTTP server to capture the auth code
func startAuthServer() {
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		if authCode == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Authorization successful! You can close this window.")
		authCodeCh <- authCode
	})

	log.Println("Starting local web server on http://localhost:8080/callback")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
