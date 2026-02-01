package config

// Config represents the application configuration
type Config struct {
	GitLab     GitLabConfig     `yaml:"gitlab"`
	Auth       AuthConfig       `yaml:"auth"`
	Validation ValidationConfig `yaml:"validation"`
	Output     OutputConfig     `yaml:"output"`
	Files      FilesConfig      `yaml:"files"`
}

// GitLabConfig contains GitLab instance settings
type GitLabConfig struct {
	// Legacy single-instance support (deprecated but kept for backward compatibility)
	Instance string    `yaml:"instance,omitempty"` // DEPRECATED: Use instances instead
	Timeout  *Duration `yaml:"timeout,omitempty"`

	// Multi-instance support
	Instances []InstanceConfig `yaml:"instances,omitempty"` // List of configured GitLab instances
}

// InstanceConfig represents a single GitLab instance configuration
type InstanceConfig struct {
	Name    string    `yaml:"name"`               // Unique identifier (e.g., "gitlab.com", "work")
	URL     string    `yaml:"url"`                // Instance URL (e.g., "https://gitlab.com")
	Token   string    `yaml:"token,omitempty"`    // Personal access token
	Timeout *Duration `yaml:"timeout,omitempty"`  // Optional timeout override for this instance
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Token string `yaml:"token"`
	Netrc bool   `yaml:"netrc"`
}

// ValidationConfig contains validation behavior settings
type ValidationConfig struct {
	SkipAPI bool   `yaml:"skip_api"`
	Strict  bool   `yaml:"strict"`
	// Project is NOT loaded from config file (yaml:"-" prevents marshaling)
	// It can only be set via CLI flag or environment variable (GCL_PROJECT)
	// This allows per-file auto-detection from .git/config
	Project string `yaml:"-"`
}

// OutputConfig contains output formatting settings
type OutputConfig struct {
	Format  string `yaml:"format"` // text, json, yaml
	Verbose bool   `yaml:"verbose"`
	Debug   bool   `yaml:"debug"`
	Color   string `yaml:"color"` // auto, always, never
}

// FilesConfig contains file discovery settings
type FilesConfig struct {
	SearchParent   bool     `yaml:"search_parent"`
	MaxDepth       int      `yaml:"max_depth"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
}
