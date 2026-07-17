package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	resendAPIBase       = "https://api.resend.com"
	oauthRedirectPath   = "/oauth/callback"
	oauthCallbackHost   = "127.0.0.1"
	oauthScope          = "emails:send"
	oauthClientName     = "Pop"
	tokenRefreshRefresh = 5 * time.Minute

	// OAuth string constants (goconst).
	oauthGrantTypeAuthCode     = "authorization_code"
	oauthGrantTypeRefreshToken = "refresh_token"
	oauthResponseType          = "code"
	oauthParamClientID         = "client_id"
	oauthParamRedirectURI      = "redirect_uri"
	oauthParamScope            = "scope"
)

// OAuthToken stores the persisted OAuth tokens and client registration.
type OAuthToken struct {
	ClientID     string    `json:"client_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// expired returns true if the access token has expired or is about to expire.
func (t *OAuthToken) expired() bool {
	return time.Now().Add(tokenRefreshRefresh).After(t.ExpiresAt)
}

// authFilePath returns the path to the OAuth token storage file.
func authFilePath() (string, error) {
	dataDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting data directory: %w", err)
	}
	// On Linux, use XDG_DATA_HOME (defaults to ~/.local/share) for token
	// storage. On macOS and Windows, os.UserConfigDir() is the appropriate
	// location (~Library/Application Support and %AppData% respectively).
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		dataDir = xdgData
	}
	dir := filepath.Join(dataDir, "pop")
	if err := os.MkdirAll(dir, 0o700); err != nil { //nolint:gosec // G703: dataDir is from a trusted source
		return "", fmt.Errorf("creating data directory: %w", err)
	}
	return filepath.Join(dir, "auth.json"), nil
}

// loadAuth reads the persisted OAuth token from disk.
func loadAuth() (*OAuthToken, error) {
	path, err := authFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err //nolint:wrapcheck // caller checks os.IsNotExist
		}
		return nil, fmt.Errorf("reading auth file: %w", err)
	}
	var token OAuthToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing auth file: %w", err)
	}
	return &token, nil
}

// saveAuth writes the OAuth token to disk with restrictive permissions.
func saveAuth(token *OAuthToken) error {
	path, err := authFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(token, "", "  ") //nolint:gosec
	if err != nil {
		return fmt.Errorf("marshaling auth: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing auth file: %w", err)
	}
	return nil
}

// deleteAuth removes the persisted OAuth token.
func deleteAuth() error {
	path, err := authFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("deleting auth file: %w", err)
	}
	return nil
}

// generateCodeVerifier generates a cryptographically random PKCE code verifier.
func generateCodeVerifier() (string, error) {
	b := make([]byte, 64) //nolint:mnd
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge computes the S256 PKCE code challenge from a verifier.
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// generateState generates a random OAuth state parameter.
func generateState() (string, error) {
	b := make([]byte, 24) //nolint:mnd
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// registerClient performs dynamic client registration with Resend.
func registerClient(ctx context.Context, redirectURI string) (string, error) {
	body := map[string]any{
		"client_name":                oauthClientName,
		"redirect_uris":              []string{redirectURI},
		"grant_types":                []string{oauthGrantTypeAuthCode, oauthGrantTypeRefreshToken},
		"response_types":             []string{oauthResponseType},
		"token_endpoint_auth_method": "none",
		oauthParamScope:              oauthScope,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling registration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIBase+"/oauth/register", strings.NewReader(string(data)))
	if err != nil {
		return "", fmt.Errorf("creating registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("registering client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", &HTTPError{Status: resp.Status, Body: string(respBody)}
	}

	var result struct {
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding registration response: %w", err)
	}
	return result.ClientID, nil
}

// HTTPError represents a non-successful HTTP response from the Resend API.
type HTTPError struct {
	Status string // e.g. "429 Too Many Requests"
	Body   string // response body
}

func (e *HTTPError) Error() string {
	return e.Status + ": " + e.Body
}

// tokenResponse is the response from the token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// exchangeCode exchanges an authorization code for tokens.
func exchangeCode(ctx context.Context, clientID, code, redirectURI, codeVerifier string) (*tokenResponse, error) {
	form := url.Values{
		"grant_type":          {oauthGrantTypeAuthCode},
		oauthParamClientID:    {clientID},
		oauthResponseType:     {code},
		oauthParamRedirectURI: {redirectURI},
		"code_verifier":       {codeVerifier},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIBase+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return doTokenRequest(req)
}

// refreshToken exchanges a refresh token for a new access token.
func refreshToken(ctx context.Context, clientID, refreshTok string) (*tokenResponse, error) {
	form := url.Values{
		"grant_type":       {oauthGrantTypeRefreshToken},
		oauthParamClientID: {clientID},
		"refresh_token":    {refreshTok},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIBase+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return doTokenRequest(req)
}

// revokeToken revokes a refresh token.
func revokeToken(ctx context.Context, clientID, token string) error {
	form := url.Values{
		oauthParamClientID: {clientID},
		"token":            {token},
		"token_type_hint":  {oauthGrantTypeRefreshToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIBase+"/oauth/revoke", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoking token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func doTokenRequest(req *http.Request) (*tokenResponse, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{Status: resp.Status, Body: string(body)}
	}

	var tokResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokResp); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	return &tokResp, nil
}

// callbackResult is pushed onto the callback channel by the OAuth callback
// HTTP handler. Both the TUI and the non-interactive CLI flows consume it.
type callbackResult struct {
	code string
	err  error
}

// openCallbackListener binds an ephemeral loopback port for the OAuth redirect
// and returns the listener along with the redirect URI to advertise to the
// provider.
func openCallbackListener(ctx context.Context) (net.Listener, string, error) {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", oauthCallbackHost+":0")
	if err != nil {
		return nil, "", fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://%s:%d%s", oauthCallbackHost, port, oauthRedirectPath)
	return listener, redirectURI, nil
}

// buildAuthURL constructs the Resend authorization endpoint URL with the
// PKCE parameters required by the code flow.
func buildAuthURL(clientID, redirectURI, state, codeChallenge string) (string, error) {
	authURL, err := url.Parse(resendAPIBase + "/oauth/authorize")
	if err != nil {
		return "", fmt.Errorf("parsing authorization URL: %w", err)
	}
	params := url.Values{
		oauthParamClientID:      {clientID},
		"response_type":         {oauthResponseType},
		oauthParamRedirectURI:   {redirectURI},
		oauthParamScope:         {oauthScope},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}
	authURL.RawQuery = params.Encode()
	return authURL.String(), nil
}

// newCallbackServer builds the OAuth callback HTTP server and starts serving
// on the given listener in a goroutine. It pushes the authorization code or
// an error onto the returned channel once the callback fires. The caller is
// responsible for shutting down the server.
func newCallbackServer(state string, listener net.Listener) (*http.Server, chan callbackResult) {
	resultCh := make(chan callbackResult, 1)
	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != oauthRedirectPath {
				http.NotFound(w, r)
				return
			}
			query := r.URL.Query()
			if errVal := query.Get("error"); errVal != "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, "OAuth error")
				resultCh <- callbackResult{err: fmt.Errorf("OAuth error: %s", errVal)}
				return
			}
			code := query.Get("code")
			stateVal := query.Get("state")
			if code == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Missing authorization code"))
				resultCh <- callbackResult{err: errors.New("missing authorization code")}
				return
			}
			if stateVal != state {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("State mismatch"))
				resultCh <- callbackResult{err: errors.New("state mismatch")}
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Authorization successful! You can close this tab."))
			resultCh <- callbackResult{code: code}
		}),
	}
	go func() { _ = server.Serve(listener) }()
	return server, resultCh
}

// getValidAccessToken returns a valid access token, refreshing if necessary.
// If no stored auth exists, it returns an empty string and no error.
func getValidAccessToken() (string, error) {
	token, err := loadAuth()
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	if !token.expired() {
		return token.AccessToken, nil
	}

	// Refresh the token.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tokResp, err := refreshToken(ctx, token.ClientID, token.RefreshToken)
	if err != nil {
		// Refresh failed — the user needs to re-authenticate.
		return "", fmt.Errorf("token refresh failed, please run 'pop auth' again: %w", err)
	}

	token.AccessToken = tokResp.AccessToken
	token.RefreshToken = tokResp.RefreshToken
	token.ExpiresAt = time.Now().Add(time.Duration(tokResp.ExpiresIn) * time.Second)

	if err := saveAuth(token); err != nil {
		return "", fmt.Errorf("saving refreshed token: %w", err)
	}

	return token.AccessToken, nil
}
