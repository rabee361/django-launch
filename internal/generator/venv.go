package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"main/internal/config"
)

// DetectTooling checks if uv and python are available on the machine.
func DetectTooling() (hasUv bool, pythonCmd string, err error) {
	if _, err := exec.LookPath("uv"); err == nil {
		hasUv = true
	}

	candidates := []string{"python", "python3", "py"}
	for _, cmd := range candidates {
		if path, err := exec.LookPath(cmd); err == nil {
			pythonCmd = path
			break
		}
	}

	if pythonCmd == "" && !hasUv {
		return false, "", fmt.Errorf("neither 'uv' nor 'python' was found in your PATH. Please install Python or uv to continue")
	}

	return hasUv, pythonCmd, nil
}

// GetVenvPython returns the path to the python executable inside the project's .venv.
func GetVenvPython(targetDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(targetDir, ".venv", "Scripts", "python.exe")
	}
	return filepath.Join(targetDir, ".venv", "bin", "python")
}

// GetVenvPip returns the path to the pip executable inside the project's .venv.
func GetVenvPip(targetDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(targetDir, ".venv", "Scripts", "pip.exe")
	}
	return filepath.Join(targetDir, ".venv", "bin", "pip")
}

// CreateVirtualEnv creates the .venv directory inside targetDir.
func CreateVirtualEnv(cfg *config.ProjectConfig) error {
	var cmd *exec.Cmd

	if cfg.UseUv {
		cmd = exec.Command("uv", "venv", ".venv")
	} else {
		cmd = exec.Command(cfg.PythonCmd, "-m", "venv", ".venv")
	}

	cmd.Dir = cfg.OutputDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create virtual environment: %v, output: %s", err, string(out))
	}
	return nil
}

// GenerateRequirementsTxt writes a requirements.txt file with chosen dependencies.
func GenerateRequirementsTxt(cfg *config.ProjectConfig) error {
	reqPath := filepath.Join(cfg.OutputDir, "requirements.txt")
	var lines []string

	lines = append(lines, "Django>=6.0")
	for _, dep := range cfg.Dependencies {
		trimmed := strings.TrimSpace(dep)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(reqPath, []byte(content), 0644)
}

// InstallDependencies installs requirements.txt inside the .venv.
func InstallDependencies(cfg *config.ProjectConfig) error {
	var cmd *exec.Cmd

	if cfg.UseUv {
		cmd = exec.Command("uv", "pip", "install", "-r", "requirements.txt", "--python", ".venv")
	} else {
		pipPath := GetVenvPip(cfg.OutputDir)
		if _, err := os.Stat(pipPath); err == nil {
			cmd = exec.Command(pipPath, "install", "-r", "requirements.txt")
		} else {
			// Fallback: run python -m pip
			pythonPath := GetVenvPython(cfg.OutputDir)
			cmd = exec.Command(pythonPath, "-m", "pip", "install", "-r", "requirements.txt")
		}
	}

	cmd.Dir = cfg.OutputDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %v, output: %s", err, string(out))
	}
	return nil
}
