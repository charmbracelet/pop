# Pop — Send Email from the Terminal

Pop is a command-line tool for sending email. This skill covers non-interactive
usage.

## Setup

Pop supports three delivery methods. Configure exactly one.

### Resend OAuth (recommended)

    pop auth

Authenticates with Resend via OAuth 2.0 in the browser. Tokens are stored
locally and reused automatically on subsequent sends, so this only needs to be
run once.

Revoke with:

    pop auth revoke

### Resend API Key

Set RESEND_API_KEY in the environment, or pass --resend.key.

    export RESEND_API_KEY=re_xxxxxxxx

### SMTP

    export POP_SMTP_HOST=smtp.gmail.com
    export POP_SMTP_PORT=587
    export POP_SMTP_USERNAME=you@example.com
    export POP_SMTP_PASSWORD=secret
    export POP_SMTP_ENCRYPTION=starttls

Encryption options: starttls (default), ssl, none.

### Other Environment Variables

    POP_FROM          Default sender address
    POP_SIGNATURE     Signature appended to the email body
    POP_PLAINTEXT     Set to "true" to send plain text instead of rendered HTML
    POP_UNSAFE_HTML   Set to "true" to allow unsafe HTML and extra markdown features

## Sending Email (Non-Interactive)

When --to, --from, --subject, and a body are ALL provided and --preview is NOT
set, Pop sends immediately without launching the TUI.

Body via stdin:

    pop < message.md --from me@example.com --to you@example.com --subject "Hello"

Body via flag:

    pop --body "Hello there" --from me@example.com --to you@example.com --subject "Hello"

### Flags

    -f, --from         Sender address (env POP_FROM)
    -t, --to           Recipients (comma-separated or repeatable)
        --cc           CC recipients
        --bcc          BCC recipients
    -s, --subject      Email subject
    -b, --body         Email body (Markdown, rendered to HTML)
    -a, --attach       Attach a file (repeatable)
    -x, --signature    Signature appended to body (env POP_SIGNATURE)
    -u, --unsafe       Allow unsafe HTML / extra markdown features (env POP_UNSAFE_HTML)
        --plaintext    Send plain text instead of rendering Markdown to HTML
        --preview      Open the TUI to review before sending

### Attachments

    pop --attach invoice.pdf --attach report.docx \
        --from me@example.com --to client@example.com \
        --subject "Documents" --body "See attached."

## Composing with Other Tools

Pipe generated content from another CLI tool into pop:

    pop <<< "$(crush run 'Write a welcome email')" \
        --subject "Welcome" --from me@example.com --to you@example.com

## Notes

- The body is Markdown, rendered to HTML before sending.
- If both Resend and SMTP are configured, Pop errors — set only one.
- Gmail users: host/port default automatically when the username ends in
  @gmail.com.
- If --preview is set or any required field is missing, the interactive TUI
  launches instead.
