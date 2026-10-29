package tui

import (
	"errors"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

var errGenerationCancelled = errors.New("generation cancelled")

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			if m.cancelGen != nil {
				m.cancelGen()
				m.cancelGen = nil
			}
			return m, tea.Quit
		}

		switch m.step {
		case StepProjectName:
			return m.updateProjectName(msg)

		case StepDocker:
			return m.updateDocker(msg)

		case StepDependencies:
			return m.updateDependencies(msg)

		case StepDataBase:
			return m.updateDataBases(msg)

		case StepCacheBackend:
			return m.updateCacheBackend(msg)

		case StepAppsCreated:
			return m.updateAppsCreated(msg)

		case StepTemplates:
			return m.updateTemplates(msg)

		case StepConfirm:
			return m.updateConfirm(msg)

		case StepExecuting:
			if msg.String() == "q" || msg.String() == "n" || msg.String() == "esc" {
				if m.cancelGen != nil {
					m.cancelGen()
					m.cancelGen = nil
				}
				m.clearProgress()
				m.err = errGenerationCancelled
				m.step = StepError
				return m, nil
			}

		case StepDone, StepError:
			if msg.String() == "q" || msg.String() == "esc" || msg.String() == "enter" {
				return m, tea.Quit
			}
		}

	case tea.PasteMsg:
		if m.step == StepProjectName {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.step != StepExecuting {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progressMsg:
		if m.step != StepExecuting {
			return m, nil
		}
		if m.currentTask != "" {
			m.progressItems = append(m.progressItems, m.currentTask)
		}
		m.currentTask = msg.text
		m.progressStep = msg.step
		m.progressTotal = msg.total
		return m, waitForActivity(m.progressSub)

	case doneMsg:
		if m.step != StepExecuting {
			return m, nil
		}
		if m.cancelGen != nil {
			m.cancelGen()
			m.cancelGen = nil
		}
		if msg.err != nil {
			m.err = msg.err
			m.clearProgress()
			m.step = StepError
			return m, nil
		}
		if m.currentTask != "" {
			m.progressItems = append(m.progressItems, m.currentTask)
			m.currentTask = ""
		}
		m.progressStep = 0
		m.progressTotal = 0
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
		if m.depCursor >= 2 {
			m.depCursor -= 2
		} else {
			// wrap to the last row of the same column
			last := len(m.depOptions) - 1
			if last%2 != m.depCursor%2 {
				last--
			}
			m.depCursor = last
		}
	case "down", "j":
		if m.depCursor+2 < len(m.depOptions) {
			m.depCursor += 2
		} else {
			m.depCursor = m.depCursor % 2
		}
	case "left", "h":
		if m.depCursor%2 == 1 {
			m.depCursor--
		}
	case "right", "l":
		if m.depCursor%2 == 0 && m.depCursor+1 < len(m.depOptions) {
			m.depCursor++
		}
	case "space", "x":
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
		m.step = StepDataBase
		return m, nil
	case "esc", "backspace":
		m.step = StepDocker
		return m, nil
	}
	return m, nil
}

func (m Model) updateDataBases(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.dbCursor > 0 {
			m.dbCursor--
		} else {
			m.dbCursor = len(m.dbOptions) - 1
		}
	case "down", "j":
		if m.dbCursor < len(m.dbOptions)-1 {
			m.dbCursor++
		} else {
			m.dbCursor = 0
		}
	case "space", "x":
		currentID := m.dbOptions[m.dbCursor].ID
		m.depChecked[currentID] = !m.depChecked[currentID]
	case "a":
		allSelected := true
		for _, opt := range m.dbOptions {
			if !m.depChecked[opt.ID] {
				allSelected = false
				break
			}
		}
		for _, opt := range m.dbOptions {
			m.depChecked[opt.ID] = !allSelected
		}
	case "enter":
		m.step = StepCacheBackend
		return m, nil
	case "esc", "backspace":
		m.step = StepDependencies
		return m, nil
	}
	return m, nil
}

func (m Model) updateCacheBackend(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "left", "h":
		if m.cacheChoice > 0 {
			m.cacheChoice--
		} else {
			m.cacheChoice = len(m.cacheOptions) - 1
		}
	case "down", "j", "right", "l":
		if m.cacheChoice < len(m.cacheOptions)-1 {
			m.cacheChoice++
		} else {
			m.cacheChoice = 0
		}
	case "enter":
		if m.cacheOptions[m.cacheChoice].ID == "redis" {
			m.depChecked["django-redis"] = true
		}
		m.step = StepAppsCreated
		m.appsPhase = 0
		return m, nil
	case "esc", "backspace":
		m.step = StepDataBase
		return m, nil
	}
	return m, nil
}

func (m Model) updateAppsCreated(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.appsPhase {
	case 0: // ask yes/no
		switch msg.String() {
		case "up", "k", "left", "h":
			m.appsChoice = 0
		case "down", "j", "right", "l":
			m.appsChoice = 1
		case "y", "Y":
			m.appsChoice = 1
			m.appsPhase = 1
			return m, nil
		case "n", "N":
			m.appsChoice = 0
			m.appsChecked = make(map[string]bool)
			m.toTemplates()
			return m, nil
		case "enter":
			if m.appsChoice == 1 {
				m.appsPhase = 1
			} else {
				m.appsChecked = make(map[string]bool)
				m.toTemplates()
			}
			return m, nil
		case "esc", "backspace":
			m.step = StepCacheBackend
			return m, nil
		}

	case 1: // select apps
		switch msg.String() {
		case "up", "k":
			if m.appsCursor > 0 {
				m.appsCursor--
			} else {
				m.appsCursor = len(m.appOptions) - 1
			}
		case "down", "j":
			if m.appsCursor < len(m.appOptions)-1 {
				m.appsCursor++
			} else {
				m.appsCursor = 0
			}
		case "space", "x":
			currentID := m.appOptions[m.appsCursor].ID
			m.appsChecked[currentID] = !m.appsChecked[currentID]
		case "a":
			allSelected := true
			for _, opt := range m.appOptions {
				if !m.appsChecked[opt.ID] {
					allSelected = false
					break
				}
			}
			for _, opt := range m.appOptions {
				m.appsChecked[opt.ID] = !allSelected
			}
		case "enter":
			m.toTemplates()
			return m, nil
		case "esc", "backspace":
			m.appsPhase = 0
			return m, nil
		}
	}
	return m, nil
}

// toTemplates clamps templatesChoice against the options available for the
// current app selection before entering the templates step.
func (m *Model) toTemplates() {
	if m.templatesChoice >= len(m.templateOptions()) {
		m.templatesChoice = 0
	}
	m.step = StepTemplates
}

func (m Model) updateTemplates(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	opts := m.templateOptions()
	switch msg.String() {
	case "up", "k", "left", "h":
		if m.templatesChoice > 0 {
			m.templatesChoice--
		} else {
			m.templatesChoice = len(opts) - 1
		}
	case "down", "j", "right", "l":
		if m.templatesChoice < len(opts)-1 {
			m.templatesChoice++
		} else {
			m.templatesChoice = 0
		}
	case "enter":
		m.step = StepConfirm
		return m, nil
	case "esc", "backspace":
		m.step = StepAppsCreated
		m.appsPhase = 0
		return m, nil
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y", "Y":
		m.step = StepExecuting
		cmd := m.startGeneration()
		return m, cmd
	case "esc", "backspace", "b":
		m.step = StepTemplates
		return m, nil
	case "q", "n", "N":
		return m, tea.Quit
	}
	return m, nil
}
