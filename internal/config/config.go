package config

// ProjectConfig stores user choices for generating the Django project.
type ProjectConfig struct {
	ProjectName  string
	WithDocker   bool
	Dependencies []string // e.g. "djangorestframework", "pillow" and others
	UseUv        bool
	PythonCmd    string
	OutputDir    string
}

// HasDependency returns true if the specified dependency was selected by the user.
func (c *ProjectConfig) HasDependency(dep string) bool {
	for _, d := range c.Dependencies {
		if d == dep {
			return true
		}
	}
	return false
}
