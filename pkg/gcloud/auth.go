package gcloud

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/phayes/freeport"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/bakito/gws/pkg/types"
)

var oauthConfig = &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	Scopes: []string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/appengine.admin",
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

func Login(ctx context.Context, cfg *types.Config) (oauth2.TokenSource, error) {
	existingToken := cfg.Token

	// Try refreshing the token
	if existingToken.RefreshToken != "" {
		tokenSource := oauthConfig.TokenSource(ctx, &existingToken)
		token, err := tokenSource.Token()
		if err == nil {
			_ = cfg.SetToken(*token)
			return newTokenSourceWithRefreshCheck(ctx, token, cfg), nil
		}
	}

	codeVerifier, codeChallenge := generatePKCE()

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	// nolint: revive // http is ok for local callback
	oauthConfig.RedirectURL = fmt.Sprintf("http://%s/callback", net.JoinHostPort("localhost", strconv.Itoa(port)))

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
	// Start a local server to handle callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		code := query.Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}

		// Exchange authorization code for token

		token, err := oauthConfig.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier),
			oauth2.SetAuthURLParam("client_secret", oauthConfig.ClientSecret),
		)
		if err != nil {
			http.Error(w, "Failed to get token", http.StatusInternalServerError)
			log.Fatalf("OAuth Exchange error: %v", err)
		}

		// Save token
		_ = cfg.SetToken(*token)

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
	_ = server.Shutdown(ctx)
	return newTokenSourceWithRefreshCheck(ctx, token, cfg), nil
}

type TokenSourceWithRefreshCheck struct {
	source      oauth2.TokenSource
	checkPeriod time.Duration
	lastToken   *oauth2.Token
	done        chan struct{}
	cancel      context.CancelFunc
	cfg         *types.Config
}

func newTokenSourceWithRefreshCheck(ctx context.Context, token *oauth2.Token, cfg *types.Config) oauth2.TokenSource {
	_, cancel := context.WithCancel(ctx)
	ts := &TokenSourceWithRefreshCheck{
		checkPeriod: 10 * time.Minute,
		source:      oauthConfig.TokenSource(ctx, token),
		cfg:         cfg,
		done:        make(chan struct{}),
		cancel:      cancel,
	}

	if cfg.TokenCheck {
		// Start periodic check
		go ts.periodicCheck(ctx)
	}
	return ts
}

func (ts *TokenSourceWithRefreshCheck) periodicCheck(ctx context.Context) {
	ticker := time.NewTicker(ts.checkPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.done:
			return
		case <-ticker.C:
			token, err := ts.source.Token()
			if err != nil {
				continue
			}

			if ts.lastToken == nil || ts.lastToken.AccessToken != token.AccessToken {
				_ = ts.cfg.SetToken(*token)
				ts.lastToken = token
			}
		}
	}
}

func (ts *TokenSourceWithRefreshCheck) Token() (*oauth2.Token, error) {
	token, err := ts.source.Token()
	if err != nil {
		return nil, err
	}

	// Also check for refresh during direct Token() calls
	if ts.lastToken == nil || ts.lastToken.AccessToken != token.AccessToken {
		_ = ts.cfg.SetToken(*token)
		ts.lastToken = token
	}

	return token, nil
}

// Stop stops the periodic check.
func (ts *TokenSourceWithRefreshCheck) Stop() {
	ts.cancel()
	close(ts.done)
}
