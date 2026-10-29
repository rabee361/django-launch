package generator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rabee361/django-launch/internal/config"
)

func TestGenerateEnvFile_SkipsWithoutEnviron(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.ProjectConfig{
		OutputDir:   dir,
		ProjectName: "demo",
		Dependencies: []string{
			"djangorestframework",
		},
	}

	if err := GenerateEnvFile(cfg, "some-secret"); err != nil {
		t.Fatalf("GenerateEnvFile failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".env")); !os.IsNotExist(err) {
		t.Errorf("expected no .env when django-environ is not selected")
	}
}

func TestGenerateEnvFile_SecretKeyAndDatabases(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.ProjectConfig{
		OutputDir:    dir,
		ProjectName:  "demo_app",
		Dependencies: []string{"django-environ"},
		Databases:    []string{"postgresql"},
	}

	if err := GenerateEnvFile(cfg, "django-insecure-sample"); err != nil {
		t.Fatalf("GenerateEnvFile failed: %v", err)
	}

	contentBytes, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatalf(".env missing: %v", err)
	}
	content := string(contentBytes)

	for _, expected := range []string{
		"SECRET_KEY=django-insecure-sample",
		"DB_NAME=demo_app",
		"DB_USER=postgres",
		"DB_PASSWORD=postgres",
		"DB_HOST=localhost",
		"DB_PORT=5432",
	} {
		if !strings.Contains(content, expected) {
			t.Errorf("expected %q in .env, got:\n%s", expected, content)
		}
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(dir, ".env"))
		if err != nil {
			t.Fatalf("failed to stat .env: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("expected .env mode 0600, got %o", perm)
		}
	}

	gitignoreBytes, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore missing: %v", err)
	}
	if !strings.Contains(string(gitignoreBytes), ".env") {
		t.Errorf("expected .env in .gitignore, got:\n%s", gitignoreBytes)
	}
}

func TestGenerateEnvFile_SqliteDatabase(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.ProjectConfig{
		OutputDir:    dir,
		ProjectName:  "demo_app",
		Dependencies: []string{"django-environ"},
		Databases:    []string{"sqlite"},
	}

	if err := GenerateEnvFile(cfg, "secret"); err != nil {
		t.Fatalf("GenerateEnvFile failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(filepath.Join(dir, ".env"))

	if !strings.Contains(string(contentBytes), "DB_NAME=db.sqlite3") {
		t.Errorf("expected sqlite DB_NAME in .env, got:\n%s", contentBytes)
	}
	if strings.Contains(string(contentBytes), "DB_USER=") {
		t.Errorf("unexpected DB_USER for sqlite in .env, got:\n%s", contentBytes)
	}
}

func TestGenerateEnvFile_RedisCache(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.ProjectConfig{
		OutputDir:    dir,
		Dependencies: []string{"django-environ"},
		CacheBackend: "redis",
	}

	if err := GenerateEnvFile(cfg, "secret"); err != nil {
		t.Fatalf("GenerateEnvFile failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(filepath.Join(dir, ".env"))

	if !strings.Contains(string(contentBytes), "REDIS_URL=redis://127.0.0.1:6379/1") {
		t.Errorf("expected REDIS_URL in .env, got:\n%s", contentBytes)
	}
}

func TestGenerateEnvFile_DoesNotOverwriteGitignore(t *testing.T) {
	dir := t.TempDir()
	existing := "node_modules/\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(existing), 0644); err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}

	cfg := &config.ProjectConfig{
		OutputDir:    dir,
		Dependencies: []string{"django-environ"},
	}

	if err := GenerateEnvFile(cfg, "secret"); err != nil {
		t.Fatalf("GenerateEnvFile failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if string(contentBytes) != existing {
		t.Errorf("expected existing .gitignore untouched, got:\n%s", contentBytes)
	}
}
