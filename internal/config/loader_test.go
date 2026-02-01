package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()

	if defaults.GitLab.Instance != "https://gitlab.com" {
		t.Errorf("Expected default instance to be 'https://gitlab.com', got '%s'", defaults.GitLab.Instance)
	}

	if defaults.GitLab.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", defaults.GitLab.Timeout)
	}

	if defaults.Validation.Strict != true {
		t.Errorf("Expected default strict to be true")
	}

	if defaults.Output.Format != "text" {
		t.Errorf("Expected default format to be 'text', got '%s'", defaults.Output.Format)
	}

	if defaults.Output.Color != "auto" {
		t.Errorf("Expected default color to be 'auto', got '%s'", defaults.Output.Color)
	}
}

func TestLoader_Load_DefaultsOnly(t *testing.T) {
	loader := NewLoader(&ConfigFlags{})

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load defaults: %v", err)
	}

	if cfg.GitLab.Instance != "https://gitlab.com" {
		t.Errorf("Expected default instance, got '%s'", cfg.GitLab.Instance)
	}
}

func TestLoader_Load_FlagsOverrideDefaults(t *testing.T) {
	flags := &ConfigFlags{
		Instance: "https://gitlab.example.com",
		Token:    "test-token",
		Strict:   true,
		Output:   "json",
		Verbose:  true,
		Color:    "never",
	}

	loader := NewLoader(flags)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.GitLab.Instance != "https://gitlab.example.com" {
		t.Errorf("Expected flag to override default instance, got '%s'", cfg.GitLab.Instance)
	}

	if cfg.Auth.Token != "test-token" {
		t.Errorf("Expected token from flag, got '%s'", cfg.Auth.Token)
	}

	if cfg.Output.Format != "json" {
		t.Errorf("Expected format from flag, got '%s'", cfg.Output.Format)
	}

	if cfg.Output.Color != "never" {
		t.Errorf("Expected color from flag, got '%s'", cfg.Output.Color)
	}
}

func TestLoader_Load_EnvVarsOverrideDefaults(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("GCL_INSTANCE", "https://env.gitlab.com")
	_ = os.Setenv("GCL_TOKEN", "env-token")
	_ = os.Setenv("GCL_OUTPUT_FORMAT", "yaml")
	_ = os.Setenv("GCL_VERBOSE", "true")
	defer func() {
		_ = os.Unsetenv("GCL_INSTANCE")
		_ = os.Unsetenv("GCL_TOKEN")
		_ = os.Unsetenv("GCL_OUTPUT_FORMAT")
		_ = os.Unsetenv("GCL_VERBOSE")
	}()

	loader := NewLoader(&ConfigFlags{})
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.GitLab.Instance != "https://env.gitlab.com" {
		t.Errorf("Expected env var to override default instance, got '%s'", cfg.GitLab.Instance)
	}

	if cfg.Auth.Token != "env-token" {
		t.Errorf("Expected token from env var, got '%s'", cfg.Auth.Token)
	}

	if cfg.Output.Format != "yaml" {
		t.Errorf("Expected format from env var, got '%s'", cfg.Output.Format)
	}

	if !cfg.Output.Verbose {
		t.Errorf("Expected verbose from env var to be true")
	}
}

func TestLoader_Load_FlagsOverrideEnvVars(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("GCL_TOKEN", "env-token")
	_ = os.Setenv("GCL_INSTANCE", "https://env.gitlab.com")
	defer func() {
		_ = os.Unsetenv("GCL_TOKEN")
		_ = os.Unsetenv("GCL_INSTANCE")
	}()

	// Flags should override env vars
	flags := &ConfigFlags{
		Instance: "https://flag.gitlab.com",
		Token:    "flag-token",
	}

	loader := NewLoader(flags)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.GitLab.Instance != "https://flag.gitlab.com" {
		t.Errorf("Expected flag to override env var, got '%s'", cfg.GitLab.Instance)
	}

	if cfg.Auth.Token != "flag-token" {
		t.Errorf("Expected flag token to override env var, got '%s'", cfg.Auth.Token)
	}
}

func TestLoader_Load_ConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
gitlab:
  instance: "https://config.gitlab.com"
  timeout: 60s
auth:
  token: "config-token"
validation:
  skip_api: true
  project: "test/project"
output:
  format: "yaml"
  verbose: true
  color: "always"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	flags := &ConfigFlags{
		ConfigFile: configPath,
	}

	loader := NewLoader(flags)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.GitLab.Instance != "https://config.gitlab.com" {
		t.Errorf("Expected instance from config file, got '%s'", cfg.GitLab.Instance)
	}

	if cfg.GitLab.Timeout != 60*time.Second {
		t.Errorf("Expected timeout from config file, got %v", cfg.GitLab.Timeout)
	}

	if cfg.Auth.Token != "config-token" {
		t.Errorf("Expected token from config file, got '%s'", cfg.Auth.Token)
	}

	if !cfg.Validation.SkipAPI {
		t.Errorf("Expected skip_api from config file to be true")
	}

	if cfg.Validation.Project != "test/project" {
		t.Errorf("Expected project from config file, got '%s'", cfg.Validation.Project)
	}

	if cfg.Output.Format != "yaml" {
		t.Errorf("Expected format from config file, got '%s'", cfg.Output.Format)
	}

	if !cfg.Output.Verbose {
		t.Errorf("Expected verbose from config file to be true")
	}

	if cfg.Output.Color != "always" {
		t.Errorf("Expected color from config file to be 'always', got '%s'", cfg.Output.Color)
	}
}

func TestLoader_Validate_InvalidOutputFormat(t *testing.T) {
	flags := &ConfigFlags{
		Output: "invalid",
	}

	loader := NewLoader(flags)
	_, err := loader.Load()

	if err == nil {
		t.Error("Expected validation error for invalid output format")
	}

	// Error is wrapped now, so check if it contains the expected message
	expectedErr := "invalid output format"
	if err != nil && !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

func TestLoader_Validate_InvalidColorSetting(t *testing.T) {
	flags := &ConfigFlags{
		Color: "invalid",
	}

	loader := NewLoader(flags)
	_, err := loader.Load()

	if err == nil {
		t.Error("Expected validation error for invalid color setting")
	}

	// Error is wrapped now
	expectedErr := "invalid color setting"
	if err != nil && !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErr, err)
	}
}

func TestLoader_GetConfigPath_FlagPriority(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	flagConfigPath := filepath.Join(tmpDir, "flag-config.yaml")

	flags := &ConfigFlags{
		ConfigFile: flagConfigPath,
	}

	loader := NewLoader(flags)
	path := loader.getConfigPath()

	if path != flagConfigPath {
		t.Errorf("Expected config path from flag, got '%s'", path)
	}
}

func TestLoader_GetConfigPath_EnvVarPriority(t *testing.T) {
	tmpDir := t.TempDir()
	envConfigPath := filepath.Join(tmpDir, "env-config.yaml")

	_ = os.Setenv("GCL_CONFIG", envConfigPath)
	defer func() { _ = os.Unsetenv("GCL_CONFIG") }()

	loader := NewLoader(&ConfigFlags{})
	path := loader.getConfigPath()

	if path != envConfigPath {
		t.Errorf("Expected config path from env var, got '%s'", path)
	}
}

func TestLoadEnvVars_BooleanParsing(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"True mixed case", "True", true},
		{"false lowercase", "false", false},
		{"FALSE uppercase", "FALSE", false},
		{"False mixed case", "False", false},
		{"random string", "random", false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Setenv("GCL_SKIP_API", tc.value)
			defer func() { _ = os.Unsetenv("GCL_SKIP_API") }()

			loader := NewLoader(&ConfigFlags{})
			cfg := loader.loadEnvVars()

			if cfg.Validation.SkipAPI != tc.expected {
				t.Errorf("Expected SKIP_API=%v for value '%s', got %v", tc.expected, tc.value, cfg.Validation.SkipAPI)
			}
		})
	}
}

func TestMergeConfigs_Precedence(t *testing.T) {
	base := Config{
		GitLab: GitLabConfig{
			Instance: "base-instance",
			Timeout:  10 * time.Second,
		},
		Auth: AuthConfig{
			Token: "base-token",
		},
		Validation: ValidationConfig{
			Project: "base-project",
		},
		Output: OutputConfig{
			Format:  "text",
			Verbose: false,
			Color:   "auto",
		},
	}

	override := Config{
		GitLab: GitLabConfig{
			Instance: "override-instance",
		},
		Auth: AuthConfig{
			Token: "override-token",
		},
		Validation: ValidationConfig{
			Strict: true,
		},
		Output: OutputConfig{
			Format: "json",
		},
	}

	loader := NewLoader(&ConfigFlags{})
	result := loader.mergeConfigs(base, override)

	// Check that override values are used
	if result.GitLab.Instance != "override-instance" {
		t.Errorf("Expected override instance, got '%s'", result.GitLab.Instance)
	}

	if result.Auth.Token != "override-token" {
		t.Errorf("Expected override token, got '%s'", result.Auth.Token)
	}

	if result.Output.Format != "json" {
		t.Errorf("Expected override format, got '%s'", result.Output.Format)
	}

	// Check that base values are preserved where not overridden
	if result.GitLab.Timeout != 10*time.Second {
		t.Errorf("Expected base timeout to be preserved, got %v", result.GitLab.Timeout)
	}

	if result.Validation.Project != "base-project" {
		t.Errorf("Expected base project to be preserved, got '%s'", result.Validation.Project)
	}

	if result.Output.Verbose != false {
		t.Errorf("Expected base verbose to be preserved, got %v", result.Output.Verbose)
	}

	// Check that boolean flags are properly merged
	if result.Validation.Strict != true {
		t.Errorf("Expected override strict to be true")
	}
}
