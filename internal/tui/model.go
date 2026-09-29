package tui

import (
	"context"
	"fmt"
	"regexp"
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
}

type DatabaseOption struct {
	ID          string
	Name        string
}

type Model struct {
	step      Step
	cfg       config.ProjectConfig
	textInput textinput.Model
	spinner   spinner.Model
	err       error
	width     int

	nameError string

	dockerChoice int

	depOptions []DependencyOption
	dbOptions []DatabaseOption
	depCursor  int
	depChecked map[string]bool

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
	ti.Placeholder = "my_amazing_django_project"
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
			},
			{
				ID:          "pillow",
				Name:        "Pillow",
			},
			{
				ID:          "django-silk",
				Name:        "Django Silk",
			},
			{
				ID:          "djangorestframework-simplejwt",
				Name:        "Django REST Framework Simple JWT",
			},
			{
				ID:          "django-filter",
				Name:        "Django Filters",
			},
			{
				ID:          "django-cors-headers",
				Name:        "Django CORS Headers",
			},
			{
				ID:          "django-environ",
				Name:        "Django Environ",
			},
			{
				ID:          "django-debug-toolbar",
				Name:        "Django Debug Toolbar",
			},
			{
				ID:          "django-modeltranslation",
				Name:        "Django ModelTranslation",
			},
		},
		dbOptions: []DatabaseOption{
			{
				ID:          "postgresql",
				Name:        "PostgreSQL",
			},
			{
				ID:          "sqlite",
				Name:        "Sqlite",
			},
		},
		depCursor:  0,
		depChecked: make(map[string]bool),
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
	m.cfg.Dependencies = deps
	m.cfg.WithDocker = (m.dockerChoice == 0)

	sub := m.progressSub
	cfgCopy := m.cfg

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
