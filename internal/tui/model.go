package tui

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/rabee361/django-launch/internal/config"
	"github.com/rabee361/django-launch/internal/generator"
)

type Step int

const (
	StepProjectName Step = iota
	StepDocker
	StepDependencies
	StepDataBase
	StepAppsCreated
	StepTemplates
	StepConfirm
	StepExecuting
	StepDone
	StepError
)

type progressMsg struct {
	step  int
	total int
	text  string
}

type doneMsg struct {
	err error
}

type DependencyOption struct {
	ID          string
	Name        string
	Description string
}

type DatabaseOption struct {
	ID          string
	Name        string
	Description string
}

type AppOption struct {
	ID          string
	Name        string
	Description string
}

type TemplateOption struct {
	ID          string
	Name        string
	Description string
}

type Model struct {
	step      Step
	cfg       config.ProjectConfig
	textInput textinput.Model
	spinner   spinner.Model
	err       error
	width     int
	height    int

	nameError string

	dockerChoice int

	depOptions  []DependencyOption
	dbOptions   []DatabaseOption
	appOptions  []AppOption
	depCursor   int
	dbCursor    int
	appsCursor  int
	depChecked  map[string]bool
	appsChecked map[string]bool
	appsChoice  int // 0 = No, 1 = Yes
	appsPhase   int // 0 = ask yes/no, 1 = select apps

	templatesChoice int

	progressSub   chan tea.Msg
	progressItems []string
	currentTask   string
	progressStep  int
	progressTotal int
	cancelGen     context.CancelFunc

	hasUv     bool
	pythonCmd string
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "my_amazing_project"
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 50
	ti.SetWidth(50)

	s := spinner.New()
	s.Spinner = spinner.Moon
	s.Style = focusedPromptStyle

	hasUv, pyCmd, _ := generator.DetectTooling()

	m := Model{
		step:      StepProjectName,
		textInput: ti,
		spinner:   s,
		cfg: config.ProjectConfig{
			UseUv:     hasUv,
			PythonCmd: pyCmd,
		},
		hasUv:        hasUv,
		pythonCmd:    pyCmd,
		dockerChoice: 0,
		depOptions: []DependencyOption{
			{
				ID:          "djangorestframework",
				Name:        "Django REST Framework",
				Description: "Toolkit for building Web APIs",
			},
			{
				ID:          "pillow",
				Name:        "Pillow",
				Description: "Python Imaging Library (required for ImageField & media)",
			},
			{
				ID:          "django-silk",
				Name:        "Django Silk",
				Description: "Profiling and SQL query inspection",
			},
			{
				ID:          "djangorestframework-simplejwt",
				Name:        "Django REST Framework Simple JWT",
				Description: "Token authentication for the REST framework",
			},
			{
				ID:          "django-filter",
				Name:        "Django Filters",
				Description: "Filtering for API views",
			},
			{
				ID:          "django-cors-headers",
				Name:        "Django CORS Headers",
				Description: "CORS response headers",
			},
			{
				ID:          "django-environ",
				Name:        "Django Environ",
				Description: "Environment variable handling with .env support",
			},
			{
				ID:          "django-debug-toolbar",
				Name:        "Django Debug Toolbar",
				Description: "Debug panels during development",
			},
			{
				ID:          "django-modeltranslation",
				Name:        "Django ModelTranslation",
				Description: "Translated model fields",
			},
		},
		dbOptions: []DatabaseOption{
			{
				ID:          "postgresql",
				Name:        "PostgreSQL",
				Description: "Full-featured relational database (psycopg is installed for you)",
			},
			{
				ID:          "sqlite",
				Name:        "Sqlite",
				Description: "File-based zero-configuration database",
			},
		},
		appOptions: []AppOption{
			{
				ID:          "api",
				Name:        "api",
				Description: "REST API endpoints app",
			},
			{
				ID:          "base",
				Name:        "base",
				Description: "Shared utilities and models app",
			},
		},
		depCursor:   0,
		dbCursor:    0,
		appsCursor:  0,
		depChecked:  make(map[string]bool),
		appsChecked: make(map[string]bool),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func validateProjectName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	// Python identifier validation: lowercase letters, digits, underscore, must not start with digit
	match, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	if !match {
		return fmt.Errorf("name must be a valid Python identifier (letters, digits, underscores, no hyphens or spaces)")
	}
	reserved := map[string]bool{
		"django": true, "test": true, "site": true, "os": true, "sys": true, "admin": true,
	}
	if reserved[strings.ToLower(name)] {
		return fmt.Errorf("'%s' is a reserved name in Python/Django, please choose another name", name)
	}
	return nil
}

// selectedApps returns the ids of the apps the user picked.
func (m Model) selectedApps() []string {
	var apps []string
	for _, opt := range m.appOptions {
		if m.appsChecked[opt.ID] {
			apps = append(apps, opt.ID)
		}
	}
	return apps
}

// templateOptions returns the templates folder choices; per-app folders only
// make sense when at least one app is created.
func (m Model) templateOptions() []TemplateOption {
	opts := []TemplateOption{
		{
			ID:          "none",
			Name:        "No templates folder",
			Description: "Skip template directories entirely",
		},
		{
			ID:          "root",
			Name:        "Project root templates/ folder",
			Description: "A single templates/ folder at the project root",
		},
	}
	if len(m.selectedApps()) > 0 {
		opts = append(opts, TemplateOption{
			ID:          "perapp",
			Name:        "templates/ in each app",
			Description: "A templates/ folder inside every created app",
		})
	}
	return opts
}

func (m *Model) startGeneration() tea.Cmd {
	m.progressSub = make(chan tea.Msg)
	m.progressItems = nil
	m.currentTask = "Starting project generation..."

	var deps []string
	for _, opt := range m.depOptions {
		if m.depChecked[opt.ID] {
			deps = append(deps, opt.ID)
		}
	}

	var dbs []string
	for _, opt := range m.dbOptions {
		if m.depChecked[opt.ID] {
			dbs = append(dbs, opt.ID)
		}
	}

	apps := m.selectedApps()

	m.cfg.Dependencies = deps
	m.cfg.Databases = dbs
	m.cfg.Apps = apps
	m.cfg.Templates = "none"
	if opts := m.templateOptions(); m.templatesChoice < len(opts) {
		m.cfg.Templates = opts[m.templatesChoice].ID
	}
	m.cfg.WithDocker = (m.dockerChoice == 0)

	sub := m.progressSub
	cfgCopy := m.cfg
	cfgCopy.Dependencies = slices.Clone(deps)
	cfgCopy.Databases = slices.Clone(dbs)
	cfgCopy.Apps = slices.Clone(apps)

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelGen = cancel

	go func() {
		err := generator.GenerateProject(ctx, &cfgCopy, func(step, total int, desc string) {
			sendProgress(ctx, sub, progressMsg{
				step:  step,
				total: total,
				text:  desc,
			})
		})
		sendProgress(ctx, sub, doneMsg{err: err})
	}()

	return tea.Batch(
		m.spinner.Tick,
		waitForActivity(m.progressSub),
	)
}

func sendProgress(ctx context.Context, sub chan tea.Msg, msg tea.Msg) {
	select {
	case sub <- msg:
	case <-ctx.Done():
	}
}

func (m *Model) clearProgress() {
	m.progressItems = nil
	m.currentTask = ""
	m.progressStep = 0
	m.progressTotal = 0
}

func waitForActivity(sub chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func (m Model) nextStepsText() string {
	var activateCmd string
	if runtime.GOOS == "windows" {
		activateCmd = fmt.Sprintf(".\\%s\\.venv\\Scripts\\activate", m.cfg.ProjectName)
	} else {
		activateCmd = fmt.Sprintf("source %s/.venv/bin/activate", m.cfg.ProjectName)
	}

	return fmt.Sprintf(
		"Next Steps:\n\n"+
			"  1. Enter your project directory:\n"+
			"     cd %s\n\n"+
			"  2. Activate virtual environment:\n"+
			"     %s\n\n"+
			"  3. Apply initial migrations:\n"+
			"     python manage.py migrate\n\n"+
			"  4. Start development server:\n"+
			"     python manage.py runserver\n",
		m.cfg.ProjectName,
		activateCmd,
	)
}
