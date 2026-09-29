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

	// Ensure OutputDir is set
	if cfg.OutputDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		cfg.OutputDir = filepath.Join(cwd, cfg.ProjectName)
	}

	// 1. Create directory
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

	// 2. Create virtual environment
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

	// 3. Write requirements.txt & install dependencies
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

	// 4. Start Django project
	if onProgress != nil {
		onProgress(currentStep, totalSteps, fmt.Sprintf("Initializing Django project '%s'...", cfg.ProjectName))
	}
	if err := StartDjangoProject(ctx, cfg); err != nil {
		return err
	}
	currentStep++

	// 5. Deep scaffold settings and urls
	if onProgress != nil {
		onProgress(currentStep, totalSteps, "Scaffolding settings.py and urls.py with chosen dependencies...")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ScaffoldSettingsAndUrls(cfg); err != nil {
		return err
	}
	currentStep++

	// 6. Docker (if requested)
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
