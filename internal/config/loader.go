package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from multiple sources with priority
type Loader struct {
	defaults Config
	flags    *ConfigFlags
}

// ConfigFlags represents CLI flag values
type ConfigFlags struct {
	ConfigFile  string
	Token       string
	Netrc       bool
	Instance    string
	Timeout     string
	Project     string
	SkipAPI     bool
	Strict      bool
	Output      string
	Verbose     bool
	Color       string
	Files       []string // From -f flag
	Directories []string // From -d flag
}

// NewLoader creates a new configuration loader
func NewLoader(flags *ConfigFlags) *Loader {
	return &Loader{
		defaults: GetDefaults(),
		flags:    flags,
	}
}

// parseTimeout parses a timeout string into a time.Duration pointer
func (l *Loader) parseTimeout(timeoutStr string) (*time.Duration, error) {
	if timeoutStr == "" {
		return nil, nil // Not set
	}
	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout format: %w", err)
	}
	return &duration, nil
}

// Load loads configuration from all sources and merges them with proper priority
// Priority (low to high): defaults -> config file -> env vars -> CLI flags
func (l *Loader) Load() (*Config, error) {
	cfg := l.defaults

	// 1. Load config file if specified or use default path
	configPath := l.getConfigPath()
	if configPath != "" {
		fileCfg, err := l.loadConfigFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		cfg = l.mergeConfigs(cfg, fileCfg)
	}

	// 2. Load environment variables
	envCfg := l.loadEnvVars()
	cfg = l.mergeConfigs(cfg, envCfg)

	// 3. Load CLI flags
	flagCfg, err := l.loadFlags()
	if err != nil {
		return nil, fmt.Errorf("failed to load flags: %w", err)
	}
	cfg = l.mergeConfigs(cfg, flagCfg)

	// Validate final configuration
	if err := l.validate(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// getConfigPath returns the config file path to use
func (l *Loader) getConfigPath() string {
	// If config file is explicitly specified via flag, use it
	if l.flags.ConfigFile != "" {
		return l.flags.ConfigFile
	}

	// Check GCL_CONFIG environment variable
	if envPath := os.Getenv("GCL_CONFIG"); envPath != "" {
		return envPath
	}

	// Check default config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	defaultPath := filepath.Join(homeDir, ".tools-config", ".gitlab-ci-lint", "config.yaml")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}

	return ""
}

// loadConfigFile loads configuration from a YAML file
func (l *Loader) loadConfigFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// loadEnvVars loads configuration from environment variables
func (l *Loader) loadEnvVars() Config {
	cfg := Config{}

	if token := os.Getenv("GCL_TOKEN"); token != "" {
		cfg.Auth.Token = token
	}

	if instance := os.Getenv("GCL_INSTANCE"); instance != "" {
		cfg.GitLab.Instance = instance
	}

	if timeout := os.Getenv("GCL_TIMEOUT"); timeout != "" {
		parsed, err := l.parseTimeout(timeout)
		if err == nil && parsed != nil {
			cfg.GitLab.Timeout = parsed
		}
	}

	if project := os.Getenv("GCL_PROJECT"); project != "" {
		cfg.Validation.Project = project
	}

	if skipAPI := os.Getenv("GCL_SKIP_API"); skipAPI != "" {
		cfg.Validation.SkipAPI = strings.ToLower(skipAPI) == "true"
	}

	if strict := os.Getenv("GCL_STRICT"); strict != "" {
		cfg.Validation.Strict = strings.ToLower(strict) == "true"
	}

	if format := os.Getenv("GCL_OUTPUT_FORMAT"); format != "" {
		cfg.Output.Format = format
	}

	if verbose := os.Getenv("GCL_VERBOSE"); verbose != "" {
		cfg.Output.Verbose = strings.ToLower(verbose) == "true"
	}

	if color := os.Getenv("GCL_COLOR"); color != "" {
		cfg.Output.Color = color
	}

	return cfg
}

// loadFlags loads configuration from CLI flags
func (l *Loader) loadFlags() (Config, error) {
	cfg := Config{}

	if l.flags.Token != "" {
		cfg.Auth.Token = l.flags.Token
	}

	if l.flags.Netrc {
		cfg.Auth.Netrc = true
	}

	if l.flags.Instance != "" {
		cfg.GitLab.Instance = l.flags.Instance
	}

	// Timeout is parsed during flag loading
	if l.flags.Timeout != "" {
		parsed, err := l.parseTimeout(l.flags.Timeout)
		if err != nil {
			return Config{}, fmt.Errorf("invalid timeout flag: %w", err)
		}
		cfg.GitLab.Timeout = parsed
	}

	if l.flags.Project != "" {
		cfg.Validation.Project = l.flags.Project
	}

	if l.flags.SkipAPI {
		cfg.Validation.SkipAPI = true
	}

	if l.flags.Strict {
		cfg.Validation.Strict = true
	}

	if l.flags.Output != "" {
		cfg.Output.Format = l.flags.Output
	}

	if l.flags.Verbose {
		cfg.Output.Verbose = true
	}

	if l.flags.Color != "" {
		cfg.Output.Color = l.flags.Color
	}

	return cfg, nil
}

// mergeConfigs merges two configs, with over taking precedence
func (l *Loader) mergeConfigs(base, over Config) Config {
	result := base

	// Merge GitLab config
	if over.GitLab.Instance != "" {
		result.GitLab.Instance = over.GitLab.Instance
	}
	if over.GitLab.Timeout != nil {
		result.GitLab.Timeout = over.GitLab.Timeout
	}

	// Merge Auth config
	if over.Auth.Token != "" {
		result.Auth.Token = over.Auth.Token
	}
	if over.Auth.Netrc {
		result.Auth.Netrc = true
	}

	// Merge Validation config
	if over.Validation.Project != "" {
		result.Validation.Project = over.Validation.Project
	}
	if over.Validation.SkipAPI {
		result.Validation.SkipAPI = true
	}
	if over.Validation.Strict {
		result.Validation.Strict = true
	}

	// Merge Output config
	if over.Output.Format != "" {
		result.Output.Format = over.Output.Format
	}
	if over.Output.Verbose {
		result.Output.Verbose = true
	}
	if over.Output.Color != "" {
		result.Output.Color = over.Output.Color
	}

	// Merge Files config
	if over.Files.SearchParent {
		result.Files.SearchParent = true
	}
	if over.Files.MaxDepth > 0 {
		result.Files.MaxDepth = over.Files.MaxDepth
	}
	if len(over.Files.IgnorePatterns) > 0 {
		result.Files.IgnorePatterns = over.Files.IgnorePatterns
	}

	return result
}

// validate validates the final configuration
func (l *Loader) validate(cfg *Config) error {
	// Validate output format
	if cfg.Output.Format != "text" && cfg.Output.Format != "json" && cfg.Output.Format != "yaml" {
		return fmt.Errorf("invalid output format: %s (must be text, json, or yaml)", cfg.Output.Format)
	}

	// Validate color setting
	if cfg.Output.Color != "auto" && cfg.Output.Color != "always" && cfg.Output.Color != "never" {
		return fmt.Errorf("invalid color setting: %s (must be auto, always, or never)", cfg.Output.Color)
	}

	// Validate instance URL
	if cfg.GitLab.Instance == "" {
		return fmt.Errorf("gitlab instance cannot be empty")
	}

	return nil
}
