package config

import "time"

// GetDefaults returns the default configuration values
func GetDefaults() Config {
	defaultTimeout := 30 * time.Second
	return Config{
		GitLab: GitLabConfig{
			Instance: "https://gitlab.com",
			Timeout:  &defaultTimeout,
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
			Color:   "auto",
		},
		Files: FilesConfig{
			SearchParent:   true,
			MaxDepth:       10,
			IgnorePatterns: nil, // Will use discoverer defaults
		},
	}
}
