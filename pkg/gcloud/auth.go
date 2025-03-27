package gcloud

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuth2 Config.
var oauthConfig = &oauth2.Config{
	ClientID:     "415121532721-l5h1pcvq0r3va06a7uv6633eco3hbevh.apps.googleusercontent.com",
	ClientSecret: "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",
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8080/callback",
	Scopes:       []string{"openid", "email", "profile"},
}

// Generate PKCE Code Verifier and SHA-256 Code Challenge.
func generatePKCE() (codeVerifier, codeChallenge string) {
	// Create a random 43-128 character code verifier
	verifierBytes := make([]byte, 32)
	_, err := rand.Read(verifierBytes)
	if err != nil {
		log.Fatalf("Failed to generate PKCE verifier: %v", err)
	}
	codeVerifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Create the SHA-256 hash of the verifier
	hash := sha256.Sum256([]byte(codeVerifier))

	// Base64 URL encode the hash to create the code challenge
	codeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return codeVerifier, codeChallenge
}

func Login() (*oauth2.Token, error) {
	codeVerifier, codeChallenge := generatePKCE()

	// Add PKCE to auth URL
	authURL := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	// Open URL in browser
	_, _ = fmt.Println("Opening URL:", authURL)
	openBrowser(authURL)

	// Create a channel for shutdown signaling
	shutdownChan := make(chan *oauth2.Token)

	server := &http.Server{Addr: ":8080", ReadHeaderTimeout: 1 * time.Second}
	// Start local server to handle callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		code := query.Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}

		// Exchange authorization code for token
		token, err := oauthConfig.Exchange(context.Background(), code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		if err != nil {
			http.Error(w, "Failed to get token", http.StatusInternalServerError)
			log.Fatalf("OAuth Exchange error: %v", err)
		}

		// Save token
		saveToken(token)

		_, _ = fmt.Fprint(w, "Authentication successful! You can close this window.")
		// Signal shutdown using a channel
		go func() {
			shutdownChan <- token
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	_, _ = fmt.Println("Waiting for authentication...")
	// Block until we receive a shutdown signal
	token := <-shutdownChan
	_, _ = fmt.Println("Shutting down server...")
	_ = server.Shutdown(context.Background())
	return token, nil
}

// Save token to file.
func saveToken(token *oauth2.Token) {
	b, err := json.Marshal(token)
	if err != nil {
		log.Fatalf("Failed to save token: %v", err)
	}

	err = os.WriteFile("token.json", b, 0o600)
	if err != nil {
		log.Fatalf("Failed to save token: %v", err)
	}

	_, _ = fmt.Println("Token saved to token.json")
}
