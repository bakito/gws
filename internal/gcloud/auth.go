package gcloud

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/phayes/freeport"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/types"
)

var oauthConfig = &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/cloud-platform",
	},
	Endpoint: google.Endpoint,
}

// Generate PKCE Code Verifier and SHA-256 Code Challenge.
func generatePKCE() (codeVerifier, codeChallenge string, err error) {
	verifierBytes := make([]byte, 32)
	_, err = rand.Read(verifierBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	codeVerifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Create the SHA-256 hash of the verifier
	hash := sha256.Sum256([]byte(codeVerifier))

	// Base64 URL encode the hash to create the code challenge
	codeChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return codeVerifier, codeChallenge, nil
}

func Login(ctx context.Context, cfg *types.Config) (oauth2.TokenSource, error) {
	existingToken := cfg.Token.Token

	// Try refreshing the token
	if existingToken.RefreshToken != "" {
		tokenSource := oauthConfig.TokenSource(ctx, &existingToken)
		token, err := tokenSource.Token()
		if err == nil {
			_ = cfg.SetToken(*token)
			return newTokenSourceWithRefreshCheck(ctx, token, cfg), nil
		}
	}

	codeVerifier, codeChallenge, err := generatePKCE()
	if err != nil {
		return nil, err
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	// Use a per-request copy so we don't race other callers that may rely on oauthConfig
	localOAuth := *oauthConfig
	//nolint:revive // http is ok for a local callback
	localOAuth.RedirectURL = fmt.Sprintf("http://%s/", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))

	state := "state"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		state = base64.RawURLEncoding.EncodeToString(b)
	}

	// Add PKCE to auth URL.
	options := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}

	prompt := "select_account"
	if sshContext := cfg.CurrentContext(); sshContext != nil && sshContext.GCloud != nil && sshContext.GCloud.Account != "" {
		prompt = "consent"
		options = append(options, oauth2.SetAuthURLParam("login_hint", sshContext.GCloud.Account))
	}
	options = append(options, oauth2.SetAuthURLParam("prompt", prompt))

	authURL := localOAuth.AuthCodeURL(state, options...)

	// Open URL in browser
	log.Log("Opening URL: " + authURL)
	openBrowser(authURL)

	// Create a channel for shutdown signaling that can carry a token or an error.
	type authResult struct {
		token *oauth2.Token
		err   error
	}
	shutdownChan := make(chan authResult, 1)

	// Use a dedicated ServeMux to avoid interfering with global handlers.
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 1 * time.Second,
		Handler:           mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("state") != state {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			shutdownChan <- authResult{nil, errors.New("invalid state")}
			return
		}
		code := query.Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			shutdownChan <- authResult{nil, errors.New("missing code")}
			return
		}

		token, err := localOAuth.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier),
			oauth2.SetAuthURLParam("client_secret", localOAuth.ClientSecret),
		)
		if err != nil {
			http.Error(w, "Failed to get token", http.StatusInternalServerError)
			log.Logf("🚨 OAuth exchange error: %v", err)
			shutdownChan <- authResult{nil, err}
			return
		}

		// Save token (may not contain a refresh token if consent not granted)
		if err := cfg.SetToken(*token); err != nil {
			// log save error but continue - we still return the token for in-memory usage
			log.Logf("Failed to persist token: %v", err)
		}

		fmt.Fprint(w, "Authentication successful! You can now close this window and return to the tool.")
		shutdownChan <- authResult{token, nil}
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Report start-up failure back
			shutdownChan <- authResult{nil, fmt.Errorf("failed to start server: %w", err)}
		}
	}()

	log.Log("Waiting for authentication...")
	// Block until we receive a shutdown signal
	res := <-shutdownChan
	_ = server.Shutdown(ctx)
	if res.err != nil {
		return nil, res.err
	}
	log.Log("Authenticated...")

	// Warn if the refresh token was not provided.
	if res.token.RefreshToken == "" {
		log.Log("Warning: no refresh token returned. You may need to re-auth with prompt=consent to get a refresh token.")
	}

	return newTokenSourceWithRefreshCheck(ctx, res.token, cfg), nil
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
	newCtx, cancel := context.WithCancel(ctx)
	ts := &TokenSourceWithRefreshCheck{
		checkPeriod: 10 * time.Minute,
		source:      oauthConfig.TokenSource(newCtx, token),
		cfg:         cfg,
		done:        make(chan struct{}),
		cancel:      cancel,
	}

	if cfg.TokenCheck {
		// Start periodic check using the cancelable context so Stop() cancels it.
		go ts.periodicCheck(newCtx)
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
	if ts.cancel != nil {
		ts.cancel()
	}
	// Close done non-blocking, protect against double close
	select {
	case <-ts.done:
		// already closed / drained
	default:
		close(ts.done)
	}
}
