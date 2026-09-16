package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"main/internal/config"
)

func TestGenerateProjectEndToEnd(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "django_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectName := "sample_demo_app"
	projectDir := filepath.Join(tempDir, projectName)

	hasUv, pyCmd, err := DetectTooling()
	if err != nil {
		t.Fatalf("tooling detection error: %v", err)
	}

	cfg := config.ProjectConfig{
		ProjectName:  projectName,
		OutputDir:    projectDir,
		WithDocker:   true,
		Dependencies: []string{"djangorestframework", "pillow", "django-silk", "djangorestframework-simplejwt"},
		UseUv:        hasUv,
		PythonCmd:    pyCmd,
	}

	err = GenerateProject(&cfg, func(step, total int, desc string) {
		t.Logf("[%d/%d] %s", step, total, desc)
	})
	if err != nil {
		t.Fatalf("GenerateProject failed: %v", err)
	}

	// 1. Check requirements.txt
	reqBytes, err := os.ReadFile(filepath.Join(projectDir, "requirements.txt"))
	if err != nil {
		t.Fatalf("requirements.txt missing: %v", err)
	}
	reqContent := string(reqBytes)
	for _, dep := range []string{"Django", "djangorestframework", "pillow", "django-silk", "djangorestframework-simplejwt"} {
		if !strings.Contains(reqContent, dep) {
			t.Errorf("expected %s in requirements.txt", dep)
		}
	}

	// 2. Check Dockerfile and .dockerignore
	if _, err := os.Stat(filepath.Join(projectDir, "Dockerfile")); err != nil {
		t.Errorf("Dockerfile missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".dockerignore")); err != nil {
		t.Errorf(".dockerignore missing: %v", err)
	}

	// 3. Check settings.py
	settingsBytes, err := os.ReadFile(filepath.Join(projectDir, projectName, "settings.py"))
	if err != nil {
		t.Fatalf("settings.py missing: %v", err)
	}
	settingsContent := string(settingsBytes)
	if !strings.Contains(settingsContent, "'rest_framework'") {
		t.Errorf("settings.py missing 'rest_framework'")
	}
	if !strings.Contains(settingsContent, "'silk'") {
		t.Errorf("settings.py missing 'silk'")
	}
	if !strings.Contains(settingsContent, "'silk.middleware.SilkyMiddleware'") {
		t.Errorf("settings.py missing 'silk.middleware.SilkyMiddleware'")
	}
	if !strings.Contains(settingsContent, "MEDIA_URL") || !strings.Contains(settingsContent, "MEDIA_ROOT") {
		t.Errorf("settings.py missing MEDIA configuration for Pillow")
	}

	// 4. Check urls.py
	urlsBytes, err := os.ReadFile(filepath.Join(projectDir, projectName, "urls.py"))
	if err != nil {
		t.Fatalf("urls.py missing: %v", err)
	}
	urlsContent := string(urlsBytes)
	if !strings.Contains(urlsContent, "path('silk/', include('silk.urls'") {
		t.Errorf("urls.py missing Silk route")
	}
	if !strings.Contains(urlsContent, "urlpatterns += static(settings.MEDIA_URL") {
		t.Errorf("urls.py missing media static URL route for Pillow")
	}

	// 5. Test Django syntax / integrity using "python manage.py check"
	pythonExe := GetVenvPython(projectDir)
	checkCmd := exec.Command(pythonExe, "manage.py", "check")
	checkCmd.Dir = projectDir
	out, err := checkCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("django check failed: %v, output:\n%s", err, string(out))
	}
	t.Logf("Django check passed:\n%s", string(out))
}
