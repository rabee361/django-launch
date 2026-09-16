package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"main/internal/config"
)

const sampleSettingsPy = `
INSTALLED_APPS = [
    'django.contrib.admin',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.messages',
    'django.contrib.staticfiles',
]

MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.common.CommonMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
]
`

const sampleUrlsPy = `from django.contrib import admin
from django.urls import path

urlpatterns = [
    path('admin/', admin.site.urls),
]
`

func writeSampleFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write sample file: %v", err)
	}
}

func TestPatchSettings_DRFOnly(t *testing.T) {
	// TEST-GEN-08: DRF Only
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.py")
	writeSampleFile(t, settingsPath, sampleSettingsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"djangorestframework"},
	}

	if err := patchSettings(settingsPath, cfg); err != nil {
		t.Fatalf("patchSettings failed: %v", err)
	}

	contentBytes, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.py: %v", err)
	}
	content := string(contentBytes)

	if !strings.Contains(content, "'rest_framework',") {
		t.Errorf("TEST-GEN-08: expected 'rest_framework', in INSTALLED_APPS")
	}
	if strings.Contains(content, "'silk',") {
		t.Errorf("TEST-GEN-08: unexpected 'silk' in INSTALLED_APPS")
	}
	if strings.Contains(content, "MEDIA_URL") {
		t.Errorf("TEST-GEN-08: unexpected MEDIA_URL in settings")
	}
}

func TestPatchSettings_SilkOnly(t *testing.T) {
	// TEST-GEN-09: Silk Only
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.py")
	writeSampleFile(t, settingsPath, sampleSettingsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"django-silk"},
	}

	if err := patchSettings(settingsPath, cfg); err != nil {
		t.Fatalf("patchSettings failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(settingsPath)
	content := string(contentBytes)

	if !strings.Contains(content, "'silk',") {
		t.Errorf("TEST-GEN-09: expected 'silk' in INSTALLED_APPS")
	}
	if !strings.Contains(content, "'silk.middleware.SilkyMiddleware',") {
		t.Errorf("TEST-GEN-09: expected 'silk.middleware.SilkyMiddleware' in MIDDLEWARE")
	}
}

func TestPatchSettings_SimpleJWT(t *testing.T) {
	// TEST-GEN-10: SimpleJWT
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.py")
	writeSampleFile(t, settingsPath, sampleSettingsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"djangorestframework-simplejwt"},
	}

	if err := patchSettings(settingsPath, cfg); err != nil {
		t.Fatalf("patchSettings failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(settingsPath)
	content := string(contentBytes)

	if !strings.Contains(content, "'rest_framework_simplejwt',") {
		t.Errorf("TEST-GEN-10: expected 'rest_framework_simplejwt' in INSTALLED_APPS")
	}
}

func TestPatchSettings_Pillow(t *testing.T) {
	// TEST-GEN-11: Pillow
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.py")
	writeSampleFile(t, settingsPath, sampleSettingsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"pillow"},
	}

	if err := patchSettings(settingsPath, cfg); err != nil {
		t.Fatalf("patchSettings failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(settingsPath)
	content := string(contentBytes)

	if !strings.Contains(content, "MEDIA_URL = '/media/'") {
		t.Errorf("TEST-GEN-11: expected MEDIA_URL in settings")
	}
	if !strings.Contains(content, "MEDIA_ROOT = BASE_DIR / 'media'") {
		t.Errorf("TEST-GEN-11: expected MEDIA_ROOT in settings")
	}
}

func TestPatchSettings_AllDependenciesCombined(t *testing.T) {
	// TEST-GEN-12: All Dependencies Combined
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.py")
	writeSampleFile(t, settingsPath, sampleSettingsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"djangorestframework", "django-silk", "djangorestframework-simplejwt", "pillow"},
	}

	if err := patchSettings(settingsPath, cfg); err != nil {
		t.Fatalf("patchSettings failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(settingsPath)
	content := string(contentBytes)

	for _, expected := range []string{
		"'rest_framework',",
		"'silk',",
		"'rest_framework_simplejwt',",
		"'silk.middleware.SilkyMiddleware',",
		"MEDIA_URL = '/media/'",
		"MEDIA_ROOT = BASE_DIR / 'media'",
	} {
		if !strings.Contains(content, expected) {
			t.Errorf("TEST-GEN-12: missing expected string %q", expected)
		}
	}
}

func TestPatchSettings_ErrorHandling(t *testing.T) {
	// TEST-GEN-13: Error on non-existent file
	cfg := &config.ProjectConfig{
		Dependencies: []string{"djangorestframework"},
	}
	err := patchSettings(filepath.Join(t.TempDir(), "non_existent.py"), cfg)
	if err == nil {
		t.Errorf("TEST-GEN-13: expected error for non-existent file, got nil")
	}
}

func TestPatchUrls_SilkRouteInjection(t *testing.T) {
	// TEST-GEN-14: Silk Route Injection
	dir := t.TempDir()
	urlsPath := filepath.Join(dir, "urls.py")
	writeSampleFile(t, urlsPath, sampleUrlsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"django-silk"},
	}

	if err := patchUrls(urlsPath, cfg); err != nil {
		t.Fatalf("patchUrls failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(urlsPath)
	content := string(contentBytes)

	if !strings.Contains(content, "from django.urls import path, include") {
		t.Errorf("TEST-GEN-14: expected import include in urls.py, got:\n%s", content)
	}
	if !strings.Contains(content, "path('silk/', include('silk.urls', namespace='silk')),") {
		t.Errorf("TEST-GEN-14: expected silk url route in urlpatterns, got:\n%s", content)
	}
}

func TestPatchUrls_ExistingInclude(t *testing.T) {
	// TEST-GEN-15: Existing include import should not be duplicated
	urlsWithInclude := `from django.contrib import admin
from django.urls import path, include

urlpatterns = [
    path('admin/', admin.site.urls),
]
`
	dir := t.TempDir()
	urlsPath := filepath.Join(dir, "urls.py")
	writeSampleFile(t, urlsPath, urlsWithInclude)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"django-silk"},
	}

	if err := patchUrls(urlsPath, cfg); err != nil {
		t.Fatalf("patchUrls failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(urlsPath)
	content := string(contentBytes)

	// Verify "include" only appears once in imports
	if strings.Count(content, "import path, include, include") > 0 {
		t.Errorf("TEST-GEN-15: duplicate include found in import: %s", content)
	}
}

func TestPatchUrls_PillowMediaServing(t *testing.T) {
	// TEST-GEN-16: Pillow Media Serving
	dir := t.TempDir()
	urlsPath := filepath.Join(dir, "urls.py")
	writeSampleFile(t, urlsPath, sampleUrlsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"pillow"},
	}

	if err := patchUrls(urlsPath, cfg); err != nil {
		t.Fatalf("patchUrls failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(urlsPath)
	content := string(contentBytes)

	if !strings.Contains(content, "from django.conf import settings") {
		t.Errorf("TEST-GEN-16: missing django.conf import settings")
	}
	if !strings.Contains(content, "from django.conf.urls.static import static") {
		t.Errorf("TEST-GEN-16: missing django.conf.urls.static import static")
	}
	if !strings.Contains(content, "if settings.DEBUG:") || !strings.Contains(content, "urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)") {
		t.Errorf("TEST-GEN-16: missing debug media url static route")
	}
}

func TestPatchUrls_BothSilkAndPillow(t *testing.T) {
	// TEST-GEN-17: Both Silk and Pillow Combined
	dir := t.TempDir()
	urlsPath := filepath.Join(dir, "urls.py")
	writeSampleFile(t, urlsPath, sampleUrlsPy)

	cfg := &config.ProjectConfig{
		Dependencies: []string{"django-silk", "pillow"},
	}

	if err := patchUrls(urlsPath, cfg); err != nil {
		t.Fatalf("patchUrls failed: %v", err)
	}

	contentBytes, _ := os.ReadFile(urlsPath)
	content := string(contentBytes)

	if !strings.Contains(content, "from django.urls import path, include") {
		t.Errorf("TEST-GEN-17: expected include in django.urls import")
	}
	if !strings.Contains(content, "from django.conf import settings") {
		t.Errorf("TEST-GEN-17: expected settings import")
	}
	if !strings.Contains(content, "path('silk/', include('silk.urls', namespace='silk')),") {
		t.Errorf("TEST-GEN-17: expected silk route")
	}
	if !strings.Contains(content, "urlpatterns += static(settings.MEDIA_URL") {
		t.Errorf("TEST-GEN-17: expected static media serving")
	}
}

func TestPatchUrls_ErrorHandling(t *testing.T) {
	// TEST-GEN-18: Non-existent urls file
	cfg := &config.ProjectConfig{
		Dependencies: []string{"django-silk"},
	}
	err := patchUrls(filepath.Join(t.TempDir(), "non_existent_urls.py"), cfg)
	if err == nil {
		t.Errorf("TEST-GEN-18: expected error for non-existent file, got nil")
	}
}

func TestScaffoldSettingsAndUrls(t *testing.T) {
	dir := t.TempDir()
	projectName := "test_proj"
	appDir := filepath.Join(dir, projectName)

	writeSampleFile(t, filepath.Join(appDir, "settings.py"), sampleSettingsPy)
	writeSampleFile(t, filepath.Join(appDir, "urls.py"), sampleUrlsPy)

	cfg := &config.ProjectConfig{
		OutputDir:    dir,
		ProjectName:  projectName,
		Dependencies: []string{"djangorestframework", "django-silk", "pillow"},
	}

	err := ScaffoldSettingsAndUrls(cfg)
	if err != nil {
		t.Fatalf("ScaffoldSettingsAndUrls failed: %v", err)
	}

	// Verify settings.py
	settingsBytes, _ := os.ReadFile(filepath.Join(appDir, "settings.py"))
	if !strings.Contains(string(settingsBytes), "'rest_framework',") {
		t.Errorf("ScaffoldSettingsAndUrls: missing 'rest_framework' in settings.py")
	}

	// Verify urls.py
	urlsBytes, _ := os.ReadFile(filepath.Join(appDir, "urls.py"))
	if !strings.Contains(string(urlsBytes), "path('silk/'") {
		t.Errorf("ScaffoldSettingsAndUrls: missing silk route in urls.py")
	}
}
