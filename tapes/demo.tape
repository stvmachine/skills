Output demo.gif

Set Shell "bash"
Set FontSize 14
Set FontFamily "FiraCode Nerd Font Mono"
Set Width 1200
Set Height 600
Set Padding 24
Set Framerate 24
Set TypingSpeed 55ms
Set Theme "Dracula"

Sleep 500ms

Type "stevmachine-skills list"
Sleep 400ms
Enter
Sleep 3s

Sleep 800ms

Type "stevmachine-skills doctor"
Sleep 400ms
Enter
Sleep 1500ms

Sleep 800ms

Type "stevmachine-skills env list"
Sleep 400ms
Enter
Sleep 1500ms

Sleep 800ms

Type "# launch Claude with decrypted credentials injected"
Sleep 500ms
Enter
Sleep 200ms
Type "dotenvx run -f ~/.stevmachine-skills/.env -- claude"
Sleep 3s
