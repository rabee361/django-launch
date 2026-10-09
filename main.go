package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/rabee361/django-launch/internal/tui"
)

func main() {
	if _, err := tea.NewProgram(tui.NewModel()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
