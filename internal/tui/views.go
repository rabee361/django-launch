package tui

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	var b strings.Builder

	content, err := os.ReadFile("ascii_art.txt")
	if err != nil {
		content = []byte("  DJANGO LAUNCH")
	}

	// Header Banner
	b.WriteString(titleStyle.Render(string(content)))
	b.WriteString("\n\n")

	switch m.step {
	case StepProjectName:
		b.WriteString(m.viewProjectName())
	case StepDocker:
		b.WriteString(m.viewDocker())
	case StepDependencies:
		b.WriteString(m.viewDependencies())
	case StepConfirm:
		b.WriteString(m.viewConfirm())
	case StepExecuting:
		b.WriteString(m.viewExecuting())
	case StepDone:
		b.WriteString(m.viewDone())
	case StepError:
		b.WriteString(m.viewError())
	}

	v := tea.NewView(b.String())

	return v
}

func (m Model) viewProjectName() string {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 1/4"))
	b.WriteString(stepHeaderStyle.Render("Project Name") + "\n\n")
	b.WriteString("Enter the name for your Django project:\n\n")
	b.WriteString(m.textInput.View() + "\n\n")

	if m.nameError != "" {
		b.WriteString(errorAlertStyle.Render("✗ "+m.nameError) + "\n\n")
	}

	b.WriteString(helpStyle.Render("Enter: continue • Ctrl+C: quit"))
	return b.String()
}

func (m Model) viewDocker() string {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 2/4"))
	b.WriteString(stepHeaderStyle.Render("Docker Support") + "\n\n")
	b.WriteString("Would you like to generate a Dockerfile and .dockerignore for this project?\n\n")

	options := []struct {
		label string
		desc  string
	}{
		{"Yes", "Generate a clean Dockerfile and .dockerignore"},
		{"No", "Skip Docker configuration"},
	}

	for i, opt := range options {
		cursor := "  "
		radio := "( )"
		style := unselectedItemStyle

		if m.dockerChoice == i {
			cursor = selectedItemStyle.Render("▸ ")
			radio = checkedStyle.Render("(•)")
			style = selectedItemStyle
		}

		b.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, radio, style.Render(opt.label), subtitleStyle.Render(opt.desc)))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ or y/n: select • Enter: continue • Esc: back"))
	return b.String()
}

func (m Model) viewDependencies() string {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 3/4"))
	b.WriteString(stepHeaderStyle.Render("Dependencies") + "\n\n")
	b.WriteString("Select optional packages to install and configure alongside Django:\n\n")

	for i, opt := range m.depOptions {
		cursor := "  "
		checkbox := uncheckedStyle.Render("[ ]")
		style := unselectedItemStyle

		if m.depChecked[opt.ID] {
			checkbox = checkedStyle.Render("[✓]")
		}

		if m.depCursor == i {
			cursor = selectedItemStyle.Render("▸ ")
			style = selectedItemStyle
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, style.Render(opt.Name)))
		b.WriteString(fmt.Sprintf("      %s\n", subtitleStyle.Render(opt.Description)))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • Space: toggle • 'a': toggle all • Enter: continue • Esc: back"))
	return b.String()
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 4/4"))
	b.WriteString(stepHeaderStyle.Render("Confirmation") + "\n\n")

	dockerText := "No"
	if m.dockerChoice == 0 {
		dockerText = "Yes (Dockerfile & .dockerignore)"
	}

	var selectedDeps []string
	for _, opt := range m.depOptions {
		if m.depChecked[opt.ID] {
			selectedDeps = append(selectedDeps, opt.Name)
		}
	}
	depsText := "None (Core Django only)"
	if len(selectedDeps) > 0 {
		depsText = strings.Join(selectedDeps, ", ")
	}

	toolText := "python -m venv & pip"
	if m.hasUv {
		toolText = "uv (ultra-fast virtualenv & pip)"
	}

	cardContent := fmt.Sprintf(
		"Project Name : %s\nTarget Path  : ./%s\nDocker       : %s\nDependencies : Django, %s\nEnvironment  : %s",
		selectedItemStyle.Render(m.cfg.ProjectName),
		m.cfg.ProjectName,
		dockerText,
		depsText,
		secondaryColorStyle(toolText),
	)

	b.WriteString(cardStyle.Render(cardContent) + "\n\n")
	b.WriteString(focusedPromptStyle.Render("Ready to generate your project? Press [Enter] to proceed, [Esc] to go back."))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter: start • Esc: back • q: cancel"))
	return b.String()
}

func (m Model) viewExecuting() string {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Generating"))
	b.WriteString(stepHeaderStyle.Render("Building Project...") + "\n\n")

	for _, item := range m.progressItems {
		b.WriteString(fmt.Sprintf("%s %s\n", checkedStyle.Render("✓"), item))
	}

	if m.currentTask != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), m.currentTask))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Please wait while your project environment and files are configured..."))
	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder

	b.WriteString(successStyle.Render("✨ Project successfully created! ✨") + "\n\n")

	var activateCmd string
	if runtime.GOOS == "windows" {
		activateCmd = fmt.Sprintf(".\\%s\\.venv\\Scripts\\activate", m.cfg.ProjectName)
	} else {
		activateCmd = fmt.Sprintf("source %s/.venv/bin/activate", m.cfg.ProjectName)
	}

	instructions := fmt.Sprintf(
		"Next Steps:\n\n"+
			"  1. Enter your project directory:\n"+
			"     cd %s\n\n"+
			"  2. Activate virtual environment:\n"+
			"     %s\n\n"+
			"  3. Apply initial migrations:\n"+
			"     python manage.py migrate\n\n"+
			"  4. Start development server:\n"+
			"     python manage.py runserver\n",
		m.cfg.ProjectName,
		activateCmd,
	)

	b.WriteString(cardStyle.Render(instructions) + "\n\n")
	b.WriteString(helpStyle.Render("Press Enter or 'q' to exit."))
	return b.String()
}

func (m Model) viewError() string {
	var b strings.Builder

	b.WriteString(errorAlertStyle.Render("✗ Generation Failed") + "\n\n")
	errMsg := "An unexpected error occurred."
	if m.err != nil {
		errMsg = m.err.Error()
	}

	b.WriteString(cardStyle.BorderForeground(errorColor).Render(errMsg) + "\n\n")
	b.WriteString(helpStyle.Render("Press Enter or 'q' to exit."))
	return b.String()
}

func secondaryColorStyle(s string) string {
	return lipgloss.NewStyle().Foreground(secondaryColor).Render(s)
}

func filepathRel(path string) string {
	rel, err := filepath.Rel(".", path)
	if err != nil {
		return path
	}
	return rel
}
