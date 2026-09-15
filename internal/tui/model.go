package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"main/internal/config"
	"main/internal/generator"
)

type Step int

const (
	StepProjectName Step = iota
	StepDocker
	StepDependencies
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

type Model struct {
	step      Step
	cfg       config.ProjectConfig
	textInput textinput.Model
	spinner   spinner.Model
	err       error

	// Step 1: Project Name
	nameError string

	// Step 2: Docker Choice (0 = Yes, 1 = No)
	dockerChoice int

	// Step 3: Dependencies
	depOptions []DependencyOption
	depCursor  int
	depChecked map[string]bool

	// Step 5: Executing
	progressSub   chan tea.Msg
	progressItems []string
	currentTask   string

	// System detection
	hasUv     bool
	pythonCmd string
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "my_django_app"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
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
		dockerChoice: 0, // Default to Yes
		depOptions: []DependencyOption{
			{
				ID:          "djangorestframework",
				Name:        "Django REST Framework",
				Description: "Powerful and flexible toolkit for building Web APIs",
			},
			{
				ID:          "pillow",
				Name:        "Pillow",
				Description: "Python Imaging Library (required for ImageField & media)",
			},
			{
				ID:          "django-silk",
				Name:        "Django Silk",
				Description: "Silky-smooth profiling and SQL query inspection",
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

	// Collect dependencies
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

	go func() {
		err := generator.GenerateProject(&cfgCopy, func(step, total int, desc string) {
			sub <- progressMsg{
				step:  step,
				total: total,
				text:  desc,
			}
		})
		sub <- doneMsg{err: err}
	}()

	return tea.Batch(
		m.spinner.Tick,
		waitForActivity(m.progressSub),
	)
}

func waitForActivity(sub chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}
