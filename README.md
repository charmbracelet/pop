# Pop

Send emails from your terminal.

## Command Line Interface

```bash
pop < message.md \
    --from "me@example.com" \
    --to "you@example.com" \
    --subject "Hello, world!" \
    --attach invoice.pdf
```

## Text-based User Interface

Launch the TUI

```bash
pop
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

[MIT](https://github.com/charmbracelet/pop/blob/master/LICENSE)

## Feedback

We’d love to hear your thoughts on this project. Feel free to drop us a note!

* [Twitter](https://twitter.com/charmcli)
* [The Fediverse](https://mastodon.social/@charmcli)
* [Discord](https://charm.sh/chat)

## License

[MIT](https://github.com/charmbracelet/vhs/raw/main/LICENSE)

***

Part of [Charm](https://charm.sh).

<a href="https://charm.sh/">
  <img
    alt="The Charm logo"
    width="400"
    src="https://stuff.charm.sh/charm-badge.jpg"
  />
</a>

Charm热爱开源 • Charm loves open source
