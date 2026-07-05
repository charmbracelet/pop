package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pkg/browser"
)

// openBrowser opens the given URL in the user's default browser. If the
// browser can't be opened, the URL is printed for the user to open manually.
func openBrowser(rawURL string) {
	if err := browser.OpenURL(rawURL); err != nil {
		fmt.Printf("Please open this URL to authenticate:\n  %s\n", rawURL)
	}
}

// startOAuthFlow performs the full OAuth authorization flow without a TUI:
// DCR → start callback server → open browser → exchange code → save tokens.
//
// It is used when stdin is not a terminal (piped input, SSH sessions with port
// forwarding, or anywhere a Bubble Tea program is unsuitable). For interactive
// terminal sessions, auth_ui.go runs an equivalent flow with a TUI.
func startOAuthFlow() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	listener, redirectURI, err := openCallbackListener(ctx)
	if err != nil {
		return err
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return err
	}
	codeChallenge := generateCodeChallenge(codeVerifier)
	state, err := generateState()
	if err != nil {
		return err
	}

	clientID, err := registerClient(ctx, "http://"+oauthCallbackHost+oauthRedirectPath)
	if err != nil {
		return err
	}

	authURL, err := buildAuthURL(clientID, redirectURI, state, codeChallenge)
	if err != nil {
		return err
	}

	server, resultCh := newCallbackServer(state, listener)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	fmt.Printf("Opening browser for authorization...\n")
	openBrowser(authURL)
	fmt.Printf("Browser not opening? Pay a visit to:\n  %s\n", authURL)

	var authCode string
	select {
	case res := <-resultCh:
		if res.err != nil {
			return res.err
		}
		authCode = res.code
	case <-ctx.Done():
		return errors.New("authorization timed out")
	}

	tokResp, err := exchangeCode(ctx, clientID, authCode, redirectURI, codeVerifier)
	if err != nil {
		return err
	}

	token := &OAuthToken{
		ClientID:     clientID,
		AccessToken:  tokResp.AccessToken,
		RefreshToken: tokResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokResp.ExpiresIn) * time.Second),
	}
	if err := saveAuth(token); err != nil {
		return fmt.Errorf("saving auth: %w", err)
	}

	fmt.Println("Successfully authenticated with Resend!")
	return nil
}
