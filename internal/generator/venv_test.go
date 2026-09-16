package generator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"main/internal/config"
)

func TestGenerateRequirementsTxt_Base(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.ProjectConfig{
		OutputDir:    tempDir,
		Dependencies: []string{},
	}

	err := GenerateRequirementsTxt(cfg)
	if err != nil {
		t.Fatalf("GenerateRequirementsTxt failed: %v", err)
	}

	reqPath := filepath.Join(tempDir, "requirements.txt")
	contentBytes, err := os.ReadFile(reqPath)
	if err != nil {
		t.Fatalf("failed to read requirements.txt: %v", err)
	}

	content := string(contentBytes)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	// TEST-GEN-01: Django>=5.0 must always be the first dependency
	if len(lines) != 1 {
		t.Errorf("TEST-GEN-01: expected 1 line in requirements.txt, got %d (lines: %v)", len(lines), lines)
	}
	if lines[0] != "Django>=5.0" {
		t.Errorf("TEST-GEN-01: expected 'Django>=5.0', got %q", lines[0])
	}
}

func TestGenerateRequirementsTxt_CustomDeps(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.ProjectConfig{
		OutputDir: tempDir,
		Dependencies: []string{
			"djangorestframework",
			"  pillow  ",
			"",
			"   ",
			"django-silk>=0.5.0",
		},
	}

	err := GenerateRequirementsTxt(cfg)
	if err != nil {
		t.Fatalf("GenerateRequirementsTxt failed: %v", err)
	}

	reqPath := filepath.Join(tempDir, "requirements.txt")
	contentBytes, err := os.ReadFile(reqPath)
	if err != nil {
		t.Fatalf("failed to read requirements.txt: %v", err)
	}

	content := string(contentBytes)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	// TEST-GEN-02: Trimmed dependencies, no blank lines
	expected := []string{
		"Django>=5.0",
		"djangorestframework",
		"pillow",
		"django-silk>=0.5.0",
	}

	if len(lines) != len(expected) {
		t.Fatalf("TEST-GEN-02: expected %d lines, got %d (lines: %v)", len(expected), len(lines), lines)
	}

	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("TEST-GEN-02: line %d expected %q, got %q", i, exp, lines[i])
		}
	}
}

func TestGetVenvPython(t *testing.T) {
	targetDir := filepath.Join("test", "my_app")
	got := GetVenvPython(targetDir)

	// TEST-GEN-05: Path resolution based on OS
	if runtime.GOOS == "windows" {
		expected := filepath.Join(targetDir, ".venv", "Scripts", "python.exe")
		if got != expected {
			t.Errorf("TEST-GEN-05: Windows GetVenvPython = %q; want %q", got, expected)
		}
	} else {
		expected := filepath.Join(targetDir, ".venv", "bin", "python")
		if got != expected {
			t.Errorf("TEST-GEN-05: Unix GetVenvPython = %q; want %q", got, expected)
		}
	}
}

func TestGetVenvPip(t *testing.T) {
	targetDir := filepath.Join("test", "my_app")
	got := GetVenvPip(targetDir)

	// TEST-GEN-06: Pip path resolution based on OS
	if runtime.GOOS == "windows" {
		expected := filepath.Join(targetDir, ".venv", "Scripts", "pip.exe")
		if got != expected {
			t.Errorf("TEST-GEN-06: Windows GetVenvPip = %q; want %q", got, expected)
		}
	} else {
		expected := filepath.Join(targetDir, ".venv", "bin", "pip")
		if got != expected {
			t.Errorf("TEST-GEN-06: Unix GetVenvPip = %q; want %q", got, expected)
		}
	}
}

func TestDetectTooling(t *testing.T) {
	// TEST-GEN-07: DetectTooling should detect tools or return descriptive error
	hasUv, pyCmd, err := DetectTooling()
	if err != nil {
		// If error is returned, ensure both hasUv is false and pyCmd is empty
		if hasUv || pyCmd != "" {
			t.Errorf("TEST-GEN-07: expected hasUv=false and pyCmd=\"\" on error, got hasUv=%v, pyCmd=%q", hasUv, pyCmd)
		}
		if !strings.Contains(err.Error(), "neither 'uv' nor 'python' was found") {
			t.Errorf("TEST-GEN-07: unexpected error message: %v", err)
		}
	} else {
		// Success case: at least one of uv or python must be detected
		if !hasUv && pyCmd == "" {
			t.Errorf("TEST-GEN-07: expected either uv or python to be found when err is nil")
		}
		if pyCmd != "" {
			if _, statErr := os.Stat(pyCmd); statErr != nil {
				t.Errorf("TEST-GEN-07: detected pythonCmd %q does not exist: %v", pyCmd, statErr)
			}
		}
	}
}
