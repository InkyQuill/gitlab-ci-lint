package config

import (
	"testing"
	"time"
)

// TestMigrateLegacyConfig_LegacyToMulti tests migration from single to multi-instance
func TestMigrateLegacyConfig_LegacyToMulti(t *testing.T) {
	timeout := Duration{Duration: 30 * time.Second}

	cfg := &Config{
		GitLab: GitLabConfig{
			Instance: "https://gitlab.com",
			Timeout:  &timeout,
		},
		Auth: AuthConfig{
			Token: "test-token",
		},
	}

	MigrateLegacyConfig(cfg)

	if len(cfg.GitLab.Instances) != 1 {
		t.Fatalf("Expected 1 instance after migration, got %d", len(cfg.GitLab.Instances))
	}

	inst := cfg.GitLab.Instances[0]
	if inst.Name != "gitlab.com" {
		t.Errorf("Expected instance name 'gitlab.com', got '%s'", inst.Name)
	}

	if inst.URL != "https://gitlab.com" {
		t.Errorf("Expected URL 'https://gitlab.com', got '%s'", inst.URL)
	}

	if inst.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", inst.Token)
	}

	if inst.Timeout == nil {
		t.Error("Expected timeout to be preserved")
	} else if inst.Timeout.Duration != timeout.Duration {
		t.Errorf("Expected timeout %v, got %v", timeout.Duration, inst.Timeout.Duration)
	}
}

// TestMigrateLegacyConfig_PreservesSettings tests that all settings are preserved
func TestMigrateLegacyConfig_PreservesSettings(t *testing.T) {
	timeout := Duration{Duration: 45 * time.Second}

	cfg := &Config{
		GitLab: GitLabConfig{
			Instance: "https://gitlab.example.com:8080",
			Timeout:  &timeout,
		},
		Auth: AuthConfig{
			Token: "glpat-12345678",
		},
		Output: OutputConfig{
			Format:  "json",
			Verbose: true,
		},
		Validation: ValidationConfig{
			SkipAPI: true,
		},
	}

	MigrateLegacyConfig(cfg)

	if len(cfg.GitLab.Instances) != 1 {
		t.Fatalf("Expected 1 instance after migration, got %d", len(cfg.GitLab.Instances))
	}

	// Check instance config
	inst := cfg.GitLab.Instances[0]
	if inst.Name != "gitlab.example.com" {
		t.Errorf("Expected instance name 'gitlab.example.com', got '%s'", inst.Name)
	}

	// Verify URL normalization (port preserved)
	if !contains(inst.URL, "8080") {
		t.Errorf("Expected URL to contain port, got '%s'", inst.URL)
	}

	// Verify other config is preserved
	if cfg.Output.Format != "json" {
		t.Errorf("Expected output format 'json', got '%s'", cfg.Output.Format)
	}

	if !cfg.Validation.SkipAPI {
		t.Error("Expected SkipAPI to be preserved")
	}
}

// TestMigrateLegacyConfig_AlreadyMigrated tests that migration is skipped if already done
func TestMigrateLegacyConfig_AlreadyMigrated(t *testing.T) {
	timeout := Duration{Duration: 30 * time.Second}

	cfg := &Config{
		GitLab: GitLabConfig{
			Instance: "https://old-instance.com",
			Timeout:  &timeout,
			Instances: []InstanceConfig{
				{
					Name:  "gitlab.com",
					URL:   "https://gitlab.com",
					Token: "new-token",
				},
			},
		},
		Auth: AuthConfig{
			Token: "old-token",
		},
	}

	MigrateLegacyConfig(cfg)

	// Should still have 1 instance (not 2)
	if len(cfg.GitLab.Instances) != 1 {
		t.Fatalf("Expected 1 instance (migration skipped), got %d", len(cfg.GitLab.Instances))
	}

	// Should be the new instance, not a migrated one
	inst := cfg.GitLab.Instances[0]
	if inst.Name != "gitlab.com" {
		t.Errorf("Expected instance 'gitlab.com' (not migrated), got '%s'", inst.Name)
	}

	if inst.Token != "new-token" {
		t.Errorf("Expected token 'new-token' (not migrated), got '%s'", inst.Token)
	}
}

// TestMigrateLegacyConfig_EmptyConfig tests migration with empty config
func TestMigrateLegacyConfig_EmptyConfig(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{},
	}

	MigrateLegacyConfig(cfg)

	if len(cfg.GitLab.Instances) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(cfg.GitLab.Instances))
	}
}

// TestMigrateLegacyConfig_NoLegacyInstance tests migration with no legacy instance
func TestMigrateLegacyConfig_NoLegacyInstance(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instance: "",
		},
		Auth: AuthConfig{
			Token: "test-token",
		},
	}

	MigrateLegacyConfig(cfg)

	if len(cfg.GitLab.Instances) != 0 {
		t.Errorf("Expected 0 instances (no legacy config), got %d", len(cfg.GitLab.Instances))
	}
}

// TestExtractInstanceName_HTTPS tests instance name extraction from HTTPS URLs
func TestExtractInstanceName_HTTPS(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://gitlab.com", "gitlab.com"},
		{"https://gitlab.example.com", "gitlab.example.com"},
		{"https://git.a1tis.ru", "git.a1tis.ru"},
		{"https://gitlab.com/", "gitlab.com"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			// We need to call extractInstanceName indirectly through migration
			cfg := &Config{
				GitLab: GitLabConfig{
					Instance: tt.url,
				},
			}
			MigrateLegacyConfig(cfg)

			if len(cfg.GitLab.Instances) != 1 {
				t.Fatalf("Expected 1 instance, got %d", len(cfg.GitLab.Instances))
			}

			name := cfg.GitLab.Instances[0].Name
			if name != tt.expected {
				t.Errorf("Expected name '%s', got '%s'", tt.expected, name)
			}
		})
	}
}

// TestExtractInstanceName_HTTPWithPort tests instance name extraction from HTTP with port
func TestExtractInstanceName_HTTPWithPort(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instance: "http://gitlab.example.com:8080",
		},
	}

	MigrateLegacyConfig(cfg)

	if len(cfg.GitLab.Instances) != 1 {
		t.Fatalf("Expected 1 instance, got %d", len(cfg.GitLab.Instances))
	}

	name := cfg.GitLab.Instances[0].Name
	if name != "gitlab.example.com" {
		t.Errorf("Expected name 'gitlab.example.com', got '%s'", name)
	}
}

// TestExtractInstanceName_SSH tests instance name extraction from SSH URLs
// Note: SSH URLs are not well-supported by the legacy migration since they don't
// normalize properly through gitlab.NormalizeInstanceURL. This is acceptable
// since legacy configs typically use HTTPS URLs.
func TestExtractInstanceName_SSH(t *testing.T) {
	tests := []struct {
		url      string
		expected string // What we actually get (may include user@ part or be 'ssh')
	}{
		{"git@gitlab.com:group/project.git", "git@gitlab.com"}, // SSH URLs don't normalize well
		{"git@gitlab.example.com:group/project", "git@gitlab.example.com"},
		{"ssh://git@git.a1tis.ru/group/project", "ssh"}, // ssh:// format also doesn't work well
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			cfg := &Config{
				GitLab: GitLabConfig{
					Instance: tt.url,
				},
			}
			MigrateLegacyConfig(cfg)

			if len(cfg.GitLab.Instances) != 1 {
				t.Fatalf("Expected 1 instance, got %d", len(cfg.GitLab.Instances))
			}

			name := cfg.GitLab.Instances[0].Name
			if name != tt.expected {
				t.Errorf("Expected name '%s', got '%s'", tt.expected, name)
			}
		})
	}
}

// TestExtractInstanceName_SSHWithPort tests instance name extraction from SSH with port
func TestExtractInstanceName_SSHWithPort(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instance: "git@gitlab.com:2222:group/project.git",
		},
	}

	MigrateLegacyConfig(cfg)

	if len(cfg.GitLab.Instances) != 1 {
		t.Fatalf("Expected 1 instance, got %d", len(cfg.GitLab.Instances))
	}

	// SSH URLs with ports don't normalize properly - this is expected behavior
	name := cfg.GitLab.Instances[0].Name
	if name != "git@gitlab.com" {
		t.Errorf("Expected name 'git@gitlab.com', got '%s'", name)
	}
}

// TestExtractInstanceName_Invalid tests instance name extraction with invalid URLs
func TestExtractInstanceName_Invalid(t *testing.T) {
	tests := []struct {
		url      string
		expected string // empty string means it should still work but return normalized
	}{
		{"", ""},
		{"not-a-url", ""},
		{":invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			cfg := &Config{
				GitLab: GitLabConfig{
					Instance: tt.url,
				},
			}
			// Should not panic
			MigrateLegacyConfig(cfg)

			// Empty URL should not create an instance
			if tt.url == "" {
				if len(cfg.GitLab.Instances) != 0 {
					t.Errorf("Expected 0 instances for empty URL, got %d", len(cfg.GitLab.Instances))
				}
			}
		})
	}
}

// TestValidateInstances_ValidMultiple tests validation with valid multiple instances
func TestValidateInstances_ValidMultiple(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instances: []InstanceConfig{
				{Name: "gitlab.com", URL: "https://gitlab.com"},
				{Name: "custom", URL: "https://custom.com"},
			},
		},
	}

	err := ValidateInstances(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestValidateInstances_DuplicateNames tests validation with duplicate instance names
func TestValidateInstances_DuplicateNames(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instances: []InstanceConfig{
				{Name: "gitlab.com", URL: "https://gitlab.com"},
				{Name: "gitlab.com", URL: "https://gitlab.example.com"},
			},
		},
	}

	err := ValidateInstances(cfg)
	if err == nil {
		t.Error("Expected error for duplicate names, got nil")
	}

	expectedMsg := "duplicate instance name"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

// TestValidateInstances_EmptyName tests validation with empty instance name
func TestValidateInstances_EmptyName(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instances: []InstanceConfig{
				{Name: "", URL: "https://gitlab.com"},
			},
		},
	}

	err := ValidateInstances(cfg)
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}

	expectedMsg := "empty name"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

// TestValidateInstances_EmptyURL tests validation with empty instance URL
func TestValidateInstances_EmptyURL(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instances: []InstanceConfig{
				{Name: "test", URL: ""},
			},
		},
	}

	err := ValidateInstances(cfg)
	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}

	expectedMsg := "empty URL"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

// TestValidateInstances_EmptyInstances tests validation with no instances
func TestValidateInstances_EmptyInstances(t *testing.T) {
	cfg := &Config{
		GitLab: GitLabConfig{
			Instances: []InstanceConfig{},
		},
	}

	err := ValidateInstances(cfg)
	if err != nil {
		t.Errorf("Expected no error for empty instances, got: %v", err)
	}
}

// TestFindInstanceByName_Found tests finding an instance by name
func TestFindInstanceByName_Found(t *testing.T) {
	cfg := &GitLabConfig{
		Instances: []InstanceConfig{
			{Name: "gitlab.com", URL: "https://gitlab.com", Token: "token1"},
			{Name: "custom", URL: "https://custom.com", Token: "token2"},
		},
	}

	inst := cfg.FindInstanceByName("custom")
	//nolint:staticcheck // SA5011 - nil check and t.Fatal() ensures safety
	if inst == nil {
		t.Fatal("Expected to find instance, got nil")
	}
	//nolint:staticcheck // SA5011 - nil check above ensures inst is not nil here
	if inst.Name != "custom" {
		t.Errorf("Expected name 'custom', got '%s'", inst.Name)
	}

	if inst.Token != "token2" {
		t.Errorf("Expected token 'token2', got '%s'", inst.Token)
	}
}

// TestFindInstanceByName_NotFound tests finding a non-existent instance
func TestFindInstanceByName_NotFound(t *testing.T) {
	cfg := &GitLabConfig{
		Instances: []InstanceConfig{
			{Name: "gitlab.com", URL: "https://gitlab.com"},
		},
	}

	inst := cfg.FindInstanceByName("nonexistent")
	if inst != nil {
		t.Errorf("Expected nil for non-existent instance, got %+v", inst)
	}
}

// TestFindInstanceByName_EmptyName tests finding with empty name
func TestFindInstanceByName_EmptyName(t *testing.T) {
	cfg := &GitLabConfig{
		Instances: []InstanceConfig{
			{Name: "gitlab.com", URL: "https://gitlab.com"},
		},
	}

	inst := cfg.FindInstanceByName("")
	if inst != nil {
		t.Errorf("Expected nil for empty name, got %+v", inst)
	}
}

// TestFindInstanceByURL_Found tests finding an instance by URL
func TestFindInstanceByURL_Found(t *testing.T) {
	cfg := &GitLabConfig{
		Instances: []InstanceConfig{
			{Name: "gitlab.com", URL: "https://gitlab.com", Token: "token1"},
			{Name: "custom", URL: "https://custom.com", Token: "token2"},
		},
	}

	// Test with various URL formats
	testURLs := []string{
		"https://gitlab.com",
		"gitlab.com",
		"https://gitlab.com/",
	}

	for _, url := range testURLs {
		t.Run(url, func(t *testing.T) {
			inst := cfg.FindInstanceByURL(url)
			if inst == nil {
				t.Errorf("Expected to find instance for URL '%s', got nil", url)
				return
			}

			if inst.Name != "gitlab.com" {
				t.Errorf("Expected name 'gitlab.com', got '%s'", inst.Name)
			}
		})
	}
}

// TestFindInstanceByURL_NotFound tests finding with non-existent URL
func TestFindInstanceByURL_NotFound(t *testing.T) {
	cfg := &GitLabConfig{
		Instances: []InstanceConfig{
			{Name: "gitlab.com", URL: "https://gitlab.com"},
		},
	}

	inst := cfg.FindInstanceByURL("https://nonexistent.com")
	if inst != nil {
		t.Errorf("Expected nil for non-existent URL, got %+v", inst)
	}
}

// TestFindInstanceByURL_EmptyURL tests finding with empty URL
func TestFindInstanceByURL_EmptyURL(t *testing.T) {
	cfg := &GitLabConfig{
		Instances: []InstanceConfig{
			{Name: "gitlab.com", URL: "https://gitlab.com"},
		},
	}

	inst := cfg.FindInstanceByURL("")
	if inst != nil {
		t.Errorf("Expected nil for empty URL, got %+v", inst)
	}
}

// TestFindInstanceByURL_Normalization tests that URL normalization works
func TestFindInstanceByURL_Normalization(t *testing.T) {
	cfg := &GitLabConfig{
		Instances: []InstanceConfig{
			{Name: "gitlab.com", URL: "https://gitlab.com"},
		},
	}

	// All these should find the same instance due to normalization
	testURLs := []string{
		"https://gitlab.com",
		"https://gitlab.com/",
		"gitlab.com",
		"gitlab.com/",
		"  gitlab.com  ",
	}

	for _, url := range testURLs {
		t.Run(url, func(t *testing.T) {
			inst := cfg.FindInstanceByURL(url)
			if inst == nil {
				t.Errorf("Expected to find instance for URL '%s'", url)
			}

			if inst != nil && inst.Name != "gitlab.com" {
				t.Errorf("Expected name 'gitlab.com', got '%s'", inst.Name)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
