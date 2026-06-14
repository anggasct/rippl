Output assets/demo.gif

# Regenerate: sed "s|REPO_ROOT|$(pwd)|g" assets/demo.tape | vhs -

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms

Hide
Type "cd REPO_ROOT"
Enter
Type "mkdir -p bin && go build -o bin/rippl ./cmd/rippl"
Enter
Sleep 3s
Show

Type "./bin/rippl analyze internal/parser/analysis.go"
Enter
Sleep 3s

# Navigate the affected-files list
Down 4
Sleep 1s

# Open detail panel
Type "d"
Sleep 2s

# Browse files from detail view
Down 2
Sleep 1s

# Scroll the viewport
PageDown
Sleep 1s

# Return to list and quit
Escape
Sleep 500ms
Type "q"
Sleep 500ms
