Output demo.gif
Set Padding 40
Set FontSize 20
Set Width 600
Set Height 400

Hide
  Type "export RESEND_API_KEY=$(pass RESEND_API_KEY)"
  Enter
  Sleep 1
  Ctrl+L
Show

Type "pop < message.md \" Enter
Type "--from 'pop@lalani.dev' \" Enter
Type "--to 'maas@charm.sh' \" Enter
Type "--subject 'Pop' \" Enter
Type "--attach README.md"
Sleep 0.5
Enter
Sleep 5
