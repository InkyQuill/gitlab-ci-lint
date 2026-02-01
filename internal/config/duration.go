package config

import (
	"fmt"
	"time"
	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration to provide custom YAML marshaling
type Duration struct {
	time.Duration
}

// MarshalYAML marshals a duration to YAML as a string (e.g., "30s")
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// UnmarshalYAML unmarshals a duration from YAML string
// Supports all time.ParseDuration formats: "300ms", "1.5h", "2h45m", etc.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	d.Duration = duration
	return nil
}

// GetTimeout safely returns the timeout duration from GitLabConfig
func (c *GitLabConfig) GetTimeout() time.Duration {
	if c.Timeout == nil {
		return 30 * time.Second
	}
	return c.Timeout.Duration
}
