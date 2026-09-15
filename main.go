package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"main/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.NewModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}