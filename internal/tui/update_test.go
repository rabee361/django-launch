package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Helper to create key press message
func keyMsg(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Text: text})
}

func specialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func ctrlKey(char rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: char, Mod: tea.ModCtrl})
}

func TestUpdate_StepProjectName(t *testing.T) {
	// TEST-TUI-07: Valid Enter advances to StepDocker
	m := NewModel()
	m.textInput.SetValue("my_valid_app")

	updatedModel, cmd := m.Update(specialKey(tea.KeyEnter))
	m = updatedModel.(Model)

	if m.step != StepDocker {
		t.Errorf("TEST-TUI-07: expected step to be StepDocker, got %v", m.step)
	}
	if m.cfg.ProjectName != "my_valid_app" {
		t.Errorf("TEST-TUI-07: expected cfg.ProjectName = 'my_valid_app', got %q", m.cfg.ProjectName)
	}
	if m.nameError != "" {
		t.Errorf("TEST-TUI-07: expected empty nameError, got %q", m.nameError)
	}
	if cmd != nil {
		t.Errorf("TEST-TUI-07: expected nil cmd on valid enter")
	}

	// TEST-TUI-08: Invalid Enter remains on StepProjectName and sets nameError
	m = NewModel()
	m.textInput.SetValue("invalid-app-name")

	updatedModel, _ = m.Update(specialKey(tea.KeyEnter))
	m = updatedModel.(Model)

	if m.step != StepProjectName {
		t.Errorf("TEST-TUI-08: expected step to remain StepProjectName, got %v", m.step)
	}
	if m.nameError == "" {
		t.Errorf("TEST-TUI-08: expected non-empty nameError for invalid name")
	}

	// Typing 'q' into text input should not quit
	m = NewModel()
	updatedModel, cmd = m.Update(keyMsg('q', "q"))
	m = updatedModel.(Model)
	if m.step != StepProjectName {
		t.Errorf("typing 'q' changed step unexpectedly")
	}
}

func TestUpdate_StepDocker(t *testing.T) {
	// TEST-TUI-09: Navigation & shortcuts
	m := NewModel()
	m.step = StepDocker
	m.dockerChoice = 0 // Yes

	// Down arrow toggles to No (1)
	updatedModel, _ := m.Update(specialKey(tea.KeyDown))
	m = updatedModel.(Model)
	if m.dockerChoice != 1 {
		t.Errorf("TEST-TUI-09: expected dockerChoice = 1 after KeyDown, got %d", m.dockerChoice)
	}

	// Up arrow toggles back to Yes (0)
	updatedModel, _ = m.Update(specialKey(tea.KeyUp))
	m = updatedModel.(Model)
	if m.dockerChoice != 0 {
		t.Errorf("TEST-TUI-09: expected dockerChoice = 0 after KeyUp, got %d", m.dockerChoice)
	}

	// 'n' selects No and advances to StepDependencies
	updatedModel, _ = m.Update(keyMsg('n', "n"))
	m = updatedModel.(Model)
	if m.dockerChoice != 1 {
		t.Errorf("TEST-TUI-09: expected dockerChoice = 1 after 'n', got %d", m.dockerChoice)
	}
	if m.step != StepDependencies {
		t.Errorf("TEST-TUI-09: expected step StepDependencies after 'n', got %v", m.step)
	}

	// Esc returns back to StepProjectName
	m.step = StepDocker
	updatedModel, _ = m.Update(specialKey(tea.KeyEsc))
	m = updatedModel.(Model)
	if m.step != StepProjectName {
		t.Errorf("TEST-TUI-09: expected step StepProjectName after Esc, got %v", m.step)
	}

	// Backspace also returns back to StepProjectName
	m.step = StepDocker
	updatedModel, _ = m.Update(specialKey(tea.KeyBackspace))
	m = updatedModel.(Model)
	if m.step != StepProjectName {
		t.Errorf("TEST-TUI-09: expected step StepProjectName after Backspace, got %v", m.step)
	}
}

func TestUpdate_StepDependencies(t *testing.T) {
	m := NewModel()
	m.step = StepDependencies

	// TEST-TUI-10: Navigation & Space toggle
	if m.depCursor != 0 {
		t.Fatalf("expected initial depCursor = 0, got %d", m.depCursor)
	}

	// Move down
	updatedModel, _ := m.Update(specialKey(tea.KeyDown))
	m = updatedModel.(Model)
	if m.depCursor != 1 {
		t.Errorf("TEST-TUI-10: expected depCursor = 1, got %d", m.depCursor)
	}

	// Wrap around backwards with KeyUp from 0
	m.depCursor = 0
	updatedModel, _ = m.Update(specialKey(tea.KeyUp))
	m = updatedModel.(Model)
	if m.depCursor != len(m.depOptions)-1 {
		t.Errorf("TEST-TUI-10: expected cursor to wrap to %d, got %d", len(m.depOptions)-1, m.depCursor)
	}

	// Wrap around forwards with KeyDown from end
	updatedModel, _ = m.Update(specialKey(tea.KeyDown))
	m = updatedModel.(Model)
	if m.depCursor != 0 {
		t.Errorf("TEST-TUI-10: expected cursor to wrap back to 0, got %d", m.depCursor)
	}

	// Space toggles current item
	targetID := m.depOptions[0].ID
	if m.depChecked[targetID] {
		t.Fatalf("expected %q initially unchecked", targetID)
	}

	updatedModel, _ = m.Update(specialKey(tea.KeySpace))
	m = updatedModel.(Model)
	if !m.depChecked[targetID] {
		t.Errorf("TEST-TUI-10: expected %q to be checked after Space", targetID)
	}

	// Toggle again with 'x'
	updatedModel, _ = m.Update(keyMsg('x', "x"))
	m = updatedModel.(Model)
	if m.depChecked[targetID] {
		t.Errorf("TEST-TUI-10: expected %q to be unchecked after 'x'", targetID)
	}

	// TEST-TUI-11: 'a' toggles all on / off
	updatedModel, _ = m.Update(keyMsg('a', "a"))
	m = updatedModel.(Model)
	for _, opt := range m.depOptions {
		if !m.depChecked[opt.ID] {
			t.Errorf("TEST-TUI-11: expected %q to be checked after 'a'", opt.ID)
		}
	}

	// Press 'a' again when all are selected -> toggles all off
	updatedModel, _ = m.Update(keyMsg('a', "a"))
	m = updatedModel.(Model)
	for _, opt := range m.depOptions {
		if m.depChecked[opt.ID] {
			t.Errorf("TEST-TUI-11: expected %q to be unchecked after second 'a'", opt.ID)
		}
	}

	// TEST-TUI-12: Enter advances to StepConfirm
	updatedModel, _ = m.Update(specialKey(tea.KeyEnter))
	m = updatedModel.(Model)
	if m.step != StepConfirm {
		t.Errorf("TEST-TUI-12: expected StepConfirm after Enter, got %v", m.step)
	}

	// Esc returns to StepDocker
	m.step = StepDependencies
	updatedModel, _ = m.Update(specialKey(tea.KeyEsc))
	m = updatedModel.(Model)
	if m.step != StepDocker {
		t.Errorf("TEST-TUI-12: expected StepDocker after Esc, got %v", m.step)
	}
}

func TestUpdate_StepConfirm(t *testing.T) {
	// TEST-TUI-13: Confirm & Cancel
	m := NewModel()
	m.step = StepConfirm
	m.cfg.ProjectName = "my_app"

	// Esc returns to StepDependencies
	updatedModel, _ := m.Update(specialKey(tea.KeyEsc))
	m = updatedModel.(Model)
	if m.step != StepDependencies {
		t.Errorf("TEST-TUI-13: expected StepDependencies after Esc, got %v", m.step)
	}

	// 'b' also returns to StepDependencies
	m.step = StepConfirm
	updatedModel, _ = m.Update(keyMsg('b', "b"))
	m = updatedModel.(Model)
	if m.step != StepDependencies {
		t.Errorf("TEST-TUI-13: expected StepDependencies after 'b', got %v", m.step)
	}

	// 'q' quits
	m.step = StepConfirm
	_, cmd := m.Update(keyMsg('q', "q"))
	if cmd == nil {
		t.Errorf("TEST-TUI-13: expected tea.Quit command after 'q', got nil")
	}

	// 'n' quits
	m.step = StepConfirm
	_, cmd = m.Update(keyMsg('n', "n"))
	if cmd == nil {
		t.Errorf("TEST-TUI-13: expected tea.Quit command after 'n', got nil")
	}
}

func TestUpdate_StepExecuting_Messages(t *testing.T) {
	m := NewModel()
	m.step = StepExecuting
	m.currentTask = "Initial task"

	// TEST-TUI-14: progressMsg appends previous task and updates currentTask
	pMsg := progressMsg{
		step:  2,
		total: 6,
		text:  "New progress task",
	}

	updatedModel, _ := m.Update(pMsg)
	m = updatedModel.(Model)

	if len(m.progressItems) != 1 || m.progressItems[0] != "Initial task" {
		t.Errorf("TEST-TUI-14: progressItems expected ['Initial task'], got %v", m.progressItems)
	}
	if m.currentTask != "New progress task" {
		t.Errorf("TEST-TUI-14: currentTask expected 'New progress task', got %q", m.currentTask)
	}

	// TEST-TUI-15: doneMsg with nil error -> StepDone
	dMsgSuccess := doneMsg{err: nil}
	updatedModel, _ = m.Update(dMsgSuccess)
	m = updatedModel.(Model)

	if m.step != StepDone {
		t.Errorf("TEST-TUI-15: expected step StepDone, got %v", m.step)
	}
	if len(m.progressItems) != 2 || m.progressItems[1] != "New progress task" {
		t.Errorf("TEST-TUI-15: expected final task added to progressItems, got %v", m.progressItems)
	}
	if m.currentTask != "" {
		t.Errorf("TEST-TUI-15: expected currentTask to be cleared, got %q", m.currentTask)
	}

	// TEST-TUI-16: doneMsg with error -> StepError
	m = NewModel()
	m.step = StepExecuting
	expectedErr := errors.New("pip install failed")
	dMsgFail := doneMsg{err: expectedErr}

	updatedModel, _ = m.Update(dMsgFail)
	m = updatedModel.(Model)

	if m.step != StepError {
		t.Errorf("TEST-TUI-16: expected step StepError, got %v", m.step)
	}
	if m.err != expectedErr {
		t.Errorf("TEST-TUI-16: expected m.err to match, got %v", m.err)
	}
}

func TestUpdate_GlobalQuit(t *testing.T) {
	// TEST-TUI-17: Ctrl+C quits regardless of step
	steps := []Step{
		StepProjectName,
		StepDocker,
		StepDependencies,
		StepConfirm,
		StepExecuting,
		StepDone,
		StepError,
	}

	for _, s := range steps {
		m := NewModel()
		m.step = s
		_, cmd := m.Update(ctrlKey('c'))
		if cmd == nil {
			t.Errorf("TEST-TUI-17: expected non-nil tea.Quit on Ctrl+C at step %v", s)
		}
	}
}
