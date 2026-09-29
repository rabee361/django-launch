package generator

import (
	"os"
	"path/filepath"

	"github.com/rabee361/django-launch/internal/config"
)

const dockerfileTemplate = `# Use official lightweight Python image
FROM python:3.12-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Install system dependencies if required
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Install python dependencies
COPY requirements.txt /app/
RUN pip install --no-cache-dir -r requirements.txt

# Copy project files
COPY . /app/

# Expose Django port
EXPOSE 8000

# Default command to run development server
CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]
`

const dockerignoreTemplate = `.venv
__pycache__/
*.py[cod]
*$py.class
*.sqlite3
db.sqlite3
media/
.git
.gitignore
.env
Dockerfile
.dockerignore
`

// GenerateDockerFiles creates Dockerfile and .dockerignore in the project directory.
func GenerateDockerFiles(cfg *config.ProjectConfig) error {
	dockerfilePath := filepath.Join(cfg.OutputDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileTemplate), 0644); err != nil {
		return err
	}

	dockerignorePath := filepath.Join(cfg.OutputDir, ".dockerignore")
	return os.WriteFile(dockerignorePath, []byte(dockerignoreTemplate), 0644)
}
