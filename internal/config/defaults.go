package config

import "time"

// GetDefaults returns the default configuration values
func GetDefaults() Config {
	return Config{
		GitLab: GitLabConfig{
			Instance: "https://gitlab.com",
			Timeout:  30 * time.Second,
		},
		Auth: AuthConfig{
			Token: "",
			Netrc: false,
		},
		Validation: ValidationConfig{
			SkipAPI: false,
			Strict:  true,
			Project: "",
		},
		Output: OutputConfig{
			Format:  "text",
			Verbose: false,
			Color:   "auto",
		},
	}
}
