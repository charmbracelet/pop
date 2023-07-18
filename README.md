# Pop

<p>
  <img src="https://stuff.charm.sh/pop/pop-logo.png" width="400" />
  <br />
  <a href="https://github.com/charmbracelet/pop/releases"><img src="https://img.shields.io/github/release/charmbracelet/pop.svg" alt="Latest Release"></a>
  <a href="https://pkg.go.dev/github.com/charmbracelet/pop?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="Go Docs"></a>
  <a href="https://github.com/charmbracelet/pop/actions"><img src="https://github.com/charmbracelet/vhs/workflows/build/badge.svg" alt="Build Status"></a>
</p>

Send emails from your terminal.

<img width="500" src="https://vhs.charm.sh/vhs-5Dyv3pvzB2YwtUSr72LqSz.gif" alt="pop mail text-based client">

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

<img width="500" src="https://vhs.charm.sh/vhs-5Cr6Gt1YVBjxGr9zdS85AO.gif" alt="pop mail command line client">

---

<img width="600" src="https://stuff.charm.sh/pop/resend-x-charm.png" alt="Resend and Charm logos">

To use `pop`, you will need a `RESEND_API_KEY` or configure an
[`SMTP`](#smtp-configuration) host.

You can grab one from: https://resend.com/api-keys.

### Resend Configuration

To use the resend delivery method, set the `RESEND_API_KEY` environment
variable.

```bash
export RESEND_API_KEY=$(pass RESEND_API_KEY)
```


### SMTP Configuration

To configure `pop` to use `SMTP`, you can set the following environment
variables.

```bash
export POP_SMTP_HOST=smtp.gmail.com
export POP_SMTP_PORT=587
export POP_SMTP_USERNAME=pop@charm.sh
export POP_SMTP_PASSWORD=hunter2
```

### Environment

To avoid typing your `From: ` email address, you can also set the `POP_FROM`
environment to pre-fill the field anytime you launch `pop`.

```bash
export POP_FROM=pop@charm.sh
export POP_SIGNATURE="Sent with [Pop](https://github.com/charmbracelet/pop)!"
```

## Installation

Use a package manager:

```bash
# macOS or Linux
brew install pop

# Nix
nix-env -iA nixpkgs.pop

# Arch (btw)
yay -S charm-pop-bin
```

Install with Go:

```sh
go install github.com/charmbracelet/pop@latest
```

Or download a binary from the [releases](https://github.com/charmbracelet/pop/releases).

## Examples

Pop can be combined with other tools to create powerful email pipelines, such as:

- [`charmbracelet/mods`](https://github.com/charmbracelet/mods)
- [`charmbracelet/gum`](https://github.com/charmbracelet/gum)
- [`maaslalani/invoice`](https://github.com/maaslalani/invoice)

### Mods

Use [`mods`](https://github.com/charmbracelet/mods) with `pop` to write an email body with AI:

> **Note**:
> Use the `--preview` flag to preview the email and make changes before sending.

```bash
pop <<< '$(mods -f "Explain why CLIs are awesome")' \
    --subject "The command line is the best" \
    --preview
```

<img width="600" src="https://vhs.charm.sh/vhs-1O3zo8Nsi2kPVW3vOBw4WH.gif" alt="Generate email with mods and send email with pop.">

- [`charmbracelet/mods`](https://github.com/charmbracelet/mods)

### Gum

Use [`gum`](https://github.com/charmbracelet/gum) with `pop` to choose an email to send to and from:

```bash
pop --from $(gum choose "vt52@charm.sh" "vt78@charm.sh" "vt100@charm.sh")
    --to $(gum filter < contacts.txt)
```

<img width="600" src="https://vhs.charm.sh/vhs-Et9ooHB6L1XVWDL9U1TfI.gif" alt="Select contact information with gum and send email with pop.">

- [`charmbracelet/gum`](https://github.com/charmbracelet/gum)

### Invoice

Use [`invoice`](https://github.com/maaslalani/invoice) with `pop` to generate and send invoices entirely from the command line.

```bash
FILENAME=invoice.pdf
invoice generate --item "Rubber Ducky" --rate 25 --quantity 2 --output $FILENAME
pop --attach $FILENAME --body "See attached invoice."
```

<img width="600" src="https://vhs.charm.sh/vhs-4TRyv82BBDKOutgWdvyshr.gif" alt="Generate invoice with invoice and attach file and send email with pop.">

- [`maaslalani/invoice`](https://github.com/maaslalani/invoice)

## Feedback

We’d love to hear your thoughts on this project. Feel free to drop us a note!

- [Twitter](https://twitter.com/charmcli)
- [The Fediverse](https://mastodon.social/@charmcli)
- [Discord](https://charm.sh/chat)

## License

[MIT](https://github.com/charmbracelet/pop/blob/main/LICENSE)

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
