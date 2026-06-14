Output assets/demo.gif

# Regenerate: sed "s|REPO_ROOT|$(pwd)|g" assets/demo.tape | vhs -

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 520
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

Hide
Type "cd REPO_ROOT"
Enter
Type "go build -o bin/rippl ./cmd/rippl"
Enter
Sleep 2s
Show

Type "./bin/rippl analyze internal/config/module.go --format text --no-color"
Enter
Sleep 4s

Type "./bin/rippl score internal/config/module.go"
Enter
Sleep 3s

Sleep 1s
