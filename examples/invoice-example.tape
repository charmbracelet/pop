Output invoice-example.gif
Set Framerate 30
Set FontSize 28
Set Padding 40
Set Width 1200
Set Height 800

Hide
  Type "export RESEND_API_KEY=$(pass RESEND_CHARM_API_KEY)"
  Enter
  Sleep 1
  Ctrl+L
Show


Type `invoice generate -i "Rubber Ducky" -r 25 -q 2 -o invoice.pdf` Enter
Sleep 1s
Type `pop --attach invoice.pdf --body "See attached invoice."` Enter
Sleep 5s
