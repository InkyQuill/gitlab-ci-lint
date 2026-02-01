package config

import "time"

// Config represents the application configuration
type Config struct {
	GitLab      GitLabConfig      `yaml:"gitlab"`
	Auth        AuthConfig        `yaml:"auth"`
	Validation  ValidationConfig  `yaml:"validation"`
	Output      OutputConfig      `yaml:"output"`
}

// GitLabConfig contains GitLab instance settings
type GitLabConfig struct {
	Instance string        `yaml:"instance"`
	Timeout  time.Duration `yaml:"timeout"`
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
	Project string `yaml:"project"`
}

// OutputConfig contains output formatting settings
type OutputConfig struct {
	Format string `yaml:"format"` // text, json, yaml
	Verbose bool   `yaml:"verbose"`
	Color   string `yaml:"color"`  // auto, always, never
}
