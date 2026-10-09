package config

import "testing"

func TestProjectConfig_HasDependency(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []string
		searchDep    string
		want         bool
	}{
		{
			name:         "TEST-CFG-01: Found single dependency",
			dependencies: []string{"djangorestframework"},
			searchDep:    "djangorestframework",
			want:         true,
		},
		{
			name:         "TEST-CFG-01: Found among multiple dependencies",
			dependencies: []string{"djangorestframework", "pillow", "django-silk"},
			searchDep:    "django-silk",
			want:         true,
		},
		{
			name:         "TEST-CFG-02: Not found in non-empty slice",
			dependencies: []string{"djangorestframework", "pillow"},
			searchDep:    "django-silk",
			want:         false,
		},
		{
			name:         "TEST-CFG-03: Empty dependencies slice",
			dependencies: []string{},
			searchDep:    "djangorestframework",
			want:         false,
		},
		{
			name:         "TEST-CFG-03: Nil dependencies slice",
			dependencies: nil,
			searchDep:    "djangorestframework",
			want:         false,
		},
		{
			name:         "TEST-CFG-03: Partial substring match should return false",
			dependencies: []string{"django-silk", "djangorestframework-simplejwt"},
			searchDep:    "django",
			want:         false,
		},
		{
			name:         "TEST-CFG-03: Exact match with empty string search when slice contains empty string",
			dependencies: []string{""},
			searchDep:    "",
			want:         true,
		},
		{
			name:         "TEST-CFG-03: Empty string search when slice has no empty string",
			dependencies: []string{"pillow"},
			searchDep:    "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ProjectConfig{
				Dependencies: tt.dependencies,
			}
			got := cfg.HasDependency(tt.searchDep)
			if got != tt.want {
				t.Errorf("HasDependency(%q) = %v; want %v (Dependencies: %v)", tt.searchDep, got, tt.want, tt.dependencies)
			}
		})
	}
}
