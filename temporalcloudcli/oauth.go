package temporalcloudcli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

type oauthCallbackResult struct {
	code string
	err  error
}

func Login(ctx context.Context, config *oauth2.Config, options ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// Generate PKCE challenge.
	verifier := oauth2.GenerateVerifier()
	options = append(options, oauth2.S256ChallengeOption(verifier))

	// Generate random state for CSRF protection.
	var stateBytes [16]byte
	if _, err := rand.Read(stateBytes[:]); err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes[:])

	authURL := config.AuthCodeURL(state, options...)
	fmt.Printf("Opening browser to authorize. If it doesn't open, visit: %s\n", authURL)
	_ = browser.OpenURL(authURL)

	// Parse redirect URL to get host and path
	url, err := url.Parse(config.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URL: %w", err)
	}

	// Start HTTP server to handle callback.
	// AIDEV-NOTE: serverErrCh captures server startup errors to prevent silent failures
	var once sync.Once
	resultCh := make(chan oauthCallbackResult, 1)
	serverErrCh := make(chan error, 1)
	server := &http.Server{
		Addr: url.Host,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			// Use sync.Once to only process the first callback.
			once.Do(func() {
				query := r.URL.Query()

				// Check for OAuth error response.
				if errCode := query.Get("error"); errCode != "" {
					errDesc := query.Get("error_description")
					if errDesc == "" {
						errDesc = errCode
					}
					resultCh <- oauthCallbackResult{err: fmt.Errorf("authorization failed: %s", errDesc)}
					http.Error(w, fmt.Sprintf("Authorization failed: %s", errDesc), http.StatusBadRequest)
					return
				}

				// Validate state to prevent CSRF.
				if query.Get("state") != state {
					resultCh <- oauthCallbackResult{err: fmt.Errorf("invalid state parameter")}
					http.Error(w, "Invalid state parameter", http.StatusBadRequest)
					return
				}

				// Check for authorization code.
				code := query.Get("code")
				if code == "" {
					resultCh <- oauthCallbackResult{err: fmt.Errorf("missing authorization code")}
					http.Error(w, "Missing authorization code", http.StatusBadRequest)
					return
				}

				resultCh <- oauthCallbackResult{code: code}
				fmt.Fprint(w, "Authorization successful! You can close this window.")
			})
		}),
	}
	// AIDEV-NOTE: Start server in goroutine and capture startup errors
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- fmt.Errorf("failed to start OAuth callback server: %w", err)
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	// Wait for callback result, server error, or context cancellation.
	var result oauthCallbackResult
	select {
	case result = <-resultCh:
	case err := <-serverErrCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if result.err != nil {
		return nil, result.err
	}

	// Exchange code for token with PKCE verifier.
	token, err := config.Exchange(ctx, result.code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

var ErrLoginRequired = errors.New("login required")

func GetToken(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, bool, error) {
	if token != nil {
		// Check if the token is still valid
		if token.Valid() {
			return token, false, nil
		}
	}

	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		if requiresLogin(err) {
			return nil, false, ErrLoginRequired
		}
		return nil, false, fmt.Errorf("failed to refresh token: %w", err)
	}
	return newToken, true, nil
}

// requiresLogin checks if the error indicates an invalid or expired refresh token.
func requiresLogin(err error) bool {
	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) && retrieveErr.ErrorCode == "invalid_grant" {
		return true
	}
	return false
}
