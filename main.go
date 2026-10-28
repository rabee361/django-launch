package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/rabee361/django-launch/internal/tui"
)

func main() {
	m, err := tea.NewProgram(tui.NewModel()).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
	if steps := m.(tui.Model).NextSteps(); steps != "" {
		fmt.Println("\n" + steps)
	}
}
