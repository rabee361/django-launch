package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rabee361/django-launch/internal/config"
)

// ProgressCallback reports progress back to the UI.
type ProgressCallback func(step int, total int, description string)

// GenerateProject orchestrates the creation of the Django project.
func GenerateProject(ctx context.Context, cfg *config.ProjectConfig, onProgress ProgressCallback) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	totalSteps := 5
	if cfg.WithDocker {
		totalSteps = 6
	}
	currentStep := 1

	if cfg.OutputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		cfg.OutputDir = filepath.Join(cwd, cfg.ProjectName)
	}

	if onProgress != nil {
		onProgress(currentStep, totalSteps, fmt.Sprintf("Creating project directory '%s'...", cfg.ProjectName))
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", cfg.OutputDir, err)
	}
	currentStep++

	toolName := "python -m venv"
	if cfg.UseUv {
		toolName = "uv venv"
	}
	if onProgress != nil {
		onProgress(currentStep, totalSteps, fmt.Sprintf("Creating virtual environment (.venv) using %s...", toolName))
	}
	if err := CreateVirtualEnv(ctx, cfg); err != nil {
		return err
	}
	currentStep++

	if onProgress != nil {
		onProgress(currentStep, totalSteps, "Writing requirements.txt and installing dependencies...")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := GenerateRequirementsTxt(cfg); err != nil {
		return fmt.Errorf("failed to create requirements.txt: %w", err)
	}
	if err := InstallDependencies(ctx, cfg); err != nil {
		return err
	}
	currentStep++

	if onProgress != nil {
		onProgress(currentStep, totalSteps, fmt.Sprintf("Initializing Django project '%s'...", cfg.ProjectName))
	}
	if err := StartDjangoProject(ctx, cfg); err != nil {
		return err
	}
	currentStep++

	if onProgress != nil {
		onProgress(currentStep, totalSteps, "Scaffolding settings.py, urls.py, apps, and environment files...")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	secretKey, err := ScaffoldSettingsAndUrls(cfg)
	if err != nil {
		return err
	}
	if err := CreateApps(ctx, cfg); err != nil {
		return err
	}
	if err := GenerateEnvFile(cfg, secretKey); err != nil {
		return err
	}
	currentStep++

	if cfg.WithDocker {
		if onProgress != nil {
			onProgress(currentStep, totalSteps, "Generating Dockerfile and .dockerignore...")
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := GenerateDockerFiles(cfg); err != nil {
			return fmt.Errorf("failed to generate Docker files: %w", err)
		}
		currentStep++
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
