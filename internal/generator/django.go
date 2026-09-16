package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"main/internal/config"
)

// StartDjangoProject runs "python -m django startproject <name> ." inside the target directory.
func StartDjangoProject(cfg *config.ProjectConfig) error {
	pythonPath := GetVenvPython(cfg.OutputDir)
	cmd := exec.Command(pythonPath, "-m", "django", "startproject", cfg.ProjectName, ".")
	cmd.Dir = cfg.OutputDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run django startproject: %v, output: %s", err, string(out))
	}
	return nil
}

// ScaffoldSettingsAndUrls modifies settings.py and urls.py according to selected dependencies.
func ScaffoldSettingsAndUrls(cfg *config.ProjectConfig) error {
	settingsPath := filepath.Join(cfg.OutputDir, cfg.ProjectName, "settings.py")
	urlsPath := filepath.Join(cfg.OutputDir, cfg.ProjectName, "urls.py")

	if err := patchSettings(settingsPath, cfg); err != nil {
		return fmt.Errorf("failed to patch settings.py: %w", err)
	}

	if err := patchUrls(urlsPath, cfg); err != nil {
		return fmt.Errorf("failed to patch urls.py: %w", err)
	}

	return nil
}

func patchSettings(settingsPath string, cfg *config.ProjectConfig) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}
	content := string(data)

	// 1. INSTALLED_APPS
	var newApps []string
	if cfg.HasDependency("djangorestframework") {
		newApps = append(newApps, "    'rest_framework',")
	}
	if cfg.HasDependency("django-silk") {
		newApps = append(newApps, "    'silk',")
	}
	if cfg.HasDependency("djangorestframework-simplejwt") {
		newApps = append(newApps, "    'rest_framework_simplejwt',")
	}

	if len(newApps) > 0 {
		reApps := regexp.MustCompile(`(INSTALLED_APPS\s*=\s*\[)`)
		if reApps.MatchString(content) {
			replacement := "${1}\n" + strings.Join(newApps, "\n")
			content = reApps.ReplaceAllString(content, replacement)
		}
	}

	// 2. MIDDLEWARE for Silk
	if cfg.HasDependency("django-silk") {
		reMiddleware := regexp.MustCompile(`(MIDDLEWARE\s*=\s*\[)`)
		if reMiddleware.MatchString(content) {
			silkMiddleware := "    'silk.middleware.SilkyMiddleware',"
			replacement := "${1}\n" + silkMiddleware
			content = reMiddleware.ReplaceAllString(content, replacement)
		}
	}

	// 3. MEDIA_URL and MEDIA_ROOT for Pillow
	if cfg.HasDependency("pillow") {
		mediaConfig := `
# Media files (Uploaded files)
MEDIA_URL = '/media/'
MEDIA_ROOT = BASE_DIR / 'media'
`
		content += mediaConfig
	}

	return os.WriteFile(settingsPath, []byte(content), 0644)
}

func patchUrls(urlsPath string, cfg *config.ProjectConfig) error {
	data, err := os.ReadFile(urlsPath)
	if err != nil {
		return err
	}
	content := string(data)

	needsInclude := cfg.HasDependency("django-silk")
	needsStatic := cfg.HasDependency("pillow")
	// needsRestFrameworkSetup := cfg.HasDependency("djangorestframework")

	// 1. Update django.urls import for include
	if needsInclude {
		reDjangoUrls := regexp.MustCompile(`(?m)^from django\.urls import (.+)$`)
		if loc := reDjangoUrls.FindStringSubmatch(content); len(loc) > 1 {
			imported := loc[1]
			if !strings.Contains(imported, "include") {
				newImport := fmt.Sprintf("from django.urls import %s, include", strings.TrimSpace(imported))
				content = reDjangoUrls.ReplaceAllLiteralString(content, newImport)
			}
		} else {
			content = "from django.urls import include\n" + content
		}
	}

	// 2. Add static / settings imports for media files
	if needsStatic {
		reAdminImport := regexp.MustCompile(`(?m)^from django\.contrib import admin`)
		staticImports := "from django.conf import settings\nfrom django.conf.urls.static import static\n"
		if reAdminImport.MatchString(content) {
			content = reAdminImport.ReplaceAllLiteralString(content, "from django.contrib import admin\n"+staticImports)
		} else {
			content = staticImports + content
		}
	}

	// 3. Add routes inside urlpatterns
	if cfg.HasDependency("django-silk") {
		silkRoute := "    path('silk/', include('silk.urls', namespace='silk')),"
		rePatterns := regexp.MustCompile(`(urlpatterns\s*=\s*\[)`)
		if rePatterns.MatchString(content) {
			content = rePatterns.ReplaceAllString(content, "${1}\n"+silkRoute)
		}
	}

	// 4. Add media serving in DEBUG mode
	if cfg.HasDependency("pillow") {
		mediaPatterns := `
if settings.DEBUG:
    urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
`
		content += mediaPatterns
	}

	return os.WriteFile(urlsPath, []byte(content), 0644)
}
