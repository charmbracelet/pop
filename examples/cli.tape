Output cli.gif
Set Framerate 30
Set Padding 70
Set FontSize 38
Set Width 1200
Set Height 800

Hide
  Type "export RESEND_API_KEY=$(pass RESEND_CHARM_API_KEY)"
  Enter
  Sleep 1
  Ctrl+L
Show

Type "pop < message.md \" Enter
Type "  --from 'pop@charm.sh' \" Enter
Type "  --to 'vt100@charm.sh' \" Enter
Type "  --subject 'Pop' \" Enter
Type "  --attach README.md"
Sleep 0.5
Enter
Sleep 5
