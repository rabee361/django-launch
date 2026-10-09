package tui

import (
	"fmt"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	var str_builder strings.Builder

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
	str_builder.WriteString(titleStyle.Render(string(bannerArt)))
	str_builder.WriteString("\n\n")

	base := strings.Count(str_builder.String(), "\n")
	nameRow, depRow, dbRow := -1, -1, -1

	switch m.step {
	case StepProjectName:
		s, row := m.viewProjectName()
		str_builder.WriteString(s)
		nameRow = base + row
	case StepDocker:
		str_builder.WriteString(m.viewDocker())
	case StepDependencies:
		s, row := m.viewDependencies()
		str_builder.WriteString(s)
		depRow = base + row
	case StepDataBase:
		s, row := m.viewDatabases()
		str_builder.WriteString(s)
		dbRow = base + row
	case StepConfirm:
		str_builder.WriteString(m.viewConfirm())
	case StepExecuting:
		str_builder.WriteString(m.viewExecuting())
	case StepDone:
		str_builder.WriteString(m.viewDone())
	case StepError:
		str_builder.WriteString(m.viewError())
	}

	frame := str_builder.String()
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
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 1/5"))
	str_builder.WriteString(stepHeaderStyle.Render("Project Name"))
	str_builder.WriteString("\n\n")
	str_builder.WriteString("Enter the name for your Django project:\n\n")

	row := strings.Count(str_builder.String(), "\n")
	str_builder.WriteString(m.textInput.View())
	str_builder.WriteString("\n\n")

	if m.nameError != "" {
		str_builder.WriteString(errorAlertStyle.Render("✗ " + m.nameError))
		str_builder.WriteString("\n\n")

	}

	str_builder.WriteString(helpStyle.Render("Type or paste • Enter: continue • Ctrl+C: quit"))
	return str_builder.String(), row
}

func (m Model) viewDocker() string {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 2/5"))
	str_builder.WriteString(stepHeaderStyle.Render("Docker Support"))
	str_builder.WriteString("\n\n")

	str_builder.WriteString("Would you like to generate a Dockerfile and .dockerignore for this project?\n\n")

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

		str_builder.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, radio, style.Render(opt.label), subtitleStyle.Render(opt.desc)))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓/←/→ or hjkl: choose • y/n: choose and continue • Enter: continue • Esc/Backspace: back"))
	return str_builder.String()
}

func (m Model) viewDependencies() (string, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 3/5"))
	str_builder.WriteString(stepHeaderStyle.Render("Dependencies"))
	str_builder.WriteString("\n\n")
	str_builder.WriteString("Select optional packages to install and configure alongside Django:\n\n")

	row := -1

	for i, opt := range m.depOptions {
		if i == m.depCursor {
			row = strings.Count(str_builder.String(), "\n")
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

		str_builder.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, style.Render(opt.Name)))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓ or k/j: navigate • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back"))
	return str_builder.String(), row
}

func (m Model) viewDatabases() (string, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 4/5"))
	str_builder.WriteString(stepHeaderStyle.Render("Databases"))
	str_builder.WriteString("\n\n")
	str_builder.WriteString("Select the Database you want to configure for your project:\n\n")

	row := -1

	for i, opt := range m.dbOptions {
		if i == m.depCursor {
			row = strings.Count(str_builder.String(), "\n")
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

		str_builder.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, style.Render(opt.Name)))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓ or k/j: navigate • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back"))
	return str_builder.String(), row
}

func (m Model) viewConfirm() string {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 5/5"))
	str_builder.WriteString(stepHeaderStyle.Render("Confirmation"))
	str_builder.WriteString("\n\n")

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

	str_builder.WriteString(cardStyle.MaxWidth(m.width).Render(cardContent))
	str_builder.WriteString("\n\n")
	str_builder.WriteString(focusedPromptStyle.Render("Ready to generate your project? Press [Enter] to start, [Esc] to go back, [q] to cancel."))
	str_builder.WriteString("\n\n")
	str_builder.WriteString(helpStyle.Render("Enter/y: start • Esc/Backspace/b: back • q/n: cancel"))
	return str_builder.String()
}

func (m Model) viewExecuting() string {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Generating"))
	str_builder.WriteString(stepHeaderStyle.Render("Building Project..."))
	str_builder.WriteString("\n\n")

	for _, item := range m.progressItems {
		str_builder.WriteString(fmt.Sprintf("%s %s\n", checkedStyle.Render("✓"), item))
	}

	if m.currentTask != "" {
		task := m.currentTask
		if m.progressTotal > 0 {
			task = fmt.Sprintf("%s (%d/%d)", task, m.progressStep, m.progressTotal)
		}
		str_builder.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), task))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("q/n/esc: cancel • Please wait while your project environment and files are configured..."))
	return str_builder.String()
}

func (m Model) viewDone() string {
	var str_builder strings.Builder

	str_builder.WriteString(successStyle.Render("✨ Project successfully created! ✨"))
	str_builder.WriteString("\n\n")

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

	str_builder.WriteString(cardStyle.Render(instructions))
	str_builder.WriteString("\n\n")
	str_builder.WriteString(helpStyle.Render("Press Enter, q, or Esc to exit."))
	return str_builder.String()
}

func (m Model) viewError() string {
	var str_builder strings.Builder

	str_builder.WriteString(errorAlertStyle.Render("✗ Generation Failed"))
	str_builder.WriteString("\n\n")
	errMsg := "An unexpected error occurred."
	if m.err != nil {
		errMsg = m.err.Error()
	}

	str_builder.WriteString(cardStyle.BorderForeground(errorColor).Render(errMsg))
	str_builder.WriteString("\n\n")
	str_builder.WriteString(helpStyle.Render("Press Enter, q, or Esc to exit."))
	return str_builder.String()
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
