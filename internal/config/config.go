package config

type ProjectConfig struct {
	ProjectName  string
	WithDocker   bool
	Dependencies []string // e.g. "djangorestframework", "pillow" and others
	Databases    []string // e.g. "sqlite", "postgresql" and others
	Apps         []string // e.g. "api", "base" starter apps to create
	Templates    string   // "none", "root", or "perapp"
	CacheBackend string   // "localmem", "redis", "dummy", "filebased", or "db"
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

func (c *ProjectConfig) HasDatabase(db string) bool {
	for _, d := range c.Databases {
		if d == db {
			return true
		}
	}
	return false
}
