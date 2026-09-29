package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestViewExecuting_ProgressFraction(t *testing.T) {
	// TEST-TUI-22: the executing view shows the step fraction on the task line
	m := NewModel()
	m.step = StepExecuting
	m.currentTask = "Installing dependencies"
	m.progressStep = 3
	m.progressTotal = 6

	view := m.View().Content
	found := false
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Installing dependencies") && strings.Contains(line, "(3/6)") {
			found = true
		}
	}
	if !found {
		t.Errorf("TEST-TUI-22: expected a single line with the task and (3/6), got:\n%s", view)
	}
}

func TestView_WidthClipping(t *testing.T) {
	// TEST-TUI-23: no rendered line is wider than m.width
	for _, step := range []Step{StepDependencies, StepConfirm} {
		m := NewModel()
		m.step = step
		m.width = 40

		view := m.View().Content
		for _, line := range strings.Split(view, "\n") {
			if w := lipgloss.Width(line); w > 40 {
				t.Errorf("TEST-TUI-23: step %v line wider than m.width (%d): %q", step, w, line)
			}
		}
	}
}

func TestView_WidthZeroUnchanged(t *testing.T) {
	// TEST-TUI-24: width 0 leaves the frame uncapped
	m := NewModel()
	m.step = StepDependencies

	uncapped := m.View().Content
	m.width = 500
	wide := m.View().Content

	if uncapped != wide {
		t.Errorf("TEST-TUI-24: expected identical frame at width 0 and 500")
	}
	if !strings.Contains(uncapped, "Python Imaging Library (required for ImageField & media)") {
		t.Errorf("TEST-TUI-24: expected full description line at width 0")
	}
	if !strings.Contains(uncapped, "Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back") {
		t.Errorf("TEST-TUI-24: expected full help line at width 0")
	}
}
