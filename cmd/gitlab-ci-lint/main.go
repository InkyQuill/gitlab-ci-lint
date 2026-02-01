package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/InkyQuill/gitlab-ci-lint/internal/config"
	"github.com/InkyQuill/gitlab-ci-lint/internal/discover"
	"github.com/InkyQuill/gitlab-ci-lint/internal/exit"
	"github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"
	"github.com/InkyQuill/gitlab-ci-lint/internal/output"
	"github.com/InkyQuill/gitlab-ci-lint/internal/setup"
	"github.com/InkyQuill/gitlab-ci-lint/internal/validator"
	"github.com/InkyQuill/gitlab-ci-lint/pkg/version"
	"github.com/spf13/cobra"
)

var (
	flags        = &config.ConfigFlags{}
	setupForce   bool
	flagListInstances bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gitlab-ci-lint [flags] [file]",
		Short: "GitLab CI/CD configuration linter",
		Long: `GitLab CI/CD configuration linter validates .gitlab-ci.yml files
in two stages:
1. Local YAML syntax validation
2. GitLab API validation (can be disabled with --skip-api)

File discovery (if no file specified):
  - Searches current directory, then parent directories
  - Use -f to specify one or more files
  - Use -d to recursively scan directories
  - Use - as filename to read from stdin`,
		Args: cobra.MaximumNArgs(1),
		Run:  runLint,
	}

	// Configuration flags
	rootCmd.Flags().StringVarP(&flags.ConfigFile, "config", "c", "", "Path to config file")
	rootCmd.Flags().StringVarP(&flags.Token, "token", "t", "", "GitLab personal access token")
	rootCmd.Flags().BoolVar(&flags.Netrc, "netrc", false, "Use .netrc for credentials")
	rootCmd.Flags().StringVar(&flags.Instance, "instance", "", "GitLab instance URL")
	rootCmd.Flags().StringVar(&flags.Timeout, "timeout", "", "API timeout (e.g., 30s)")
	rootCmd.Flags().StringVar(&flags.Project, "project", "", "Project ID (e.g., '123' or 'group/project'). Overrides auto-detection from .git/config")
	rootCmd.Flags().BoolVarP(&flags.SkipAPI, "skip-api", "s", false, "Skip API validation")
	rootCmd.Flags().BoolVar(&flags.Strict, "strict", false, "Use strict YAML parsing")
	rootCmd.Flags().StringVarP(&flags.Output, "output", "o", "", "Output format: text|json|yaml")
	rootCmd.Flags().BoolVar(&flags.Debug, "debug", false, "Debug output (show API requests, project detection)")
	rootCmd.Flags().CountVarP(&flags.Verbosity, "verbose", "v", "Verbose output (use -vv for debug)")
	rootCmd.Flags().StringVar(&flags.Color, "color", "", "Color output: auto|always|never")
	// File discovery flags
	rootCmd.Flags().StringSliceVarP(&flags.Files, "file", "f", []string{}, "Path(s) to CI file(s). Can be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&flags.Directories, "directory", "d", []string{}, "Directory path(s) to recursively search for CI files.")

	// Multi-instance flags
	rootCmd.Flags().BoolVar(&flagListInstances, "list-instances", false,
		"List all configured GitLab instances and exit")

	// Version flag
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gitlab-ci-lint version %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Build date: %s\n", version.BuildDate)
		},
	}

	rootCmd.AddCommand(versionCmd)

	// Setup command
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		Long:  "Interactive configuration wizard for gitlab-ci-lint",
		RunE:  runSetup,
	}
	setupCmd.Flags().BoolVarP(&setupForce, "force", "f", false,
		"Overwrite existing configuration without asking")

	rootCmd.AddCommand(setupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}
}

// prepareGitLabRegistry prepares and validates GitLab clients for all configured instances
func prepareGitLabRegistry(ctx context.Context, cfg *config.Config, formatter *output.Formatter) (*gitlab.ClientRegistry, error) {
	if cfg.Validation.SkipAPI {
		return nil, nil
	}

	// Debug logging
	if formatter.GetDebugLogger() != nil {
		formatter.GetDebugLogger().LogSection("GitLab Client Registry")
	}

	// Check if we have instances configured
	if len(cfg.GitLab.Instances) == 0 {
		// No instances configured - warn user
		if cfg.Output.Verbose {
			formatter.FormatMessage(os.Stderr, "warning",
				"No GitLab instances configured. Run 'gitlab-ci-lint setup' to configure instances.")
		}
		if formatter.GetDebugLogger() != nil {
			formatter.GetDebugLogger().Log("REGISTRY", "no instances configured, skipping API validation")
		}
		return nil, nil
	}

	// Create registry from configured instances
	timeout := cfg.GitLab.GetTimeout()
	instances := convertConfigInstancesToRegistry(cfg.GitLab.Instances)
	registry := gitlab.NewClientRegistry(instances, timeout, formatter.GetDebugLogger())

	// Validate all tokens in verbose mode
	if cfg.Output.Verbose {
		formatter.FormatMessage(os.Stdout, "info", "Validating GitLab tokens...")
	}

	validationResults := registry.ValidateAllTokens(ctx)
	hasErrors := false
	for name, err := range validationResults {
		if err != nil {
			hasErrors = true
			if formatter.GetDebugLogger() != nil {
				formatter.GetDebugLogger().Log("REGISTRY",
					fmt.Sprintf("%s token validation FAILED: %v", name, err))
			}
			if cfg.Output.Verbose {
				formatter.FormatMessage(os.Stderr, "warning",
					fmt.Sprintf("Token validation failed for instance '%s': %v", name, err))
			}
		} else {
			if formatter.GetDebugLogger() != nil {
				formatter.GetDebugLogger().Log("REGISTRY",
					fmt.Sprintf("%s token validation: success", name))
			}
		}
	}

	if hasErrors && cfg.Output.Verbose {
		formatter.FormatMessage(os.Stderr, "warning",
			"Some tokens failed validation. API validation may not work for those instances.")
	}

	return registry, nil
}

// Note: min function is currently unused but kept for potential future use
//nolint:unused // Function kept for potential future use
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// listInstancesAction lists all configured GitLab instances
func listInstancesAction(cfg *config.Config) error {
	fmt.Println("\nConfigured GitLab Instances:")
	fmt.Println()

	if len(cfg.GitLab.Instances) == 0 {
		fmt.Println("  (none)")
		fmt.Println("\nRun 'gitlab-ci-lint setup' to configure instances")
		return nil
	}

	for _, inst := range cfg.GitLab.Instances {
		fmt.Printf("  %s\n", inst.Name)
		fmt.Printf("    URL: %s\n", inst.URL)
		if inst.Token != "" {
			fmt.Printf("    Token: glpat-***%s\n", safeTruncate(inst.Token, 4))
		} else {
			fmt.Printf("    Token: (none)\n")
		}
		if inst.Timeout != nil {
			fmt.Printf("    Timeout: %s\n", inst.Timeout.Duration)
		}
		fmt.Println()
	}

	return nil
}

// convertConfigInstancesToRegistry converts config.InstanceConfig to gitlab.InstanceConfig
func convertConfigInstancesToRegistry(instances []config.InstanceConfig) []gitlab.InstanceConfig {
	result := make([]gitlab.InstanceConfig, 0, len(instances))
	for _, inst := range instances {
		var timeout *time.Duration
		if inst.Timeout != nil {
			timeout = &inst.Timeout.Duration
		}
		result = append(result, gitlab.InstanceConfig{
			Name:    inst.Name,
			URL:     inst.URL,
			Token:   inst.Token,
			Timeout: timeout,
		})
	}
	return result
}

func runLint(cmd *cobra.Command, args []string) {
	// Handle --list-instances flag
	if flagListInstances {
		loader := config.NewLoader(flags)
		cfg, err := loader.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			os.Exit(exit.ExitGeneralError)
		}
		if err := listInstancesAction(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(exit.ExitGeneralError)
		}
		os.Exit(exit.ExitSuccess)
	}

	// Handle stdin input
	if len(args) == 1 && args[0] == "-" {
		handleStdinInput(cmd)
		return
	}

	// Load configuration
	loader := config.NewLoader(flags)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}

	// Create formatter with debug support
	var formatter *output.Formatter
	if cfg.Output.Debug {
		formatter = output.NewFormatterWithDebug(cfg.Output.Color, cfg.Output.Verbose, os.Stderr)
	} else {
		formatter = output.NewFormatter(cfg.Output.Color, cfg.Output.Verbose)
	}

	// Display configuration warnings if verbose
	if len(loader.GetWarnings()) > 0 && cfg.Output.Verbose {
		for _, warn := range loader.GetWarnings() {
			formatter.FormatMessage(os.Stderr, "warning", warn)
		}
	}

	// Prepare GitLab client registry ONCE (before file processing)
	registry, err := prepareGitLabRegistry(cmd.Context(), cfg, formatter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to prepare GitLab client registry: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}

	// Discover files to validate
	discoverer := discover.NewDiscoverer()
	if cfg.Files.MaxDepth > 0 {
		discoverer.SetMaxDepth(cfg.Files.MaxDepth)
	}
	if len(cfg.Files.IgnorePatterns) > 0 {
		discoverer.SetIgnorePatterns(cfg.Files.IgnorePatterns)
	}

	files, err := discoverFiles(discoverer, cfg, args, formatter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering files: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No CI configuration files found\n")
		os.Exit(exit.ExitGeneralError)
	}

	// Track results across all files
	var allResults []validator.FileResult
	anyInvalid := false

	// Process each file
	for _, filePath := range files {
		fileResult := validateFile(cmd, filePath, cfg, formatter, registry)
		if fileResult != nil {
			allResults = append(allResults, *fileResult)
			if !fileResult.Valid {
				anyInvalid = true
			}
		}
	}

	// Print summary if processing multiple files
	if len(allResults) > 1 {
		formatter.FormatSummary(os.Stdout, cfg.Output.Format, allResults)
	}

	// Exit with appropriate code
	if anyInvalid {
		os.Exit(exit.ExitValidationError)
	}
	os.Exit(exit.ExitSuccess)
}

// handleStdinInput processes CI configuration from stdin
func handleStdinInput(cmd *cobra.Command) {
	// Load configuration
	loader := config.NewLoader(flags)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}

	// Create formatter with debug support
	var formatter *output.Formatter
	if cfg.Output.Debug {
		formatter = output.NewFormatterWithDebug(cfg.Output.Color, cfg.Output.Verbose, os.Stderr)
	} else {
		formatter = output.NewFormatter(cfg.Output.Color, cfg.Output.Verbose)
	}

	// Display configuration warnings if verbose
	if len(loader.GetWarnings()) > 0 && cfg.Output.Verbose {
		for _, warn := range loader.GetWarnings() {
			formatter.FormatMessage(os.Stderr, "warning", warn)
		}
	}

	// Read from stdin
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read from stdin: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}

	// Stage 1: Local YAML validation
	localValidator := validator.NewLocalValidator(cfg.Validation.Strict)
	localResult := localValidator.Validate(content)

	// If local validation failed, print errors and exit
	if !localResult.Valid {
		formatter.FormatResult(os.Stdout, cfg.Output.Format, []validator.Result{localResult}, "stdin")
		os.Exit(exit.ExitValidationError)
	}

	// Stage 2: API validation (unless skipped)
	var results []validator.Result
	results = append(results, localResult)

	if !cfg.Validation.SkipAPI {
		// Stdin cannot be validated against GitLab API because:
		// 1. No .git/config to auto-detect instance/project
		// 2. No default instance fallback (removed in v2.0)
		// 3. Project path cannot be determined for stdin input
		//
		// Users must provide files explicitly for API validation.
		fmt.Fprintf(os.Stderr, "\nWarning: Stdin input cannot be validated against GitLab API.\n")
		fmt.Fprintf(os.Stderr, "  Reason: No .git/config for instance detection (stdin has no file path)\n")
		fmt.Fprintf(os.Stderr, "  Solution: Provide file path(s) explicitly: gitlab-ci-lint path/to/.gitlab-ci.yml\n")
		fmt.Fprintf(os.Stderr, "  Or use --skip-api flag to suppress this message\n")
		fmt.Fprintf(os.Stderr, "\nSkipping API validation. Only local YAML validation was performed.\n\n")
	}

	// Print all results
	formatter.FormatResult(os.Stdout, cfg.Output.Format, results, "stdin")

	// Success
	os.Exit(exit.ExitSuccess)
}

// discoverFiles orchestrates file discovery based on config and args
func discoverFiles(discoverer *discover.Discoverer, cfg *config.Config, args []string, formatter *output.Formatter) ([]string, error) {
	var files []string
	var allErrors []error

	// Debug logging
	if formatter.GetDebugLogger() != nil {
		formatter.GetDebugLogger().LogSection("File Discovery")
	}

	// Priority 1: -f flag (explicit files)
	if len(flags.Files) > 0 {
		if formatter.GetDebugLogger() != nil {
			formatter.GetDebugLogger().LogDiscovery(fmt.Sprintf("method=-f flag, count=%d", len(flags.Files)))
		}
		result, err := discoverer.ValidateAndExpandPaths(flags.Files)
		if err != nil {
			return nil, err
		}
		files = append(files, result.Files...)
		allErrors = append(allErrors, result.Errors...)
	}

	// Priority 2: -d flag (directories)
	if len(flags.Directories) > 0 {
		if formatter.GetDebugLogger() != nil {
			formatter.GetDebugLogger().LogDiscovery(fmt.Sprintf("method=-d flag, directories=%d", len(flags.Directories)))
		}
		for _, dir := range flags.Directories {
			result, err := discoverer.FindInDirectoryTree(dir)
			if err != nil {
				allErrors = append(allErrors, err)
				continue
			}
			files = append(files, result.Files...)
			allErrors = append(allErrors, result.Errors...)
		}
	}

	// Priority 3: Positional argument (backward compatibility)
	if len(args) == 1 {
		if formatter.GetDebugLogger() != nil {
			formatter.GetDebugLogger().LogDiscovery(fmt.Sprintf("method=positional arg, file=%s", args[0]))
		}
		result, err := discoverer.ValidateAndExpandPaths([]string{args[0]})
		if err != nil {
			return nil, err
		}
		files = append(files, result.Files...)
		allErrors = append(allErrors, result.Errors...)
	}

	// Priority 4: Auto-discovery (no args)
	if len(files) == 0 && len(args) == 0 && len(flags.Files) == 0 && len(flags.Directories) == 0 {
		if formatter.GetDebugLogger() != nil {
			formatter.GetDebugLogger().LogDiscovery("method=auto-discovery (search current + parent dirs)")
		}
		if cfg.Files.SearchParent {
			result, err := discoverer.FindInCurrentAndParents()
			if err != nil {
				return nil, err
			}
			files = append(files, result.Files...)
		}
	}

	if formatter.GetDebugLogger() != nil {
		formatter.GetDebugLogger().LogDiscovery(fmt.Sprintf("total files found=%d", len(files)))
	}

	// Print any discovery errors in verbose mode
	if cfg.Output.Verbose && len(allErrors) > 0 {
		for _, err := range allErrors {
			formatter.FormatMessage(os.Stderr, "warning", err.Error())
		}
	}

	return files, nil
}

// validateFile validates a single file with per-instance routing
func validateFile(cmd *cobra.Command, filePath string, cfg *config.Config, formatter *output.Formatter, registry *gitlab.ClientRegistry) *validator.FileResult {
	// Read CI config file
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file '%s': %v\n", filePath, err)
		return nil
	}

	// Create file result
	result := &validator.FileResult{
		FilePath: filePath,
		Valid:    true,
	}

	// Stage 1: Local YAML validation
	if formatter.GetDebugLogger() != nil {
		formatter.GetDebugLogger().LogValidate(fmt.Sprintf("Stage 1: Local validation for %s", filePath))
	}
	localValidator := validator.NewLocalValidator(cfg.Validation.Strict)
	localResult := localValidator.Validate(content)
	result.Stages = append(result.Stages, localResult)

	// If local validation failed, mark as invalid but continue to API if configured
	if !localResult.Valid {
		result.Valid = false
	}

	// Stage 2: API validation with per-file routing (unless skipped)
	if !cfg.Validation.SkipAPI && registry != nil {
		// Detect instance and project from .git/config
		instanceURL, projectPath, err := gitlab.DetectInstanceForFile(filePath)

		// Skip API validation if not in a git repository or no instance detected
		if err != nil || instanceURL == "" {
			if formatter.GetDebugLogger() != nil {
				formatter.GetDebugLogger().Log("ROUTE",
					fmt.Sprintf("file=%s no instance detected - skipping API validation", filePath))
			}
			formatter.FormatResult(os.Stdout, cfg.Output.Format, result.Stages, filepath.Base(filePath))
			return result
		}

		// Check if we have a token for this instance
		client, err := registry.GetClient(instanceURL)
		if err != nil {
			// No token configured for this instance - skip API validation gracefully
			if formatter.GetDebugLogger() != nil {
				formatter.GetDebugLogger().Log("ROUTE",
					fmt.Sprintf("file=%s no token for %s - skipping API validation",
						filePath, instanceURL))
			}
			formatter.FormatResult(os.Stdout, cfg.Output.Format, result.Stages, filepath.Base(filePath))
			return result
		}

		// Debug logging for routing
		instanceName := getInstanceName(registry, instanceURL)
		if formatter.GetDebugLogger() != nil {
			formatter.GetDebugLogger().Log("ROUTE",
				fmt.Sprintf("file=%s instance=%s (%s) project=%s",
					filePath, instanceName, instanceURL, projectPath))
		}

		// Run API validation
		if formatter.GetDebugLogger() != nil {
			formatter.GetDebugLogger().LogValidate("Stage 2: API validation")
		}
		apiValidator := validator.NewAPIValidator(client, projectPath)
		apiValidator.SetDebugLogger(formatter.GetDebugLogger())
		apiResult := apiValidator.Validate(content)
		result.Stages = append(result.Stages, apiResult)

		if !apiResult.Valid {
			result.Valid = false
		}
	}

	// Print results for this file
	formatter.FormatResult(os.Stdout, cfg.Output.Format, result.Stages, filepath.Base(filePath))

	return result
}

// detectFileContext determines the instance URL, project path, and instance name for a file
// Priority: .git/config (auto-detect) > default instance
// Returns (instanceURL, projectPath, instanceName)
// getInstanceName finds the instance name for a given URL
func getInstanceName(registry *gitlab.ClientRegistry, url string) string {
	instances := registry.GetAllInstances()
	normalizedURL := gitlab.NormalizeInstanceURL(url)
	for _, inst := range instances {
		if gitlab.NormalizeInstanceURL(inst.URL) == normalizedURL {
			return inst.Name
		}
	}
	return "unknown"
}

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("Welcome to GitLab CI Lint setup!")
	fmt.Println("This wizard will help you configure gitlab-ci-lint.")
	fmt.Println()

	// Get default config path
	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		return fmt.Errorf("failed to determine config path: %w", err)
	}

	// Check for existing configuration
	existingConfig := false
	if config.ConfigExists(configPath) && !setupForce {
		existingConfig = true
		var modifyOption string
		prompt := &survey.Select{
			Message: "Configuration file already exists. What would you like to do?",
			Options: []string{"Modify existing configuration", "Create new configuration (overwrite)", "Cancel"},
			Default: "Modify existing configuration",
		}
		if err := survey.AskOne(prompt, &modifyOption); err != nil {
			return err
		}

		if modifyOption == "Cancel" {
			fmt.Println("Setup cancelled.")
			return nil
		}

		if modifyOption == "Create new configuration (overwrite)" {
			setupForce = true
		}
	}

	// Load existing config if we're modifying
	var cfg config.Config
	if existingConfig && !setupForce {
		loader := config.NewLoader(&config.ConfigFlags{ConfigFile: configPath})
		loadedCfg, err := loader.Load()
		if err != nil {
			fmt.Printf("Warning: Failed to load existing config: %v\n", err)
			fmt.Println("Starting with default configuration...")
			cfg = config.GetDefaults()
		} else {
			cfg = *loadedCfg
			// Apply migration in case it's an old config
			config.MigrateLegacyConfig(&cfg)
		}
	} else {
		cfg = config.GetDefaults()
	}

	// Instance management loop
	shouldContinueToOutput := false
	shouldSaveAndExit := false
	for {
		fmt.Println("\n--- GitLab Instance Configuration ---")
		displayInstances(cfg.GitLab.Instances)

		var action string
		actionPrompt := &survey.Select{
			Message: "What would you like to do?",
			Options: []string{
				"Add new instance",
				"Edit existing instance",
				"Remove instance",
				"Test all instances",
				"Continue to output configuration",
				"Save and exit",
			},
		}
		if len(cfg.GitLab.Instances) == 0 {
			// Force adding first instance
			action = "Add new instance"
			fmt.Println("  (You must configure at least one instance)")
		} else if err := survey.AskOne(actionPrompt, &action); err != nil {
			return err
		}

		switch action {
		case "Add new instance":
			if err := addInstance(cmd.Context(), &cfg); err != nil {
				return err
			}
		case "Edit existing instance":
			if err := editInstance(cmd.Context(), &cfg); err != nil {
				return err
			}
		case "Remove instance":
			if err := removeInstance(&cfg); err != nil {
				return err
			}
		case "Test all instances":
			if err := testAllInstances(cmd.Context(), &cfg); err != nil {
				return err
			}
		case "Continue to output configuration":
			if len(cfg.GitLab.Instances) == 0 {
				fmt.Println("\nError: At least one instance must be configured")
				continue
			}
			shouldContinueToOutput = true
		case "Save and exit":
			if len(cfg.GitLab.Instances) == 0 {
				fmt.Println("\nError: At least one instance must be configured")
				continue
			}
			shouldSaveAndExit = true
		}

		if shouldContinueToOutput || shouldSaveAndExit {
			break
		}
	}

	// Output configuration (only if continuing, not saving immediately)
	var outputFormat string
	var verbose bool

	if shouldContinueToOutput {
		fmt.Println("\n--- Output Configuration ---")

		outputPrompt := &survey.Select{
			Message: "Default output format:",
			Options: []string{"text", "json", "yaml"},
			Default: cfg.Output.Format,
			Help:    "Choose how validation results should be displayed",
		}
		if err := survey.AskOne(outputPrompt, &outputFormat); err != nil {
			return err
		}
		cfg.Output.Format = outputFormat

		verbosePrompt := &survey.Confirm{
			Message: "Enable verbose output by default?",
			Default: cfg.Output.Verbose,
		}
		if err := survey.AskOne(verbosePrompt, &verbose); err != nil {
			return err
		}
		cfg.Output.Verbose = verbose
	}

	// Review configuration
	fmt.Println("\n--- Configuration Summary ---")
	fmt.Printf("  Configured Instances:\n")
	for _, inst := range cfg.GitLab.Instances {
		tokenPreview := ""
		if inst.Token != "" {
			tokenPreview = fmt.Sprintf(" (glpat-***%s)", safeTruncate(inst.Token, 4))
		}
		fmt.Printf("    - %s: %s%s\n", inst.Name, inst.URL, tokenPreview)
	}
	fmt.Printf("  Output Format: %s\n", cfg.Output.Format)
	fmt.Printf("  Verbose: %t\n", cfg.Output.Verbose)
	fmt.Printf("  Config Location: %s\n", configPath)
	fmt.Println("\n  Note: Projects will be auto-detected from .git/config.")
	fmt.Println("  Override with --project flag or GCL_PROJECT environment variable.")

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Save this configuration?",
		Default: true,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		fmt.Println("Configuration not saved.")
		return nil
	}

	// Clean up legacy fields before saving
	cleanupLegacyFields(&cfg)

	// Write configuration
	fmt.Println("\nSaving configuration...")
	if err := config.WriteConfig(configPath, &cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("\n✓ Configuration saved to: %s\n", configPath)

	// Test the configuration
	fmt.Println("\n--- Testing Configuration ---")
	var runTest bool
	testPrompt := &survey.Confirm{
		Message: "Would you like to test the configuration now?",
		Default: true,
	}
	if err := survey.AskOne(testPrompt, &runTest); err != nil {
		return err
	}

	if runTest {
		if err := testAllInstances(cmd.Context(), &cfg); err != nil {
			fmt.Printf("\nWarning: Configuration test failed: %v\n", err)
			fmt.Println("You may need to adjust your configuration.")
		} else {
			fmt.Println("\n✓ Configuration test passed!")
		}
	}

	fmt.Println("\nSetup complete!")
	fmt.Println("You can now use 'gitlab-ci-lint' to validate your CI configurations.")
	fmt.Println("Use 'gitlab-ci-lint --list-instances' to see all configured instances.")

	return nil
}

// displayInstances displays all configured instances
func displayInstances(instances []config.InstanceConfig) {
	if len(instances) == 0 {
		fmt.Println("  (none)")
		return
	}

	for i, inst := range instances {
		tokenPreview := ""
		if inst.Token != "" {
			tokenPreview = fmt.Sprintf(" (glpat-***%s)", safeTruncate(inst.Token, 4))
		}
		fmt.Printf("  %d. %s: %s%s\n", i+1, inst.Name, inst.URL, tokenPreview)
	}
}

// addInstance adds a new GitLab instance
func addInstance(ctx context.Context, cfg *config.Config) error {
	var name, url, token string

	// Name
	namePrompt := &survey.Input{
		Message: "Instance name (e.g., gitlab.com, work):",
		Help:    "A unique identifier for this instance",
	}
	if err := survey.AskOne(namePrompt, &name, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Check for duplicate
	for _, inst := range cfg.GitLab.Instances {
		if inst.Name == name {
			return fmt.Errorf("instance '%s' already exists", name)
		}
	}

	// URL
	urlPrompt := &survey.Input{
		Message: "Instance URL:",
		Default: "https://gitlab.com",
	}
	if err := survey.AskOne(urlPrompt, &url, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	url = gitlab.NormalizeInstanceURL(url)

	// Test connection
	fmt.Print("Testing connection...")
	if err := setup.TestConnection(ctx, url); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	fmt.Println(" ✓")

	// Token (optional)
	var useToken bool
	tokenPrompt := &survey.Confirm{
		Message: "Configure authentication token?",
		Default: true,
	}
	if err := survey.AskOne(tokenPrompt, &useToken); err != nil {
		return err
	}

	if useToken {
		maxRetries := 3
		for attempt := 0; attempt < maxRetries; attempt++ {
			var tokenInput string
			tokenInputPrompt := &survey.Password{
				Message: "Personal access token:",
				Help:    "Enter your GitLab personal access token (requires 'api' scope)",
			}
			if err := survey.AskOne(tokenInputPrompt, &tokenInput); err != nil {
				return err
			}

			if tokenInput == "" {
				break
			}

			// Validate token
			fmt.Print("Validating token...")
			result, err := setup.ValidateToken(url, tokenInput)
			if err != nil {
				return fmt.Errorf("token validation error: %w", err)
			}

			if result.Valid {
				fmt.Printf("\n  ✓ Token valid! Authenticated as: %s\n", result.Username)
				token = tokenInput
				break
			} else {
				fmt.Printf("\n  ✗ Token validation failed: %s\n", result.Error)

				if attempt < maxRetries-1 {
					var retry bool
					retryPrompt := &survey.Confirm{
						Message: "Would you like to re-enter the token?",
						Default: true,
					}
					if err := survey.AskOne(retryPrompt, &retry); err != nil {
						return err
					}
					if !retry {
						break
					}
				}
			}
		}
	}

	// Add instance
	instance := config.InstanceConfig{
		Name:  name,
		URL:   url,
		Token: token,
	}
	cfg.GitLab.Instances = append(cfg.GitLab.Instances, instance)

	fmt.Printf("\n✓ Instance '%s' added successfully\n", name)
	return nil
}

// editInstance edits an existing instance
func editInstance(ctx context.Context, cfg *config.Config) error {
	if len(cfg.GitLab.Instances) == 0 {
		return fmt.Errorf("no instances to edit")
	}

	var instanceNames []string
	for _, inst := range cfg.GitLab.Instances {
		instanceNames = append(instanceNames, inst.Name)
	}

	var selected int
	selectPrompt := &survey.Select{
		Message: "Select instance to edit:",
		Options: instanceNames,
	}
	if err := survey.AskOne(selectPrompt, &selected); err != nil {
		return err
	}

	inst := &cfg.GitLab.Instances[selected]

	// Edit URL
	var newURL string
	urlPrompt := &survey.Input{
		Message: "Instance URL:",
		Default: inst.URL,
	}
	if err := survey.AskOne(urlPrompt, &newURL, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	inst.URL = gitlab.NormalizeInstanceURL(newURL)

	// Edit token
	var editToken bool
	editTokenPrompt := &survey.Confirm{
		Message: "Edit authentication token?",
		Default: false,
	}
	if err := survey.AskOne(editTokenPrompt, &editToken); err != nil {
		return err
	}

	if editToken {
		maxRetries := 3
		for attempt := 0; attempt < maxRetries; attempt++ {
			var newToken string
			tokenPrompt := &survey.Password{
				Message: "Personal access token:",
				Help:    "Enter your GitLab personal access token (requires 'api' scope)",
			}
			if err := survey.AskOne(tokenPrompt, &newToken); err != nil {
				return err
			}

			if newToken == "" {
				inst.Token = ""
				break
			}

			// Validate token
			fmt.Print("Validating token...")
			result, err := setup.ValidateToken(inst.URL, newToken)
			if err != nil {
				return fmt.Errorf("token validation error: %w", err)
			}

			if result.Valid {
				fmt.Printf("\n  ✓ Token valid! Authenticated as: %s\n", result.Username)
				inst.Token = newToken
				break
			} else {
				fmt.Printf("\n  ✗ Token validation failed: %s\n", result.Error)

				if attempt < maxRetries-1 {
					var retry bool
					retryPrompt := &survey.Confirm{
						Message: "Would you like to re-enter the token?",
						Default: true,
					}
					if err := survey.AskOne(retryPrompt, &retry); err != nil {
						return err
					}
					if !retry {
						break
					}
				}
			}
		}
	}

	fmt.Printf("\n✓ Instance '%s' updated successfully\n", inst.Name)
	return nil
}

// removeInstance removes an instance
func removeInstance(cfg *config.Config) error {
	if len(cfg.GitLab.Instances) == 0 {
		return fmt.Errorf("no instances to remove")
	}

	var instanceNames []string
	for _, inst := range cfg.GitLab.Instances {
		instanceNames = append(instanceNames, inst.Name)
	}

	var selected int
	selectPrompt := &survey.Select{
		Message: "Select instance to remove:",
		Options: instanceNames,
	}
	if err := survey.AskOne(selectPrompt, &selected); err != nil {
		return err
	}

	inst := cfg.GitLab.Instances[selected]

	// Confirm removal
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Remove instance '%s'?", inst.Name),
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		fmt.Println("Removal cancelled.")
		return nil
	}

	// Remove instance
	cfg.GitLab.Instances = append(cfg.GitLab.Instances[:selected], cfg.GitLab.Instances[selected+1:]...)

	fmt.Printf("✓ Instance '%s' removed successfully\n", inst.Name)
	return nil
}

// testAllInstances tests all configured instances
func testAllInstances(ctx context.Context, cfg *config.Config) error {
	if len(cfg.GitLab.Instances) == 0 {
		return fmt.Errorf("no instances to test")
	}

	fmt.Println("\nTesting all instances...")

	allPassed := true
	for _, inst := range cfg.GitLab.Instances {
		fmt.Printf("\n  Testing '%s' (%s)...\n", inst.Name, inst.URL)

		// Test connection
		if err := setup.TestConnection(ctx, inst.URL); err != nil {
			fmt.Printf("    ✗ Connection failed: %v\n", err)
			allPassed = false
			continue
		}
		fmt.Println("    ✓ Connection successful")

		// Test token if configured
		if inst.Token != "" {
			result, err := setup.ValidateToken(inst.URL, inst.Token)
			if err != nil {
				fmt.Printf("    ✗ Token validation error: %v\n", err)
				allPassed = false
				continue
			}

			if !result.Valid {
				fmt.Printf("    ✗ Token validation failed: %s\n", result.Error)
				allPassed = false
				continue
			}

			fmt.Printf("    ✓ Token valid (authenticated as: %s)\n", result.Username)
		} else {
			fmt.Println("    ⚠ No token configured")
		}
	}

	if !allPassed {
		return fmt.Errorf("some instances failed validation")
	}

	return nil
}

// cleanupLegacyFields removes deprecated fields before saving
func cleanupLegacyFields(cfg *config.Config) {
	// Clear legacy single-instance fields
	cfg.GitLab.Instance = ""
	cfg.Auth.Token = ""
}

// safeTruncate returns the last n characters of a string, or the whole string if shorter
func safeTruncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
