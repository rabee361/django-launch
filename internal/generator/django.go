package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rabee361/django-launch/internal/config"
)

var (
	reInstalledApps    = regexp.MustCompile(`(INSTALLED_APPS\s*=\s*\[)`)
	reMiddleware       = regexp.MustCompile(`(MIDDLEWARE\s*=\s*\[)`)
	reUrlPatterns      = regexp.MustCompile(`(urlpatterns\s*=\s*\[)`)
	reDjangoUrlsImport = regexp.MustCompile(`(?m)^from django\.urls import (.+)$`)
	reAdminImport      = regexp.MustCompile(`(?m)^from django\.contrib import admin`)
	rePathlibImport    = regexp.MustCompile(`(?m)^from pathlib import Path`)
	reBaseDir          = regexp.MustCompile(`(?m)^BASE_DIR\s*=.*$`)
	reSecretKey        = regexp.MustCompile(`(?m)^SECRET_KEY\s*=\s*['"]([^'"]+)['"]\s*$`)
	reDatabasesBlock   = regexp.MustCompile(`(?ms)^DATABASES\s*=\s*\{.*?^\}`)
)

func StartDjangoProject(ctx context.Context, cfg *config.ProjectConfig) error {
	pythonPath := GetVenvPython(cfg.OutputDir)
	cmd := exec.CommandContext(ctx, pythonPath, "-m", "django", "startproject", cfg.ProjectName, ".")
	cmd.Dir = cfg.OutputDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run django startproject: %w, output: %s", err, string(out))
	}
	return nil
}

// ScaffoldSettingsAndUrls patches settings.py and urls.py for the chosen
// dependencies and returns the SECRET_KEY extracted from the stock settings
// so it can be persisted into a generated .env file.
func ScaffoldSettingsAndUrls(cfg *config.ProjectConfig) (string, error) {
	settingsPath := filepath.Join(cfg.OutputDir, cfg.ProjectName, "settings.py")
	urlsPath := filepath.Join(cfg.OutputDir, cfg.ProjectName, "urls.py")

	secretKey, err := patchSettings(settingsPath, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to patch settings.py: %w", err)
	}

	if err := patchUrls(urlsPath, cfg); err != nil {
		return "", fmt.Errorf("failed to patch urls.py: %w", err)
	}

	return secretKey, nil
}

func patchSettings(settingsPath string, cfg *config.ProjectConfig) (string, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return "", err
	}
	content := string(data)

	useEnviron := cfg.HasDependency("django-environ")

	secretKey := ""
	if loc := reSecretKey.FindStringSubmatch(content); len(loc) > 1 {
		secretKey = loc[1]
	}
	if useEnviron && secretKey == "" {
		return "", fmt.Errorf("django-environ is selected but no SECRET_KEY could be found in settings.py")
	}

	// INSTALLED_APPS
	var newApps []string
	if cfg.HasDependency("djangorestframework") {
		newApps = append(newApps, "    'rest_framework',")
	}
	if cfg.HasDependency("django-silk") {
		newApps = append(newApps, "    'silk',")
	}
	if cfg.HasDependency("django-filter") {
		newApps = append(newApps, "    'django_filters',")
	}
	if cfg.HasDependency("django-debug-toolbar") {
		newApps = append(newApps, "    'debug_toolbar',")
	}
	if cfg.HasDependency("django-cors-headers") {
		newApps = append(newApps, "    'corsheaders',")
	}
	if cfg.HasDependency("djangorestframework-simplejwt") {
		newApps = append(newApps, "    'rest_framework_simplejwt',")
	}
	if cfg.HasDependency("django-modeltranslation") {
		newApps = append(newApps, "    'modeltranslation',")
	}

	if len(newApps) > 0 && reInstalledApps.MatchString(content) {
		replacement := "${1}\n" + strings.Join(newApps, "\n")
		content = reInstalledApps.ReplaceAllString(content, replacement)
	}

	// MIDDLEWARE for packages
	if cfg.HasDependency("django-silk") && reMiddleware.MatchString(content) {
		replacement := "${1}\n    'silk.middleware.SilkyMiddleware',"
		content = reMiddleware.ReplaceAllString(content, replacement)
	}

	if cfg.HasDependency("django-debug-toolbar") && reMiddleware.MatchString(content) {
		replacement := "${1}\n    'debug_toolbar.middleware.DebugToolbarMiddleware',"
		content = reMiddleware.ReplaceAllString(content, replacement)
	}

	// MEDIA_URL and MEDIA_ROOT
	if cfg.HasDependency("pillow") {
		content += `
# Media files (Uploaded files)
MEDIA_URL = '/media/'
MEDIA_ROOT = BASE_DIR / 'media'
`
	}

	// DATABASES block
	if block := buildDatabasesBlock(cfg); block != "" {
		if reDatabasesBlock.MatchString(content) {
			content = reDatabasesBlock.ReplaceAllLiteralString(content, block)
		} else {
			content += "\n" + block
		}
	}

	// Imports required by the blocks above
	needsOs := useEnviron || len(cfg.Databases) > 0
	content = addSettingsImports(content, needsOs, useEnviron)

	if useEnviron {
		// Load .env into the environment right after BASE_DIR is defined
		content = reBaseDir.ReplaceAllString(content, "$0\n\nenviron.Env.read_env()")
		// Keep the extracted key as a dev fallback, prefer the .env value
		content = reSecretKey.ReplaceAllString(content, "SECRET_KEY = os.environ.get('SECRET_KEY', '${1}')")
	}

	return secretKey, os.WriteFile(settingsPath, []byte(content), 0644)
}

// buildDatabasesBlock returns a DATABASES dict wired to environment
// variables, or "" when no supported database is selected.
func buildDatabasesBlock(cfg *config.ProjectConfig) string {
	var selected []string
	for _, db := range cfg.Databases {
		if db == "postgresql" || db == "sqlite" {
			selected = append(selected, db)
		}
	}
	if len(selected) == 0 {
		return ""
	}

	render := func(db, alias, dbName string) string {
		if db == "postgresql" {
			return fmt.Sprintf(`    '%s': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': %s,
        'USER': %s,
        'PASSWORD': %s,
        'HOST': %s,
        'PORT': %s,
    },`, alias,
				envOr("DB_NAME", dbName),
				envOr("DB_USER", "postgres"),
				envOr("DB_PASSWORD", "postgres"),
				envOr("DB_HOST", "localhost"),
				envOr("DB_PORT", "5432"))
		}
		return fmt.Sprintf(`    '%s': {
        'ENGINE': 'django.db.backends.sqlite3',
        'NAME': %s,
    },`, alias, envOr("DB_NAME", "str(BASE_DIR / '"+dbName+"')"))
	}

	defaultName, otherName := cfg.ProjectName, cfg.ProjectName+"_other"
	if selected[0] == "sqlite" {
		defaultName = "db.sqlite3"
	}
	if len(selected) > 1 && selected[1] == "sqlite" {
		otherName = "db_other.sqlite3"
	}

	var sb strings.Builder
	sb.WriteString(`DATABASES = {
`)
	sb.WriteString(render(selected[0], "default", defaultName))
	sb.WriteString("\n")
	if len(selected) > 1 && selected[1] != selected[0] {
		sb.WriteString(render(selected[1], "other", otherName))
		sb.WriteString("\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

func envOr(key, fallback string) string {
	if strings.HasPrefix(fallback, "str(") {
		return fmt.Sprintf("os.environ.get('%s', %s)", key, fallback)
	}
	return fmt.Sprintf("os.environ.get('%s', '%s')", key, fallback)
}

// addSettingsImports inserts `import os` and/or `import environ` before the
// first import in settings.py.
func addSettingsImports(content string, needsOs, needsEnviron bool) string {
	var imports []string
	if needsOs && !regexp.MustCompile(`(?m)^import os\s*$`).MatchString(content) {
		imports = append(imports, "import os")
	}
	if needsEnviron && !regexp.MustCompile(`(?m)^import environ\s*$`).MatchString(content) {
		imports = append(imports, "import environ")
	}
	if len(imports) == 0 {
		return content
	}
	block := strings.Join(imports, "\n") + "\n"

	if rePathlibImport.MatchString(content) {
		return rePathlibImport.ReplaceAllLiteralString(content, block+"from pathlib import Path")
	}
	return block + content
}

func patchUrls(urlsPath string, cfg *config.ProjectConfig) error {
	data, err := os.ReadFile(urlsPath)
	if err != nil {
		return err
	}
	content := string(data)

	needsInclude := cfg.HasDependency("django-silk") || len(cfg.Apps) > 0
	needsStatic := cfg.HasDependency("pillow")
	needsDebugToolBar := cfg.HasDependency("django-debug-toolbar")

	// 1. Update django.urls import for include
	if needsInclude {
		if loc := reDjangoUrlsImport.FindStringSubmatch(content); len(loc) > 1 {
			imported := loc[1]
			if !strings.Contains(imported, "include") {
				newImport := fmt.Sprintf("from django.urls import %s, include", strings.TrimSpace(imported))
				content = reDjangoUrlsImport.ReplaceAllLiteralString(content, newImport)
			}
		} else {
			content = "from django.urls import include\n" + content
		}
	}

	// Add static / settings imports for media files
	if needsStatic {
		staticImports := "from django.conf import settings\nfrom django.conf.urls.static import static\n"
		if reAdminImport.MatchString(content) {
			content = reAdminImport.ReplaceAllLiteralString(content, "from django.contrib import admin\n"+staticImports)
		} else {
			content = staticImports + content
		}
	}

	if needsDebugToolBar {
		debugToolBarImports := "from debug_toolbar.toolbar import debug_toolbar_urls"
		if reAdminImport.MatchString(content) {
			content = reAdminImport.ReplaceAllLiteralString(content, "from django.contrib import admin\n"+debugToolBarImports)
		} else {
			content = debugToolBarImports + content
		}
	}

	// Add routes inside urlpatterns
	routes := []string{}
	if cfg.HasDependency("django-silk") {
		routes = append(routes, "    path('silk/', include('silk.urls', namespace='silk')),")
	}
	for _, app := range cfg.Apps {
		if app != "" {
			routes = append(routes, fmt.Sprintf("    path('%s/', include('%s.urls')),", app, app))
		}
	}
	if len(routes) > 0 && reUrlPatterns.MatchString(content) {
		content = reUrlPatterns.ReplaceAllString(content, "${1}\n"+strings.Join(routes, "\n"))
	}

	// Add media serving in DEBUG mode
	if cfg.HasDependency("pillow") {
		content += `
if settings.DEBUG:
    urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
`
	}

	// Add debug tool bar config
	if cfg.HasDependency("django-debug-toolbar") {
		content += `
urlpatterns += debug_toolbar_urls()
`
	}

	return os.WriteFile(urlsPath, []byte(content), 0644)
}

const appUrlsPy = `from django.urls import path

urlpatterns = [
]
`

// CreateApps runs `manage.py startapp` for every requested app, writes a
// minimal urls.py for each, and creates the requested template folders.
func CreateApps(ctx context.Context, cfg *config.ProjectConfig) error {
	pythonPath := GetVenvPython(cfg.OutputDir)

	for _, app := range cfg.Apps {
		if app == "" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		cmd := exec.CommandContext(ctx, pythonPath, "-m", "django", "startapp", app)
		cmd.Dir = cfg.OutputDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to run django startapp %s: %w, output: %s", app, err, string(out))
		}

		appDir := filepath.Join(cfg.OutputDir, app)
		if err := os.WriteFile(filepath.Join(appDir, "urls.py"), []byte(appUrlsPy), 0644); err != nil {
			return fmt.Errorf("failed to write urls.py for app %s: %w", app, err)
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	switch cfg.Templates {
	case "root":
		if err := createTemplateDir(filepath.Join(cfg.OutputDir, "templates")); err != nil {
			return err
		}
	case "perapp":
		for _, app := range cfg.Apps {
			if app == "" {
				continue
			}
			if err := createTemplateDir(filepath.Join(cfg.OutputDir, app, "templates")); err != nil {
				return err
			}
		}
	}

	return nil
}

// createTemplateDir creates an empty template folder kept in git via .gitkeep.
func createTemplateDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory %s: %w", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), nil, 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Join(dir, ".gitkeep"), err)
	}
	return nil
}
