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
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/phayes/freeport"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const tokenFileName = ".gws-token.json"

var oauthConfig = &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	Scopes: []string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/appengine.admin",
		"https://www.googleapis.com/auth/sqlservice.login",
		"https://www.googleapis.com/auth/compute",
	},
	Endpoint: google.Endpoint,
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

func Login() (oauth2.TokenSource, error) {
	existingToken := &oauth2.Token{}

	loadExistingToken(existingToken)

	if existingToken.ExpiresIn > 10*60 {
		return oauth2.StaticTokenSource(existingToken), nil
	}
	// Try refreshing the token
	if existingToken.RefreshToken != "" {
		tokenSource := oauthConfig.TokenSource(context.Background(), existingToken)
		token, err := tokenSource.Token()
		if err == nil {
			saveToken(token)
			return oauthConfig.TokenSource(context.Background(), token), nil
		}
	}

	codeVerifier, codeChallenge := generatePKCE()

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	oauthConfig.RedirectURL = fmt.Sprintf("http://%s/callback", net.JoinHostPort("localhost", strconv.Itoa(port)))
	println(oauthConfig.RedirectURL)

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

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), ReadHeaderTimeout: 1 * time.Second}
	// Start local server to handle callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		code := query.Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}

		// Exchange authorization code for token

		token, err := oauthConfig.Exchange(context.Background(), code,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier),
			oauth2.SetAuthURLParam("client_secret", oauthConfig.ClientSecret),
		)
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
	_, _ = fmt.Println("Authenticated...")
	_ = server.Shutdown(context.Background())
	return oauth2.StaticTokenSource(token), nil
}

// Save token to file.
func saveToken(token *oauth2.Token) {
	b, err := json.Marshal(token)
	if err != nil {
		log.Fatalf("Failed to save token: %v", err)
	}

	err = os.WriteFile(tokenFileName, b, 0o600)
	if err != nil {
		log.Fatalf("Failed to save token: %v", err)
	}
}
