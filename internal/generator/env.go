package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rabee361/django-launch/internal/config"
)

const gitignoreTemplate = `.env
.venv/
__pycache__/
*.py[cod]
db.sqlite3
db_other.sqlite3
media/
`

// GenerateEnvFile writes a .env file (mode 0600) plus a .gitignore when
// django-environ is selected, so secrets stay out of version control.
func GenerateEnvFile(cfg *config.ProjectConfig, secretKey string) error {
	if !cfg.HasDependency("django-environ") {
		return nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("SECRET_KEY=%s", secretKey))
	lines = append(lines, databaseEnvLines(cfg)...)
	lines = append(lines, cacheEnvLines(cfg)...)

	envPath := filepath.Join(cfg.OutputDir, ".env")
	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to write .env: %w", err)
	}

	gitignorePath := filepath.Join(cfg.OutputDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(gitignoreTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write .gitignore: %w", err)
		}
	}
	return nil
}

func databaseEnvLines(cfg *config.ProjectConfig) []string {
	for _, db := range cfg.Databases {
		switch db {
		case "postgresql":
			return []string{
				fmt.Sprintf("DB_NAME=%s", cfg.ProjectName),
				"DB_USER=postgres",
				"DB_PASSWORD=postgres",
				"DB_HOST=localhost",
				"DB_PORT=5432",
			}
		case "sqlite":
			return []string{
				"DB_NAME=db.sqlite3",
			}
		}
	}
	return nil
}

func cacheEnvLines(cfg *config.ProjectConfig) []string {
	if cfg.CacheBackend == "redis" {
		return []string{"REDIS_URL=redis://127.0.0.1:6379/1"}
	}
	return nil
}
