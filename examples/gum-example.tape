Output gum-example.gif
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


Type `pop --from "$(gum choose vt{52,78,100}@charm.sh)" \` Enter
Type `    --to "$(gum filter < contacts.txt)"` Enter

Sleep 1s

Down@200ms 2
Sleep 1s
Enter
Sleep 1s

Type "pop"
Sleep 1s
Enter

Sleep 5s
