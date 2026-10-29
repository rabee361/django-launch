package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const bannerArt = `
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

var (
	bannerPrefix = titleStyle.Render(bannerArt) + "\n"
	bannerRows   = strings.Count(bannerPrefix, "\n")
)

// headerFor renders the full banner when it fits in the terminal height and a
// compact one-line title otherwise, so the alt-screen never clips the content.
func (m Model) headerFor(bodyRows int) string {
	if m.height == 0 || m.height >= bannerRows+bodyRows {
		return bannerPrefix
	}
	return compactTitleStyle.Render("DJANGO LAUNCH") + "\n\n"
}

func (m Model) View() tea.View {
	var body strings.Builder
	nameRow, depRow, depX, dbRow, cacheRow, appsRow, templatesRow := -1, -1, 0, -1, -1, -1, -1

	switch m.step {
	case StepProjectName:
		s, row := m.viewProjectName()
		body.WriteString(s)
		nameRow = row
	case StepDocker:
		body.WriteString(m.viewDocker())
	case StepDependencies:
		s, row, x := m.viewDependencies()
		body.WriteString(s)
		depRow, depX = row, x
	case StepDataBase:
		s, row := m.viewDatabases()
		body.WriteString(s)
		dbRow = row
	case StepCacheBackend:
		s, row := m.viewCacheBackend()
		body.WriteString(s)
		cacheRow = row
	case StepAppsCreated:
		s, row := m.viewAppsCreated()
		body.WriteString(s)
		appsRow = row
	case StepTemplates:
		s, row := m.viewTemplates()
		body.WriteString(s)
		templatesRow = row
	case StepConfirm:
		body.WriteString(m.viewConfirm())
	case StepExecuting:
		body.WriteString(m.viewExecuting())
	case StepDone:
		body.WriteString(m.viewDone())
	case StepError:
		body.WriteString(m.viewError())
	}

	bodyStr := body.String()
	prefix := m.headerFor(strings.Count(bodyStr, "\n"))
	base := strings.Count(prefix, "\n")
	nameRow, depRow, dbRow = nameRow+base, depRow+base, dbRow+base
	cacheRow, appsRow, templatesRow = cacheRow+base, appsRow+base, templatesRow+base

	frame := prefix + bodyStr
	if m.width > 0 {
		frame = clipLines(frame, m.width)
	}

	v := tea.NewView(frame)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	switch m.step {
	case StepProjectName:
		if c := m.textInput.Cursor(); c != nil {
			c.Y = nameRow
			v.Cursor = c
		}
	case StepDependencies:
		if depRow >= 0 {
			c := tea.NewCursor(depX, depRow)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	case StepDataBase:
		if dbRow >= 0 {
			c := tea.NewCursor(0, dbRow)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	case StepCacheBackend:
		if cacheRow >= 0 {
			c := tea.NewCursor(0, cacheRow)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	case StepAppsCreated:
		if appsRow >= 0 {
			c := tea.NewCursor(0, appsRow)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	case StepTemplates:
		if templatesRow >= 0 {
			c := tea.NewCursor(0, templatesRow)
			c.Shape = tea.CursorBar
			v.Cursor = c
		}
	}
	return v
}

func (m Model) viewProjectName() (string, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 1/8"))
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

	str_builder.WriteString(stepBadgeStyle.Render("Step 2/8"))
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

func (m Model) viewDependencies() (string, int, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 3/8"))
	str_builder.WriteString(stepHeaderStyle.Render("Dependencies"))
	str_builder.WriteString("\n\n")
	str_builder.WriteString("Select optional packages to install and configure alongside Django:\n\n")

	cellWidth := 64
	if m.width > 0 {
		cellWidth = (m.width - 2) / 2
		if cellWidth > 64 {
			cellWidth = 64
		}
		if cellWidth < 12 {
			cellWidth = 12
		}
	}

	row := -1
	col := 0

	for i := 0; i < len(m.depOptions); i += 2 {
		base := strings.Count(str_builder.String(), "\n")
		if m.depCursor == i {
			row, col = base, 0
		}
		line := m.depCell(i, cellWidth)
		if i+1 < len(m.depOptions) {
			if m.depCursor == i+1 {
				row, col = base, cellWidth+2
			}
			line += "  " + m.depCell(i+1, cellWidth)
		}
		str_builder.WriteString(line + "\n")
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓ or k/j: rows • ←/→ or h/l: columns • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back"))
	return str_builder.String(), row, col
}

func (m Model) depCell(idx int, width int) string {
	opt := m.depOptions[idx]

	checkbox := uncheckedStyle.Render("[ ]")
	if m.depChecked[opt.ID] {
		checkbox = checkedStyle.Render("[✓]")
	}

	nameStyle := unselectedItemStyle
	if m.depCursor == idx {
		nameStyle = selectedItemStyle
	}

	line := fmt.Sprintf("%s %s", checkbox, nameStyle.Render(opt.Name))
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(line)
}

func (m Model) viewDatabases() (string, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 4/8"))
	str_builder.WriteString(stepHeaderStyle.Render("Databases"))
	str_builder.WriteString("\n\n")
	str_builder.WriteString("Select the Database you want to configure for your project:\n\n")

	row := -1

	for i, opt := range m.dbOptions {
		if i == m.dbCursor {
			row = strings.Count(str_builder.String(), "\n")
		}

		cursor := "  "
		checkbox := uncheckedStyle.Render("[ ]")
		style := unselectedItemStyle

		if m.depChecked[opt.ID] {
			checkbox = checkedStyle.Render("[✓]")
		}

		if m.dbCursor == i {
			cursor = selectedItemStyle.Render("▸ ")
			style = selectedItemStyle
		}

		str_builder.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, checkbox, style.Render(opt.Name), subtitleStyle.Render(opt.Description)))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓ or k/j: navigate • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back"))
	return str_builder.String(), row
}

func (m Model) viewCacheBackend() (string, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 5/8"))
	str_builder.WriteString(stepHeaderStyle.Render("Cache Backend"))
	str_builder.WriteString("\n\n")
	str_builder.WriteString("Choose the cache backend for the CACHES setting:\n\n")

	row := -1

	for i, opt := range m.cacheOptions {
		if i == m.cacheChoice {
			row = strings.Count(str_builder.String(), "\n")
		}

		cursor := "  "
		radio := "( )"
		style := unselectedItemStyle

		if m.cacheChoice == i {
			cursor = selectedItemStyle.Render("▸ ")
			radio = checkedStyle.Render("(•)")
			style = selectedItemStyle
		}

		str_builder.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, radio, style.Render(opt.Name), subtitleStyle.Render(opt.Description)))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓ or hjkl: choose • Enter: continue • Esc/Backspace: back"))
	return str_builder.String(), row
}

func (m Model) viewAppsCreated() (string, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 6/8"))
	str_builder.WriteString(stepHeaderStyle.Render("Starter Apps"))
	str_builder.WriteString("\n\n")

	if m.appsPhase == 0 {
		str_builder.WriteString("Would you like to create starter apps for this project?\n\n")

		options := []struct {
			label string
			desc  string
		}{
			{"Yes", "Create starter apps (api, base)"},
			{"No", "Skip app creation"},
		}

		row := -1
		for i, opt := range options {
			if m.appsChoice == i {
				row = strings.Count(str_builder.String(), "\n")
			}

			cursor := "  "
			radio := "( )"
			style := unselectedItemStyle

			if m.appsChoice == i {
				cursor = selectedItemStyle.Render("▸ ")
				radio = checkedStyle.Render("(•)")
				style = selectedItemStyle
			}

			str_builder.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, radio, style.Render(opt.label), subtitleStyle.Render(opt.desc)))
		}

		str_builder.WriteString("\n")
		str_builder.WriteString(helpStyle.Render("↑/↓ or hjkl: choose • y: yes, then pick apps • n: no, skip • Enter: continue • Esc/Backspace: back"))
		return str_builder.String(), row
	}

	str_builder.WriteString("Select the apps to create:\n\n")

	row := -1
	for i, opt := range m.appOptions {
		if i == m.appsCursor {
			row = strings.Count(str_builder.String(), "\n")
		}

		cursor := "  "
		checkbox := uncheckedStyle.Render("[ ]")
		style := unselectedItemStyle

		if m.appsChecked[opt.ID] {
			checkbox = checkedStyle.Render("[✓]")
		}

		if m.appsCursor == i {
			cursor = selectedItemStyle.Render("▸ ")
			style = selectedItemStyle
		}

		str_builder.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, checkbox, style.Render(opt.Name), subtitleStyle.Render(opt.Description)))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓ or k/j: navigate • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back to question"))
	return str_builder.String(), row
}

func (m Model) viewTemplates() (string, int) {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 7/8"))
	str_builder.WriteString(stepHeaderStyle.Render("Templates Folder"))
	str_builder.WriteString("\n\n")
	str_builder.WriteString("Where should template folders live?\n\n")

	opts := m.templateOptions()
	choice := m.templatesChoice
	if choice >= len(opts) {
		choice = 0
	}

	row := -1
	for i, opt := range opts {
		if choice == i {
			row = strings.Count(str_builder.String(), "\n")
		}

		cursor := "  "
		radio := "( )"
		style := unselectedItemStyle

		if choice == i {
			cursor = selectedItemStyle.Render("▸ ")
			radio = checkedStyle.Render("(•)")
			style = selectedItemStyle
		}

		str_builder.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, radio, style.Render(opt.Name), subtitleStyle.Render(opt.Description)))
	}

	str_builder.WriteString("\n")
	str_builder.WriteString(helpStyle.Render("↑/↓ or hjkl: choose • Enter: continue • Esc/Backspace: back"))
	return str_builder.String(), row
}

func (m Model) viewConfirm() string {
	var str_builder strings.Builder

	str_builder.WriteString(stepBadgeStyle.Render("Step 8/8"))
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

	var selectedDBs []string
	for _, opt := range m.dbOptions {
		if m.depChecked[opt.ID] {
			selectedDBs = append(selectedDBs, opt.Name)
		}
	}
	dbText := "None (Django defaults)"
	if len(selectedDBs) > 0 {
		dbText = strings.Join(selectedDBs, ", ")
	}

	cacheText := m.cacheOptions[m.cacheChoice].Name

	appsText := "None"
	if apps := m.selectedApps(); len(apps) > 0 {
		appsText = strings.Join(apps, ", ")
	}

	templatesText := "None"
	if opts := m.templateOptions(); m.templatesChoice < len(opts) {
		templatesText = opts[m.templatesChoice].Name
	}

	toolText := "python -m venv & pip"
	if m.hasUv {
		toolText = "uv (ultra-fast virtualenv & pip)"
	}

	cardContent := fmt.Sprintf(
		"Project Name : %s\nTarget Path  : ./%s\nDocker       : %s\nDependencies : Django, %s\nDatabase     : %s\nCache        : %s\nApps         : %s\nTemplates    : %s\nEnvironment  : %s",
		selectedItemStyle.Render(m.cfg.ProjectName),
		m.cfg.ProjectName,
		dockerText,
		depsText,
		dbText,
		cacheText,
		appsText,
		templatesText,
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

	str_builder.WriteString(cardStyle.Render(m.nextStepsText()))
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
