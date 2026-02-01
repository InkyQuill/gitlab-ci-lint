package config

import (
	"fmt"
	"strings"

	"github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"
)

// MigrateLegacyConfig converts legacy single-instance configuration to new multi-instance format.
// This function automatically migrates when:
// - GitLab.Instances is empty (new format not used)
// - GitLab.Instance is set (legacy format present)
//
// The migration preserves all existing configuration and ensures backward compatibility.
func MigrateLegacyConfig(cfg *Config) {
	// Skip migration if already using new format
	if len(cfg.GitLab.Instances) > 0 {
		return
	}

	// Skip migration if no legacy instance configured
	if cfg.GitLab.Instance == "" {
		return
	}

	// Create instance from legacy configuration (no Default marking)
	instance := InstanceConfig{
		Name:    extractInstanceName(cfg.GitLab.Instance),
		URL:     gitlab.NormalizeInstanceURL(cfg.GitLab.Instance),
		Token:   cfg.Auth.Token,
		Timeout: cfg.GitLab.Timeout,
	}

	cfg.GitLab.Instances = []InstanceConfig{instance}

	// Note: We keep the legacy fields for now to maintain maximum compatibility
	// They can be cleaned up when saving the config
}

// extractInstanceName derives a suitable instance name from the instance URL.
// For example:
//   - https://gitlab.com → "gitlab.com"
//   - https://gitlab.example.com → "gitlab.example.com"
//   - https://git.a1tis.ru → "git.a1tis.ru"
func extractInstanceName(url string) string {
	// Normalize the URL first
	normalized := gitlab.NormalizeInstanceURL(url)

	// Remove protocol prefix
	name := strings.TrimPrefix(normalized, "https://")
	name = strings.TrimPrefix(name, "http://")

	// Remove port if present
	if idx := strings.Index(name, ":"); idx != -1 {
		name = name[:idx]
	}

	return name
}

// ValidateInstances validates the instance configuration.
// Returns an error if:
// - Instance names are not unique
// - Instance URLs are invalid
func ValidateInstances(cfg *Config) error {
	if len(cfg.GitLab.Instances) == 0 {
		return nil
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for _, inst := range cfg.GitLab.Instances {
		if inst.Name == "" {
			return fmt.Errorf("instance has empty name")
		}

		if names[inst.Name] {
			return fmt.Errorf("duplicate instance name: %s", inst.Name)
		}
		names[inst.Name] = true

		// Validate URL
		if inst.URL == "" {
			return fmt.Errorf("instance '%s' has empty URL", inst.Name)
		}

		// Normalize URL to check for duplicates
		normalizedURL := gitlab.NormalizeInstanceURL(inst.URL)
		if normalizedURL == "" {
			return fmt.Errorf("instance '%s' has invalid URL: %s", inst.Name, inst.URL)
		}
	}

	return nil
}

// FindInstanceByName searches for an instance by name.
// Returns nil if not found.
func (c *GitLabConfig) FindInstanceByName(name string) *InstanceConfig {
	for i := range c.Instances {
		if c.Instances[i].Name == name {
			return &c.Instances[i]
		}
	}
	return nil
}

// FindInstanceByURL searches for an instance by normalized URL.
// Returns nil if not found.
func (c *GitLabConfig) FindInstanceByURL(url string) *InstanceConfig {
	normalizedURL := gitlab.NormalizeInstanceURL(url)
	for i := range c.Instances {
		if gitlab.NormalizeInstanceURL(c.Instances[i].URL) == normalizedURL {
			return &c.Instances[i]
		}
	}
	return nil
}
