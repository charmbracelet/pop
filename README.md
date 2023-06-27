# Pop

<p>
  <img src="https://stuff.charm.sh/pop/pop.png" width="400" />
  <br />
  <a href="https://github.com/charmbracelet/pop/releases"><img src="https://img.shields.io/github/release/charmbracelet/pop.svg" alt="Latest Release"></a>
  <a href="https://pkg.go.dev/github.com/charmbracelet/pop?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="Go Docs"></a>
  <a href="https://github.com/charmbracelet/pop/actions"><img src="https://github.com/charmbracelet/vhs/workflows/build/badge.svg" alt="Build Status"></a>
</p>

Send emails from your terminal.

<img width="500" src="https://vhs.charm.sh/vhs-25OHmk90ODL9BefXZ1P99I.gif" alt="pop mail text-based client">

## Text-based User Interface

Launch the TUI

```bash
pop
```

## Command Line Interface

```bash
pop < message.md \
    --from "me@example.com" \
    --to "you@example.com" \
    --subject "Hello, world!" \
    --attach invoice.pdf
```

<img width="500" src="https://vhs.charm.sh/vhs-28eAQHRSZilAiXKqTccLCr.gif" alt="pop mail command line client">

---

<img width="600" src="https://stuff.charm.sh/pop/resend-x-charm.png" alt="Resend and Charm logos">

To use `pop`, you will need a `RESEND_API_KEY`.

You can grab one from: https://resend.com/api-keys.

### Environment

To avoid typing your `From: ` email address, you can also set the `RESEND_FROM`
environment to pre-fill the field anytime you launch `pop`.

```bash
export RESEND_API_KEY=$(pass RESEND_API_KEY)
export RESEND_FROM=pop@charm.sh
```

## Installation

Use a package manager:

```bash
# macOS
brew install pop

# Arch
yay -S pop

# Nix
nix-env -iA nixpkgs.pop
```

Install with Go:

```sh
go install github.com/charmbracelet/pop@latest
```

Or download a binary from the [releases](https://github.com/charmbracelet/pop/releases).

## License

[MIT](https://github.com/charmbracelet/pop/blob/main/LICENSE)

## Feedback

We’d love to hear your thoughts on this project. Feel free to drop us a note!

- [Twitter](https://twitter.com/charmcli)
- [The Fediverse](https://mastodon.social/@charmcli)
- [Discord](https://charm.sh/chat)

## License

[MIT](https://github.com/charmbracelet/pop/raw/main/LICENSE)

---

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/">
  <img
    alt="The Charm logo"
    width="400"
    src="https://stuff.charm.sh/charm-badge.jpg"
  />
</a>

Charm 热爱开源 • Charm loves open source
