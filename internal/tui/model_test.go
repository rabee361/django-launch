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

	if len(m.depOptions) != 4 {
		t.Errorf("TEST-TUI-06: expected 4 dependency options, got %d", len(m.depOptions))
	}

	expectedIDs := []string{"djangorestframework", "pillow", "django-silk", "djangorestframework-simplejwt","django-cors-headers", "django-debug-toolbar","django-filter"}
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

	// Verify Init command is non-nil
	cmd := m.Init()
	if cmd == nil {
		t.Errorf("TEST-TUI-06: expected non-nil Init cmd (Blink)")
	}
}
