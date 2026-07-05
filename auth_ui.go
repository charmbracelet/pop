package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
	"github.com/pkg/browser"
)

// authState represents the state of the OAuth TUI flow.
type authState int

const (
	authStateIntro authState = iota
	authStatePreparing
	authStateWaiting
	authStateExchanging
	authStateError
)

// authKeyMap represents the key bindings for the OAuth TUI.
type authKeyMap struct {
	Continue key.Binding
	Cancel   key.Binding
}

// defaultAuthKeybinds returns the default key bindings for the OAuth TUI.
func defaultAuthKeybinds() authKeyMap {
	return authKeyMap{
		Continue: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "continue"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c", "cancel"),
		),
	}
}

// ShortHelp returns the key bindings for the short help screen.
func (k authKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Continue, k.Cancel}
}

// FullHelp returns the key bindings for the full help screen.
func (k authKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Continue, k.Cancel}}
}

// updateAuthKeymap enables/disables key bindings based on the current state.
func (m *authModel) updateAuthKeymap() {
	m.keymap.Continue.SetEnabled(m.state == authStateIntro)
}

// authModel is the Bubble Tea model for the OAuth authorization flow.
type authModel struct {
	state authState

	spinner spinner.Model

	clientID      string
	codeVerifier  string
	oauthState    string
	redirectURI   string
	authURL       string
	browserFailed bool

	resultCh chan authCallbackResult
	server   *http.Server

	help   help.Model
	keymap authKeyMap

	authCode string
	err      error
	canceled bool
	quitting bool
	width    int
}

// authCallbackResult is pushed onto the result channel by the HTTP callback
// handler.
type authCallbackResult struct {
	code string
	err  error
}

type authReadyMsg struct {
	clientID     string
	codeVerifier string
	oauthState   string
	redirectURI  string
	authURL      string
	resultCh     chan authCallbackResult
	server       *http.Server
}

type authCallbackMsg struct {
	code string
	err  error
}

type authTokenMsg struct {
	err error
}

type authErrMsg struct {
	err error
}

func newAuthModel(noBrowser bool) authModel {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(charmtone.Julep)
	s.Spinner = spinner.Dot
	m := authModel{
		state:         authStateIntro,
		spinner:       s,
		help:          help.New(),
		keymap:        defaultAuthKeybinds(),
		browserFailed: noBrowser,
	}
	m.updateAuthKeymap()
	return m
}

// startOAuthFlowTUI runs the OAuth authorization flow with a small TUI that
// guides the user through opening a browser and waiting for the callback.
func startOAuthFlowTUI(noBrowser bool) error {
	p := tea.NewProgram(newAuthModel(noBrowser))
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("running auth program: %w", err)
	}
	m, ok := final.(authModel)
	if !ok {
		return errors.New("unexpected auth program model")
	}
	if m.err != nil {
		return m.err
	}
	if m.canceled {
		fmt.Println(authCanceledView())
		return nil
	}
	fmt.Println(authSuccessView())
	return nil
}

func authSuccessView() string {
	return "\n  " + activeLabelStyle.Render("Success!") + " You're authenticated with Resend.\n"
}

func authCanceledView() string {
	return "\n  Authentication canceled.\n"
}

// Init initializes the auth TUI.
func (m authModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the auth TUI.
func (m authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case authReadyMsg:
		m.clientID = msg.clientID
		m.codeVerifier = msg.codeVerifier
		m.oauthState = msg.oauthState
		m.redirectURI = msg.redirectURI
		m.authURL = msg.authURL
		m.resultCh = msg.resultCh
		m.server = msg.server
		if !m.browserFailed {
			m.browserFailed = browser.OpenURL(m.authURL) != nil
		}
		m.state = authStateWaiting
		m.updateAuthKeymap()
		return m, tea.Batch(waitForCallbackCmd(m.resultCh), m.spinner.Tick)

	case authCallbackMsg:
		if msg.err != nil {
			m.shutdownServer()
			m.err = msg.err
			m.state = authStateError
			m.updateAuthKeymap()
			m.quitting = true
			return m, tea.Quit
		}
		m.shutdownServer()
		m.authCode = msg.code
		m.state = authStateExchanging
		m.updateAuthKeymap()
		return m, tea.Batch(
			exchangeTokenCmd(m.clientID, m.authCode, m.redirectURI, m.codeVerifier),
			m.spinner.Tick,
		)

	case authTokenMsg:
		m.shutdownServer()
		if msg.err != nil {
			m.err = msg.err
			m.state = authStateError
			m.updateAuthKeymap()
			m.quitting = true
			return m, tea.Quit
		}
		m.quitting = true
		return m, tea.Quit

	case authErrMsg:
		m.shutdownServer()
		m.err = msg.err
		m.state = authStateError
		m.updateAuthKeymap()
		m.quitting = true
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		switch m.state {
		case authStatePreparing, authStateWaiting, authStateExchanging:
			return m, cmd
		default:
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.help.SetWidth(msg.Width)
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.Cancel):
			m.shutdownServer()
			m.canceled = true
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.Continue):
			if m.state == authStateIntro {
				m.state = authStatePreparing
				m.updateAuthKeymap()
				return m, tea.Batch(setupOAuthCmd(), m.spinner.Tick)
			}
		}
		// Update help model for full-help toggle.
		var cmd tea.Cmd
		m.help, cmd = m.help.Update(msg)
		return m, cmd
	}

	return m, nil
}

func authHeader() string {
	return "\n  " + noticeHeaderStyle.SetString("Resend").String() + " Let's Auth"
}

// authWidth returns the usable text width for the auth TUI, accounting for
// the 2-space indent. Falls back to 80 when the terminal width is unknown.
func (m authModel) authWidth() int {
	const indent = 2 //nolint:mnd
	w := m.width - indent
	if w < 20 { //nolint:mnd
		w = 80 - indent
	}
	return w
}

// View renders the auth TUI.
func (m authModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var content string
	wrap := lipgloss.NewStyle().MaxWidth(m.authWidth())
	switch m.state {
	case authStateIntro:
		content = authHeader() + "\n\n  " + wrap.Render("To authenticate we’re going to open the browser. Ready?")
	case authStatePreparing:
		content = authHeader() + "\n\n  " + m.spinner.View() + wrap.Render("Preparing...")
	case authStateWaiting:
		content = authHeader() + "\n\n  " + m.spinner.View() + wrap.Render("Waiting for authorization...")
		if m.browserFailed {
			urlStyle := lipgloss.NewStyle().
				Foreground(charmtone.Guac).
				Underline(true).
				Hyperlink(m.authURL).
				MaxWidth(m.authWidth())
			content += "\n\n  " + wrap.Render("Visit the following URL to authenticate:") + "\n  " + urlStyle.Render(m.authURL)
		}
	case authStateExchanging:
		content = authHeader() + "\n\n  " + m.spinner.View() + wrap.Render("Exchanging token...")
	case authStateError:
		content = authHeader() + "\n\n  " + errorStyle.Render(wrap.Render(m.err.Error()))
	default:
		return tea.NewView("")
	}

	content += "\n\n  " + m.help.View(m.keymap)
	return tea.NewView(content)
}

func (m authModel) shutdownServer() {
	if m.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = m.server.Shutdown(ctx)
}

// setupOAuthCmd performs dynamic client registration, PKCE generation, and
// starts the callback HTTP server, returning everything needed to continue.
func setupOAuthCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		const localhost = "127.0.0.1"
		lc := net.ListenConfig{}
		listener, err := lc.Listen(ctx, "tcp", localhost+":0")
		if err != nil {
			return authErrMsg{err: fmt.Errorf("starting callback server: %w", err)}
		}
		port := listener.Addr().(*net.TCPAddr).Port
		redirectURI := fmt.Sprintf("http://%s:%d%s", localhost, port, oauthRedirectPath)

		codeVerifier, err := generateCodeVerifier()
		if err != nil {
			return authErrMsg{err: err}
		}
		codeChallenge := generateCodeChallenge(codeVerifier)
		state, err := generateState()
		if err != nil {
			return authErrMsg{err: err}
		}

		clientID, err := registerClient(ctx, "http://"+localhost+oauthRedirectPath)
		if err != nil {
			return authErrMsg{err: err}
		}

		authURL, err := url.Parse(resendAPIBase + "/oauth/authorize")
		if err != nil {
			return authErrMsg{err: fmt.Errorf("parsing authorization URL: %w", err)}
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

		resultCh := make(chan authCallbackResult, 1)
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
					resultCh <- authCallbackResult{err: fmt.Errorf("OAuth error: %s", errVal)}
					return
				}
				code := query.Get("code")
				stateVal := query.Get("state")
				if code == "" {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("Missing authorization code"))
					resultCh <- authCallbackResult{err: errors.New("missing authorization code")}
					return
				}
				if stateVal != state {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte("State mismatch"))
					resultCh <- authCallbackResult{err: errors.New("state mismatch")}
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("Authorization successful! You can close this tab."))
				resultCh <- authCallbackResult{code: code}
			}),
		}
		go func() { _ = server.Serve(listener) }()

		return authReadyMsg{
			clientID:     clientID,
			codeVerifier: codeVerifier,
			oauthState:   state,
			redirectURI:  redirectURI,
			authURL:      authURL.String(),
			resultCh:     resultCh,
			server:       server,
		}
	}
}

// waitForCallbackCmd blocks until the OAuth callback fires or the flow times
// out.
func waitForCallbackCmd(resultCh chan authCallbackResult) tea.Cmd {
	return func() tea.Msg {
		select {
		case res := <-resultCh:
			return authCallbackMsg(res)
		case <-time.After(5 * time.Minute):
			return authCallbackMsg{err: errors.New("authorization timed out")}
		}
	}
}

// exchangeTokenCmd exchanges the authorization code for tokens and persists
// them to disk.
func exchangeTokenCmd(clientID, code, redirectURI, codeVerifier string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		tokResp, err := exchangeCode(ctx, clientID, code, redirectURI, codeVerifier)
		if err != nil {
			return authTokenMsg{err: err}
		}
		token := &OAuthToken{
			ClientID:     clientID,
			AccessToken:  tokResp.AccessToken,
			RefreshToken: tokResp.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(tokResp.ExpiresIn) * time.Second),
		}
		if err := saveAuth(token); err != nil {
			return authTokenMsg{err: fmt.Errorf("saving auth: %w", err)}
		}
		return authTokenMsg{}
	}
}
