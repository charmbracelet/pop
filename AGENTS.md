# AGENTS.md

Guide for AI agents working on the Pop codebase.

## What is Pop?

Pop is a command-line email client built in Go. It supports two delivery methods
— Resend (OAuth or API key) and SMTP — and provides both an interactive TUI
(Bubble Tea) and non-interactive CLI sending.

## Build & Development Commands

```bash
task build       # Build the project
task lint        # Run golangci-lint (strict config in .golangci.yml)
task lint:fix    # Auto-fix lint issues
task vet         # Run go vet
task test        # Run tests
task fmt         # Format code (gofumpt + goimports)
task tidy        # Tidy go modules
task             # Run the project (go run . {{.CLI_ARGS}})
```

Always run `task lint` before committing. The linter enables `gofumpt` and
`goimports` formatters, `wrapcheck` (errors from external packages must be
wrapped with `fmt.Errorf("...: %w", err)`), `goconst` (repeated string literals
should be constants), `gosec`, `revive`, and others.

## Architecture

### Source Layout

All source files are in the package root (no sub-packages):

- `main.go` — Root cobra command, flag definitions, delivery method detection,
  non-interactive send path
- `model.go` — Bubble Tea TUI model (`Model`), state machine, view rendering
- `keymap.go` — TUI key bindings and help
- `email.go` — Email sending logic (SMTP via go-simple-mail, Resend via
  resend-go), Markdown-to-HTML conversion (goldmark)
- `attachments.go` — Attachment list item type and delegate for the TUI
- `auth.go` — OAuth 2.0 with PKCE: token storage, client registration, token
  exchange, refresh, callback server
- `auth_cli.go` — Non-interactive OAuth flow (stdin not a terminal)
- `auth_ui.go` — Interactive OAuth flow (Bubble Tea TUI with spinner states)
- `style.go` — Lipgloss styles and charmtone color palette
- `skill.go` — `pop skill` command (prints skill definition)
- `skill.md` — Embedded skill body (via `go:embed`)
- `install_skill.go` — `pop install-skill` command with per-agent subcommands

### CLI vs TUI

The root command's `RunE` determines whether to send non-interactively or launch
the TUI:

- **Non-interactive**: All of `--to`, `--from`, `--subject`, and a body are
  provided AND `--preview` is not set. Sends immediately.
- **TUI**: Launched when any required field is missing or stdin is a terminal
  with no piped body.

### Delivery Methods

Pop auto-detects the delivery method from environment and flags in `main.go`:

- `Resend` — via `RESEND_API_KEY` env, `--resend.key` flag, or OAuth token from
  `pop auth`
- `SMTP` — via `POP_SMTP_*` environment variables or `--smtp.*` flags
- Setting both is an error (`Unknown` delivery method)

### OAuth Flow

OAuth uses PKCE with Resend's OAuth 2.0 API:

1. Dynamic client registration (`POST /oauth/register`)
2. Start local callback server on `127.0.0.1`
3. Open browser for authorization
4. Exchange code for tokens
5. Persist tokens to `auth.json` in the user config dir

Two entry points:

- **TUI** (`auth_ui.go`): `startOAuthFlowTUI()` — interactive Bubble Tea program
  with spinner states (intro → preparing → waiting → exchanging → done/error)
- **CLI** (`auth_cli.go`): `startOAuthFlow()` — plain stdout, used when stdin is
  not a terminal

Token refresh happens automatically in `getValidAccessToken()` before each send
if the token is within 5 minutes of expiry.

### Colors and Styling

Pop uses the **charmtone** color palette
(`github.com/charmbracelet/x/exp/charmtone`) for all colors, not raw hex values.
See `style.go` for the full set of named colors (Charple, Zest, Soda, Steam,
etc.).

**All CLI-based styled output (not Bubble Tea views) must go through a
`colorprofile` writer** to handle terminal color downsampling correctly:

```go
w := colorprofile.NewWriter(os.Stderr, os.Environ())
_, _ = fmt.Fprintf(w, "...")
```

Bubble Tea applies colorprofile automatically to its rendered output, so TUI
views don't need this. This is only for direct `fmt.Fprintf`/`fmt.Println`
calls to stdout/stderr outside of a Bubble Tea program. See `main.go` and
`install_skill.go` for examples.

Error messages follow a consistent pattern:

```
\n  ERROR  <message>\n\n
```

Use `errorHeaderStyle` for the badge and `inlineCodeStyle` for code/paths. Set
`cmd.SilenceUsage = true` and `cmd.SilenceErrors = true` in cobra `RunE` when
printing custom errors.

### Skills

Pop can generate and install skill definitions for AI agents:

- `pop skill` — Prints the skill to stdout (Crush-compatible YAML frontmatter +
  markdown body)
- `pop install-skill <target>` — Installs to a specific agent's rules directory

The skill body lives in `skill.md` and is embedded at build time via
`//go:embed`. Supported targets: `crush`, `claude`, `codex`, `cursor`. Each
target writes to the correct path and format for that agent.

## Conventions

- Use `gofumpt` formatting (stricter than `gofmt`)
- Wrap all errors from external packages with `fmt.Errorf("context: %w", err)`
  (wrapcheck)
- Extract repeated string literals to constants (goconst)
- Use `charmtone` colors, never raw hex
- Use `colorprofile.NewWriter` for all styled terminal output
- Check `_, _ =` on `fmt.Fprintf` to errWriter (errcheck)
- File permissions: `0o750` for directories, `0o600` for files (gosec)
- Use `//nolint:gosec // reason` or `//nolint:mnd` only where truly needed
