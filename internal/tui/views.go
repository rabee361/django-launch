package tui

import (
	"fmt"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	var b strings.Builder

	bannerArt := `
██████╗      ██╗ █████╗ ███╗   ██╗ ██████╗  ██████╗ 
██╔══██╗     ██║██╔══██╗████╗  ██║██╔════╝ ██╔═══██╗
██║  ██║     ██║███████║██╔██╗ ██║██║  ███╗██║   ██║
██║  ██║██   ██║██╔══██║██║╚██╗██║██║   ██║██║   ██║
██████╔╝╚█████╔╝██║  ██║██║ ╚████║╚██████╔╝╚██████╔╝
╚═════╝  ╚════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝  ╚═════╝ 
                                                    
██╗      █████╗ ██╗   ██╗███╗   ██╗ ██████╗██╗  ██╗ 
██║     ██╔══██╗██║   ██║████╗  ██║██╔════╝██║  ██║ 
██║     ███████║██║   ██║██╔██╗ ██║██║     ███████║ 
██║     ██╔══██║██║   ██║██║╚██╗██║██║     ██╔══██║ 
███████╗██║  ██║╚██████╔╝██║ ╚████║╚██████╗██║  ██║ 
╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝ 
`
	b.WriteString(titleStyle.Render(string(bannerArt)))
	b.WriteString("\n\n")

	base := strings.Count(b.String(), "\n")
	nameRow, depRow, dbRow := -1, -1, -1

	switch m.step {
	case StepProjectName:
		s, row := m.viewProjectName()
		b.WriteString(s)
		nameRow = base + row
	case StepDocker:
		b.WriteString(m.viewDocker())
	case StepDependencies:
		s, row := m.viewDependencies()
		b.WriteString(s)
		depRow = base + row
	case StepDataBase:
		s, row := m.viewDatabases()
		b.WriteString(s)
		dbRow = base + row	
	case StepConfirm:
		b.WriteString(m.viewConfirm())
	case StepExecuting:
		b.WriteString(m.viewExecuting())
	case StepDone:
		b.WriteString(m.viewDone())
	case StepError:
		b.WriteString(m.viewError())
	}

	frame := b.String()
	if m.width > 0 {
		frame = clipLines(frame, m.width)
	}

	v := tea.NewView(frame)
	switch m.step {
	case StepProjectName:
		if c := m.textInput.Cursor(); c != nil {
			c.Y = nameRow
			v.Cursor = c
		}
	case StepDependencies:
		if depRow >= 0 {
			c := tea.NewCursor(0, depRow)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	case StepDataBase:
		if dbRow >= 0 {
			c := tea.NewCursor(0, depRow)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	}
	return v
}

func (m Model) viewProjectName() (string, int) {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 1/5"))
	b.WriteString(stepHeaderStyle.Render("Project Name") + "\n\n")
	b.WriteString("Enter the name for your Django project:\n\n")

	row := strings.Count(b.String(), "\n")
	b.WriteString(m.textInput.View() + "\n\n")

	if m.nameError != "" {
		b.WriteString(errorAlertStyle.Render("✗ "+m.nameError) + "\n\n")
	}

	b.WriteString(helpStyle.Render("Type or paste • Enter: continue • Ctrl+C: quit"))
	return b.String(), row
}

func (m Model) viewDocker() string {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 2/5"))
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
	b.WriteString(helpStyle.Render("↑/↓/←/→ or hjkl: choose • y/n: choose and continue • Enter: continue • Esc/Backspace: back"))
	return b.String()
}

func (m Model) viewDependencies() (string, int) {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 3/5"))
	b.WriteString(stepHeaderStyle.Render("Dependencies") + "\n\n")
	b.WriteString("Select optional packages to install and configure alongside Django:\n\n")

	row := -1

	for i, opt := range m.depOptions {
		if i == m.depCursor {
			row = strings.Count(b.String(), "\n")
		}

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
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ or k/j: navigate • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back"))
	return b.String(), row
}

func (m Model) viewDatabases() (string, int) {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 4/5"))
	b.WriteString(stepHeaderStyle.Render("Databases") + "\n\n")
	b.WriteString("Select the Database you want to configure for your project:\n\n")

	row := -1

	for i, opt := range m.dbOptions {
		if i == m.depCursor {
			row = strings.Count(b.String(), "\n")
		}

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
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ or k/j: navigate • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back"))
	return b.String(), row
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(stepBadgeStyle.Render("Step 5/5"))
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

	b.WriteString(cardStyle.MaxWidth(m.width).Render(cardContent) + "\n\n")
	b.WriteString(focusedPromptStyle.Render("Ready to generate your project? Press [Enter] to start, [Esc] to go back, [q] to cancel."))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter/y: start • Esc/Backspace/b: back • q/n: cancel"))
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
		task := m.currentTask
		if m.progressTotal > 0 {
			task = fmt.Sprintf("%s (%d/%d)", task, m.progressStep, m.progressTotal)
		}
		b.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), task))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("q/n/esc: cancel • Please wait while your project environment and files are configured..."))
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
	b.WriteString(helpStyle.Render("Press Enter, q, or Esc to exit."))
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
	b.WriteString(helpStyle.Render("Press Enter, q, or Esc to exit."))
	return b.String()
}

func secondaryColorStyle(s string) string {
	return lipgloss.NewStyle().Foreground(secondaryColor).Render(s)
}

func clipLines(s string, width int) string {
	lines := strings.Split(s, "\n")
	style := lipgloss.NewStyle().MaxWidth(width)
	for i := range lines {
		lines[i] = style.Render(lines[i])
	}
	return strings.Join(lines, "\n")
}
