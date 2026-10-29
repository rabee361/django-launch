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
	if !strings.Contains(uncapped, "Pillow") {
		t.Errorf("TEST-TUI-24: expected dependency name in frame")
	}
	if !strings.Contains(uncapped, "↑/↓ or k/j: rows • ←/→ or h/l: columns • Space/x: toggle • a: toggle all • Enter: continue • Esc/Backspace: back") {
		t.Errorf("TEST-TUI-24: expected full help line at width 0")
	}
}

func TestView_ResponsiveHeader(t *testing.T) {
	// TEST-TUI-34: full banner when it fits, compact title when it does not
	m := NewModel()
	m.step = StepDocker
	m.height = 0
	if !strings.Contains(m.View().Content, "██████╗") {
		t.Errorf("TEST-TUI-34: expected full banner at height 0")
	}

	m.height = 100
	if !strings.Contains(m.View().Content, "██████╗") {
		t.Errorf("TEST-TUI-34: expected full banner when it fits")
	}

	m.height = 8
	content := m.View().Content
	if strings.Contains(content, "██████╗") {
		t.Errorf("TEST-TUI-34: expected banner hidden on a short terminal")
	}
	if !strings.Contains(content, "DJANGO LAUNCH") {
		t.Errorf("TEST-TUI-34: expected compact title on a short terminal, got:\n%s", content)
	}

	// Compact mode must still render the step content
	if !strings.Contains(content, "Docker Support") {
		t.Errorf("TEST-TUI-34: expected step content below the compact title, got:\n%s", content)
	}
}

func TestView_StepBadges(t *testing.T) {
	// TEST-TUI-35: all wizard steps are numbered out of 8
	cases := map[Step]string{
		StepProjectName:  "Step 1/8",
		StepDocker:       "Step 2/8",
		StepDependencies: "Step 3/8",
		StepDataBase:     "Step 4/8",
		StepCacheBackend: "Step 5/8",
		StepAppsCreated:  "Step 6/8",
		StepTemplates:    "Step 7/8",
		StepConfirm:      "Step 8/8",
	}
	for step, badge := range cases {
		m := NewModel()
		m.step = step
		if !strings.Contains(m.View().Content, badge) {
			t.Errorf("TEST-TUI-35: step %v missing badge %q", step, badge)
		}
	}
}

func TestView_DataBasesUsesDbCursor(t *testing.T) {
	// TEST-TUI-36: regression - a large depCursor must not panic or misplace
	// the database view (it used to index dbOptions with depCursor)
	m := NewModel()
	m.step = StepDataBase
	m.depCursor = 8

	view := m.View().Content
	if !strings.Contains(view, "PostgreSQL") || !strings.Contains(view, "Sqlite") {
		t.Errorf("TEST-TUI-36: expected database options rendered, got:\n%s", view)
	}
	if !strings.Contains(view, "psycopg is installed for you") {
		t.Errorf("TEST-TUI-36: expected database descriptions rendered, got:\n%s", view)
	}
}

func TestView_AppsCreated_AskPhase(t *testing.T) {
	// TEST-TUI-37: ask phase renders the yes/no question
	m := NewModel()
	m.step = StepAppsCreated
	m.appsPhase = 0

	view := m.View().Content
	if !strings.Contains(view, "Would you like to create starter apps") {
		t.Errorf("TEST-TUI-37: expected apps question, got:\n%s", view)
	}
	if !strings.Contains(view, "Yes") || !strings.Contains(view, "No") {
		t.Errorf("TEST-TUI-37: expected Yes/No options, got:\n%s", view)
	}
}

func TestView_AppsCreated_SelectPhase(t *testing.T) {
	// TEST-TUI-38: select phase renders the app checkboxes
	m := NewModel()
	m.step = StepAppsCreated
	m.appsPhase = 1

	view := m.View().Content
	for _, expected := range []string{"api", "base", "REST API endpoints app"} {
		if !strings.Contains(view, expected) {
			t.Errorf("TEST-TUI-38: expected %q in apps view, got:\n%s", expected, view)
		}
	}
}

func TestView_Templates(t *testing.T) {
	// TEST-TUI-39: per-app option hidden without apps, shown with apps
	m := NewModel()
	m.step = StepTemplates

	view := m.View().Content
	if !strings.Contains(view, "No templates folder") || !strings.Contains(view, "Project root templates/ folder") {
		t.Errorf("TEST-TUI-39: expected base template options, got:\n%s", view)
	}
	if strings.Contains(view, "templates/ in each app") {
		t.Errorf("TEST-TUI-39: per-app option must be hidden without apps, got:\n%s", view)
	}

	m.appsChecked["api"] = true
	view = m.View().Content
	if !strings.Contains(view, "templates/ in each app") {
		t.Errorf("TEST-TUI-39: expected per-app option with apps selected, got:\n%s", view)
	}
}

func TestView_Dependencies_Grid(t *testing.T) {
	// TEST-TUI-41: dependencies render in a two-column grid
	m := NewModel()
	m.step = StepDependencies

	view := m.View().Content
	found := false
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Django REST Framework") && strings.Contains(line, "Pillow") {
			found = true
		}
	}
	if !found {
		t.Errorf("TEST-TUI-41: expected the first two dependencies on one grid row, got:\n%s", view)
	}
}

func TestView_CacheBackend(t *testing.T) {
	// TEST-TUI-42: cache backend step renders the backend choices
	m := NewModel()
	m.step = StepCacheBackend

	view := m.View().Content
	for _, expected := range []string{"Dummy", "Local-memory", "File-based", "Database", "Redis"} {
		if !strings.Contains(view, expected) {
			t.Errorf("TEST-TUI-42: expected %q in cache backend view, got:\n%s", expected, view)
		}
	}
}

func TestView_ConfirmIncludesNewRows(t *testing.T) {
	// TEST-TUI-40: confirm card lists database, cache, apps, templates
	m := NewModel()
	m.step = StepConfirm
	m.cfg.ProjectName = "demo_app"
	m.depChecked["postgresql"] = true
	m.cacheChoice = 2 // File-based
	m.appsChecked["api"] = true
	m.templatesChoice = 1 // root

	view := m.View().Content
	for _, expected := range []string{
		"Database     : PostgreSQL",
		"Cache        : File-based",
		"Apps         : api",
		"Templates    : Project root templates/ folder",
	} {
		if !strings.Contains(view, expected) {
			t.Errorf("TEST-TUI-40: expected %q in confirm card, got:\n%s", expected, view)
		}
	}
}
