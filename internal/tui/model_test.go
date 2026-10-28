package tui

import (
	"strings"
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		// TEST-TUI-01: Valid Names
		{
			name:    "TEST-TUI-01: Standard lowercase name",
			input:   "myapp",
			wantErr: false,
		},
		{
			name:    "TEST-TUI-01: Underscores and digits",
			input:   "my_project_1",
			wantErr: false,
		},
		{
			name:    "TEST-TUI-01: Leading underscore",
			input:   "_internal_app",
			wantErr: false,
		},
		{
			name:    "TEST-TUI-01: Mixed case with trailing spaces",
			input:   "  MyCustomProject  ",
			wantErr: false,
		},

		// TEST-TUI-02: Empty & Whitespace
		{
			name:        "TEST-TUI-02: Empty string",
			input:       "",
			wantErr:     true,
			errContains: "project name cannot be empty",
		},
		{
			name:        "TEST-TUI-02: Whitespace only",
			input:       "   \t  ",
			wantErr:     true,
			errContains: "project name cannot be empty",
		},

		// TEST-TUI-03: Invalid Characters
		{
			name:        "TEST-TUI-03: Hyphenated name",
			input:       "my-django-app",
			wantErr:     true,
			errContains: "name must be a valid Python identifier",
		},
		{
			name:        "TEST-TUI-03: Spaces in between",
			input:       "my project",
			wantErr:     true,
			errContains: "name must be a valid Python identifier",
		},
		{
			name:        "TEST-TUI-03: Special symbols",
			input:       "app$cool",
			wantErr:     true,
			errContains: "name must be a valid Python identifier",
		},

		// TEST-TUI-04: Starting with Digit
		{
			name:        "TEST-TUI-04: Starts with number",
			input:       "1project",
			wantErr:     true,
			errContains: "name must be a valid Python identifier",
		},

		// TEST-TUI-05: Reserved Keywords
		{
			name:        "TEST-TUI-05: Reserved name django",
			input:       "django",
			wantErr:     true,
			errContains: "reserved name in Python/Django",
		},
		{
			name:        "TEST-TUI-05: Reserved name Django (uppercase)",
			input:       "Django",
			wantErr:     true,
			errContains: "reserved name in Python/Django",
		},
		{
			name:        "TEST-TUI-05: Reserved name test",
			input:       "test",
			wantErr:     true,
			errContains: "reserved name in Python/Django",
		},
		{
			name:        "TEST-TUI-05: Reserved name site",
			input:       "site",
			wantErr:     true,
			errContains: "reserved name in Python/Django",
		},
		{
			name:        "TEST-TUI-05: Reserved name os",
			input:       "os",
			wantErr:     true,
			errContains: "reserved name in Python/Django",
		},
		{
			name:        "TEST-TUI-05: Reserved name sys",
			input:       "sys",
			wantErr:     true,
			errContains: "reserved name in Python/Django",
		},
		{
			name:        "TEST-TUI-05: Reserved name admin",
			input:       "admin",
			wantErr:     true,
			errContains: "reserved name in Python/Django",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("validateProjectName(%q) expected error, got nil", tt.input)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateProjectName(%q) error = %q; want it to contain %q", tt.input, err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Fatalf("validateProjectName(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestNewModel(t *testing.T) {
	// TEST-TUI-06: NewModel Default State
	m := NewModel()

	if m.step != StepProjectName {
		t.Errorf("TEST-TUI-06: expected initial step to be StepProjectName, got %v", m.step)
	}

	if m.dockerChoice != 0 {
		t.Errorf("TEST-TUI-06: expected default dockerChoice to be 0 (Yes), got %d", m.dockerChoice)
	}

	if len(m.depOptions) != 9 {
		t.Errorf("TEST-TUI-06: expected 9 dependency options, got %d", len(m.depOptions))
	}

	expectedIDs := []string{
		"djangorestframework",
		"pillow",
		"django-silk",
		"djangorestframework-simplejwt",
		"django-filter",
		"django-cors-headers",
		"django-environ",
		"django-debug-toolbar",
		"django-modeltranslation",
	}
	for i, id := range expectedIDs {
		if m.depOptions[i].ID != id {
			t.Errorf("TEST-TUI-06: depOption[%d] expected ID %q, got %q", i, id, m.depOptions[i].ID)
		}
	}

	if m.depCursor != 0 {
		t.Errorf("TEST-TUI-06: expected depCursor = 0, got %d", m.depCursor)
	}

	if m.depChecked == nil {
		t.Errorf("TEST-TUI-06: expected depChecked map to be initialized, got nil")
	}

	if len(m.dbOptions) != 2 {
		t.Errorf("TEST-TUI-06: expected 2 database options, got %d", len(m.dbOptions))
	}
	if m.dbCursor != 0 {
		t.Errorf("TEST-TUI-06: expected dbCursor = 0, got %d", m.dbCursor)
	}

	if len(m.appOptions) != 2 {
		t.Errorf("TEST-TUI-06: expected 2 app options, got %d", len(m.appOptions))
	}
	if m.appOptions[0].ID != "api" || m.appOptions[1].ID != "base" {
		t.Errorf("TEST-TUI-06: expected app options api and base, got %v", m.appOptions)
	}
	if m.appsCursor != 0 {
		t.Errorf("TEST-TUI-06: expected appsCursor = 0, got %d", m.appsCursor)
	}
	if m.appsChecked == nil {
		t.Errorf("TEST-TUI-06: expected appsChecked map to be initialized, got nil")
	}
	if m.appsChoice != 0 {
		t.Errorf("TEST-TUI-06: expected appsChoice = 0 (No), got %d", m.appsChoice)
	}
	if m.appsPhase != 0 {
		t.Errorf("TEST-TUI-06: expected appsPhase = 0 (ask), got %d", m.appsPhase)
	}
	if m.templatesChoice != 0 {
		t.Errorf("TEST-TUI-06: expected templatesChoice = 0, got %d", m.templatesChoice)
	}

	// Verify Init command is non-nil
	cmd := m.Init()
	if cmd == nil {
		t.Errorf("TEST-TUI-06: expected non-nil Init cmd (Blink)")
	}
}

func TestTemplateOptions(t *testing.T) {
	m := NewModel()

	// TEST-TUI-27: per-app option only appears when apps are selected
	opts := m.templateOptions()
	if len(opts) != 2 {
		t.Fatalf("TEST-TUI-27: expected 2 template options without apps, got %d", len(opts))
	}
	if opts[0].ID != "none" || opts[1].ID != "root" {
		t.Errorf("TEST-TUI-27: expected none and root options, got %v", opts)
	}

	m.appsChecked["api"] = true
	opts = m.templateOptions()
	if len(opts) != 3 {
		t.Fatalf("TEST-TUI-27: expected 3 template options with apps, got %d", len(opts))
	}
	if opts[2].ID != "perapp" {
		t.Errorf("TEST-TUI-27: expected perapp as third option, got %v", opts[2])
	}
}
