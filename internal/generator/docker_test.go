package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"main/internal/config"
)

func TestGenerateDockerFiles(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.ProjectConfig{
		OutputDir: tempDir,
	}

	err := GenerateDockerFiles(cfg)
	if err != nil {
		t.Fatalf("GenerateDockerFiles failed: %v", err)
	}

	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	dockerignorePath := filepath.Join(tempDir, ".dockerignore")

	// TEST-GEN-03: File existence and permissions
	dfInfo, err := os.Stat(dockerfilePath)
	if err != nil {
		t.Fatalf("TEST-GEN-03: Dockerfile was not created: %v", err)
	}
	if dfInfo.IsDir() {
		t.Errorf("TEST-GEN-03: Dockerfile should be a regular file, got directory")
	}

	diInfo, err := os.Stat(dockerignorePath)
	if err != nil {
		t.Fatalf("TEST-GEN-03: .dockerignore was not created: %v", err)
	}
	if diInfo.IsDir() {
		t.Errorf("TEST-GEN-03: .dockerignore should be a regular file, got directory")
	}

	// TEST-GEN-04: Dockerfile content verification
	dfContent, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("failed to read Dockerfile: %v", err)
	}
	dfStr := string(dfContent)

	expectedDirectives := []string{
		"FROM python:3.12-slim",
		"PYTHONDONTWRITEBYTECODE=1",
		"PYTHONUNBUFFERED=1",
		"WORKDIR /app",
		"apt-get update",
		"COPY requirements.txt /app/",
		"RUN pip install --no-cache-dir -r requirements.txt",
		"COPY . /app/",
		"EXPOSE 8000",
		`CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]`,
	}

	for _, dir := range expectedDirectives {
		if !strings.Contains(dfStr, dir) {
			t.Errorf("TEST-GEN-04: Dockerfile missing directive: %q", dir)
		}
	}

	// TEST-GEN-04: .dockerignore content verification
	diContent, err := os.ReadFile(dockerignorePath)
	if err != nil {
		t.Fatalf("failed to read .dockerignore: %v", err)
	}
	diStr := string(diContent)

	expectedIgnoreEntries := []string{
		".venv",
		"__pycache__/",
		"*.py[cod]",
		"db.sqlite3",
		"media/",
		".git",
		"Dockerfile",
		".dockerignore",
	}

	for _, entry := range expectedIgnoreEntries {
		if !strings.Contains(diStr, entry) {
			t.Errorf("TEST-GEN-04: .dockerignore missing entry: %q", entry)
		}
	}
}

func TestGenerateDockerFiles_InvalidDir(t *testing.T) {
	// Provide a non-existent nested directory without creating parents
	invalidDir := filepath.Join(t.TempDir(), "non", "existent", "dir")
	cfg := &config.ProjectConfig{
		OutputDir: invalidDir,
	}

	err := GenerateDockerFiles(cfg)
	if err == nil {
		t.Errorf("GenerateDockerFiles expected error for non-existent directory, got nil")
	}
}
