package tui

import (
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

		switch m.step {
		case StepProjectName:
			return m.updateProjectName(msg)

		case StepDocker:
			return m.updateDocker(msg)

		case StepDependencies:
			return m.updateDependencies(msg)

		case StepConfirm:
			return m.updateConfirm(msg)

		case StepDone, StepError:
			if msg.String() == "q" || msg.String() == "esc" {
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progressMsg:
		if m.currentTask != "" {
			m.progressItems = append(m.progressItems, m.currentTask)
		}
		m.currentTask = msg.text
		return m, waitForActivity(m.progressSub)

	case doneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = StepError
			return m, nil
		}
		if m.currentTask != "" {
			m.progressItems = append(m.progressItems, m.currentTask)
			m.currentTask = ""
		}
		m.step = StepDone
		return m, nil
	}

	return m, nil
}

func (m Model) updateProjectName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.textInput.Value())
		if err := validateProjectName(name); err != nil {
			m.nameError = err.Error()
			return m, nil
		}
		m.nameError = ""
		m.cfg.ProjectName = name
		m.step = StepDocker
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updateDocker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "left", "h":
		m.dockerChoice = 0
	case "down", "j", "right", "l":
		m.dockerChoice = 1
	case "y", "Y":
		m.dockerChoice = 0
		m.step = StepDependencies
		return m, nil
	case "n", "N":
		m.dockerChoice = 1
		m.step = StepDependencies
		return m, nil
	case "enter":
		m.step = StepDependencies
		return m, nil
	case "esc", "backspace":
		m.step = StepProjectName
		return m, nil
	}
	return m, nil
}

func (m Model) updateDependencies(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.depCursor > 0 {
			m.depCursor--
		} else {
			m.depCursor = len(m.depOptions) - 1
		}
	case "down", "j":
		if m.depCursor < len(m.depOptions)-1 {
			m.depCursor++
		} else {
			m.depCursor = 0
		}
	case " ", "x":
		currentID := m.depOptions[m.depCursor].ID
		m.depChecked[currentID] = !m.depChecked[currentID]
	case "a":
		allSelected := true
		for _, opt := range m.depOptions {
			if !m.depChecked[opt.ID] {
				allSelected = false
				break
			}
		}
		for _, opt := range m.depOptions {
			m.depChecked[opt.ID] = !allSelected
		}
	case "enter":
		m.step = StepConfirm
		return m, nil
	case "esc", "backspace":
		m.step = StepDocker
		return m, nil
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y", "Y":
		m.step = StepExecuting
		return m, m.startGeneration()
	case "esc", "backspace", "b":
		m.step = StepDependencies
		return m, nil
	case "q", "n", "N":
		return m, tea.Quit
	}
	return m, nil
}
