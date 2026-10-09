package config

type ProjectConfig struct {
	ProjectName  string
	WithDocker   bool
	Dependencies []string // e.g. "djangorestframework", "pillow" and others
	Databases    []string // e.g. "sqlite3", "postgresql" and others
	UseUv        bool
	PythonCmd    string
	OutputDir    string
}

func (c *ProjectConfig) HasDependency(dep string) bool {
	for _, d := range c.Dependencies {
		if d == dep {
			return true
		}
	}
	return false
}
