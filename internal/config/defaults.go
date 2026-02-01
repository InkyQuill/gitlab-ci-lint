package config

import "time"

// GetDefaults returns the default configuration values
func GetDefaults() Config {
	defaultTimeout := &Duration{Duration: 30 * time.Second}
	return Config{
		GitLab: GitLabConfig{
			// Legacy single-instance defaults (for backward compatibility)
			Instance: "https://gitlab.com",
			Timeout:  defaultTimeout,
			// New multi-instance defaults (empty by default)
			Instances: nil, // No instances configured by default
		},
		Auth: AuthConfig{
			Token: "",
			Netrc: false,
		},
		Validation: ValidationConfig{
			SkipAPI: false,
			Strict:  false,
			Project: "",
		},
		Output: OutputConfig{
			Format:  "text",
			Verbose: false,
			Debug:   false,
			Color:   "auto",
		},
		Files: FilesConfig{
			SearchParent:   true,
			MaxDepth:       10,
			IgnorePatterns: nil, // Will use discoverer defaults
		},
	}
}
