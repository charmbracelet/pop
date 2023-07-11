Output mods-example.gif
Set Framerate 30
Set FontSize 28
Set Padding 40
Set Width 1200
Set Height 800

Hide
  Type "export RESEND_API_KEY=$(pass RESEND_CHARM_API_KEY)"
  Enter
  Sleep 1
  Type "export OPENAI_API_KEY=$(pass OPENAI_API_KEY)"
  Enter
  Sleep 1
  Ctrl+L
Show


Type `pop <<< "$(mods -f 'Explain why CLIs are awesome')" \` Enter
Type `    --subject "The command line is the best" \` Enter
Type `    --preview` Enter

Sleep 5s

Hide
Sleep 60s
Show

Sleep 5s
