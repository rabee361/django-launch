package tui

import (
	"context"
	"errors"
	"testing"

	"charm.land/bubbles/v2/spinner"
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
	if m.progressStep != 2 || m.progressTotal != 6 {
		t.Errorf("TEST-TUI-14: expected progressStep 2 and progressTotal 6, got %d/%d", m.progressStep, m.progressTotal)
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
	m.progressItems = []string{"first task"}
	m.currentTask = "last task"
	m.progressStep = 3
	m.progressTotal = 6
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
	if len(m.progressItems) != 0 {
		t.Errorf("TEST-TUI-16: expected progressItems to be cleared, got %v", m.progressItems)
	}
	if m.currentTask != "" {
		t.Errorf("TEST-TUI-16: expected currentTask to be cleared, got %q", m.currentTask)
	}
	if m.progressStep != 0 || m.progressTotal != 0 {
		t.Errorf("TEST-TUI-16: expected progressStep and progressTotal to be cleared, got %d/%d", m.progressStep, m.progressTotal)
	}
}

func TestUpdate_ExecutingCancel(t *testing.T) {
	// TEST-TUI-21: q/n/esc cancel a running generation and clear progress state
	for _, key := range []string{"q", "n", "esc"} {
		m := NewModel()
		m.step = StepExecuting
		m.progressItems = []string{"Creating virtual environment (.venv)..."}
		m.currentTask = "Creating virtual environment (.venv)..."
		m.progressStep = 2
		m.progressTotal = 5

		var press tea.KeyPressMsg
		if key == "esc" {
			press = specialKey(tea.KeyEsc)
		} else {
			press = keyMsg(rune(key[0]), key)
		}

		updatedModel, cmd := m.Update(press)
		m = updatedModel.(Model)

		if m.step != StepError {
			t.Errorf("TEST-TUI-21: expected StepError after %q, got %v", key, m.step)
		}
		if !errors.Is(m.err, errGenerationCancelled) {
			t.Errorf("TEST-TUI-21: expected errGenerationCancelled after %q, got %v", key, m.err)
		}
		if cmd != nil {
			t.Errorf("TEST-TUI-21: expected nil cmd after %q, got %v", key, cmd)
		}
		if m.cancelGen != nil {
			t.Errorf("TEST-TUI-21: expected cancelGen to be cleared after %q", key)
		}
		if len(m.progressItems) != 0 || m.currentTask != "" || m.progressStep != 0 || m.progressTotal != 0 {
			t.Errorf("TEST-TUI-21: expected progress state cleared after %q, got items=%v task=%q step=%d total=%d",
				key, m.progressItems, m.currentTask, m.progressStep, m.progressTotal)
		}
	}

	m := NewModel()
	m.step = StepExecuting
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelGen = cancel

	updatedModel, _ := m.Update(keyMsg('q', "q"))
	m = updatedModel.(Model)

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("TEST-TUI-21: expected the generation context to be cancelled")
	}
	if m.cancelGen != nil {
		t.Errorf("TEST-TUI-21: expected cancelGen to be cleared after the cancel key")
	}

	// TEST-TUI-21 (ctrl+c): cancelling quits the program and stops the generator
	m = NewModel()
	m.step = StepExecuting
	ctx, cancel = context.WithCancel(context.Background())
	m.cancelGen = cancel

	updatedModel, cmd := m.Update(ctrlKey('c'))
	m = updatedModel.(Model)

	if cmd == nil {
		t.Errorf("TEST-TUI-21: expected a quit cmd for ctrl+c during generation")
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("TEST-TUI-21: expected ctrl+c to cancel the generation context")
	}
	if m.cancelGen != nil {
		t.Errorf("TEST-TUI-21: expected cancelGen to be cleared after ctrl+c")
	}
}

func TestUpdate_ConfirmStartsGeneration(t *testing.T) {
	// TEST-TUI-25: confirming the summary starts generation and returns a cancellable model
	m := NewModel()
	m.step = StepConfirm
	m.cfg.OutputDir = t.TempDir()
	m.cfg.UseUv = false
	m.cfg.PythonCmd = "non_existent_python_binary_for_tests"

	updatedModel, cmd := m.Update(keyMsg('y', "y"))
	m = updatedModel.(Model)

	if m.step != StepExecuting {
		t.Errorf("TEST-TUI-25: expected StepExecuting after confirm, got %v", m.step)
	}
	if m.cancelGen == nil {
		t.Errorf("TEST-TUI-25: expected cancelGen to be set after confirm")
	}
	if m.progressSub == nil {
		t.Errorf("TEST-TUI-25: expected progressSub to be set after confirm")
	}
	if cmd == nil {
		t.Errorf("TEST-TUI-25: expected a non-nil cmd after confirm")
	}

	if m.cancelGen != nil {
		m.cancelGen()
	}
}

func TestUpdate_DroppedStaleMessages(t *testing.T) {
	// TEST-TUI-26: progress and done messages arriving outside StepExecuting are dropped
	m := NewModel()
	m.step = StepError
	m.err = errGenerationCancelled

	updatedModel, cmd := m.Update(doneMsg{err: errors.New("context canceled")})
	m = updatedModel.(Model)
	if m.step != StepError {
		t.Errorf("TEST-TUI-26: expected stale doneMsg to keep StepError, got %v", m.step)
	}
	if !errors.Is(m.err, errGenerationCancelled) {
		t.Errorf("TEST-TUI-26: expected stale doneMsg not to overwrite err, got %v", m.err)
	}
	if cmd != nil {
		t.Errorf("TEST-TUI-26: expected nil cmd for a dropped doneMsg, got %v", cmd)
	}

	updatedModel, cmd = m.Update(doneMsg{err: nil})
	m = updatedModel.(Model)
	if m.step != StepError {
		t.Errorf("TEST-TUI-26: expected stale successful doneMsg to keep StepError, got %v", m.step)
	}
	if cmd != nil {
		t.Errorf("TEST-TUI-26: expected nil cmd for a dropped doneMsg, got %v", cmd)
	}

	updatedModel, cmd = m.Update(progressMsg{step: 4, total: 6, text: "late task"})
	m = updatedModel.(Model)
	if m.step != StepError {
		t.Errorf("TEST-TUI-26: expected stale progressMsg to keep StepError, got %v", m.step)
	}
	if m.currentTask != "" || len(m.progressItems) != 0 || m.progressStep != 0 || m.progressTotal != 0 {
		t.Errorf("TEST-TUI-26: expected stale progressMsg not to mutate progress state, got task=%q items=%v step=%d total=%d",
			m.currentTask, m.progressItems, m.progressStep, m.progressTotal)
	}
	if cmd != nil {
		t.Errorf("TEST-TUI-26: expected nil cmd for a dropped progressMsg, got %v", cmd)
	}
}

func TestUpdate_SpinnerTick(t *testing.T) {
	// TEST-TUI-19: spinner ticks are only re-armed while executing
	m := NewModel()
	m.step = StepDone

	updatedModel, cmd := m.Update(spinner.TickMsg{})
	m = updatedModel.(Model)
	if cmd != nil {
		t.Errorf("TEST-TUI-19: expected nil cmd for spinner.TickMsg at StepDone, got %v", cmd)
	}

	m.step = StepProjectName
	updatedModel, cmd = m.Update(spinner.TickMsg{})
	m = updatedModel.(Model)
	if cmd != nil {
		t.Errorf("TEST-TUI-19: expected nil cmd for spinner.TickMsg at StepProjectName, got %v", cmd)
	}

	m.step = StepExecuting
	updatedModel, cmd = m.Update(spinner.TickMsg{})
	m = updatedModel.(Model)
	if cmd == nil {
		t.Errorf("TEST-TUI-19: expected non-nil cmd for spinner.TickMsg at StepExecuting")
	}
}

func TestUpdate_WindowSize(t *testing.T) {
	// TEST-TUI-20: WindowSizeMsg stores the terminal width
	m := NewModel()

	updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updatedModel.(Model)

	if m.width != 120 {
		t.Errorf("TEST-TUI-20: expected width 120, got %d", m.width)
	}
	if cmd != nil {
		t.Errorf("TEST-TUI-20: expected nil cmd, got %v", cmd)
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

func TestUpdate_Paste(t *testing.T) {
	// TEST-TUI-18: PasteMsg is routed to the text input on step 1
	m := NewModel()
	updated, _ := m.Update(tea.PasteMsg{Content: "pasted_name"})
	m = updated.(Model)
	if got := m.textInput.Value(); got != "pasted_name" {
		t.Errorf("TEST-TUI-18: expected pasted value %q, got %q", "pasted_name", got)
	}

	// PasteMsg is ignored on steps without a text input
	m2 := NewModel()
	m2.step = StepDocker
	m2.textInput.SetValue("kept")
	updated, _ = m2.Update(tea.PasteMsg{Content: "ignored"})
	m2 = updated.(Model)
	if got := m2.textInput.Value(); got != "kept" {
		t.Errorf("TEST-TUI-18: expected value %q on StepDocker, got %q", "kept", got)
	}
}
