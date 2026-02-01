package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	flags      = &config.ConfigFlags{}
	setupForce bool
)

// GitLabClient holds the client and its configuration
type GitLabClient struct {
	client    *gitlab.Client
	token     string
	instance  string
	validated bool
}

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
	rootCmd.Flags().StringVar(&flags.Project, "project", "", "Project ID (e.g., '123' or 'group/project')")
	rootCmd.Flags().BoolVarP(&flags.SkipAPI, "skip-api", "s", false, "Skip API validation")
	rootCmd.Flags().BoolVar(&flags.Strict, "strict", false, "Use strict YAML parsing")
	rootCmd.Flags().StringVarP(&flags.Output, "output", "o", "", "Output format: text|json|yaml")
	rootCmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().StringVar(&flags.Color, "color", "", "Color output: auto|always|never")
	// File discovery flags
	rootCmd.Flags().StringSliceVarP(&flags.Files, "file", "f", []string{}, "Path(s) to CI file(s). Can be specified multiple times.")
	rootCmd.Flags().StringSliceVarP(&flags.Directories, "directory", "d", []string{}, "Directory path(s) to recursively search for CI files.")

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

// prepareGitLabClient prepares and validates a GitLab client
func prepareGitLabClient(ctx context.Context, cfg *config.Config, formatter *output.Formatter) (*GitLabClient, error) {
	if cfg.Validation.SkipAPI {
		return nil, nil
	}

	// Normalize instance URL
	instance := gitlab.NormalizeInstanceURL(cfg.GitLab.Instance)

	// Get token from config or netrc
	token := cfg.Auth.Token
	if token == "" && cfg.Auth.Netrc {
		var err error
		token, err = gitlab.ExtractTokenFromNetrc(instance)
		if err != nil && cfg.Output.Verbose {
			formatter.FormatMessage(os.Stderr, "warning", fmt.Sprintf("Failed to read .netrc: %v", err))
		}
	}

	// Create client (use timeout safely)
	timeout := 30 * time.Second
	if cfg.GitLab.Timeout != nil {
		timeout = *cfg.GitLab.Timeout
	}
	client := gitlab.NewClient(instance, token, timeout)

	// Validate token (if provided)
	if token != "" {
		if cfg.Output.Verbose {
			formatter.FormatMessage(os.Stdout, "info", "Validating GitLab token...")
		}
		if err := client.ValidateToken(ctx); err != nil {
			return nil, fmt.Errorf("token validation failed: %w", err)
		}
	}

	return &GitLabClient{
		client:    client,
		token:     token,
		instance:  instance,
		validated: token != "",
	}, nil
}

func runLint(cmd *cobra.Command, args []string) {
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

	// Create formatter
	formatter := output.NewFormatter(cfg.Output.Color, cfg.Output.Verbose)

	// Prepare GitLab client ONCE (before file processing)
	gitlabClient, err := prepareGitLabClient(cmd.Context(), cfg, formatter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to prepare GitLab client: %v\n", err)
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

	files, err := discoverFiles(discoverer, cfg, args)
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
		fileResult := validateFile(cmd, filePath, cfg, formatter, gitlabClient)
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

	// Create formatter
	formatter := output.NewFormatter(cfg.Output.Color, cfg.Output.Verbose)

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
		// Prepare GitLab client
		gitlabClient, err := prepareGitLabClient(cmd.Context(), cfg, formatter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to prepare GitLab client: %v\n", err)
			os.Exit(exit.ExitGeneralError)
		}

		// Run API validation
		apiValidator := validator.NewAPIValidator(gitlabClient.client, cfg.Validation.Project)
		apiResult := apiValidator.Validate(content)
		results = append(results, apiResult)

		// Print all results
		formatter.FormatResult(os.Stdout, cfg.Output.Format, results, "stdin")

		// Exit with appropriate code
		if !apiResult.Valid {
			os.Exit(exit.ExitValidationError)
		}
	} else {
		// Only local validation
		formatter.FormatResult(os.Stdout, cfg.Output.Format, results, "stdin")
	}

	// Success
	os.Exit(exit.ExitSuccess)
}

// discoverFiles orchestrates file discovery based on config and args
func discoverFiles(discoverer *discover.Discoverer, cfg *config.Config, args []string) ([]string, error) {
	var files []string
	var allErrors []error

	// Priority 1: -f flag (explicit files)
	if len(flags.Files) > 0 {
		result, err := discoverer.ValidateAndExpandPaths(flags.Files)
		if err != nil {
			return nil, err
		}
		files = append(files, result.Files...)
		allErrors = append(allErrors, result.Errors...)
	}

	// Priority 2: -d flag (directories)
	if len(flags.Directories) > 0 {
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
		result, err := discoverer.ValidateAndExpandPaths([]string{args[0]})
		if err != nil {
			return nil, err
		}
		files = append(files, result.Files...)
		allErrors = append(allErrors, result.Errors...)
	}

	// Priority 4: Auto-discovery (no args)
	if len(files) == 0 && len(args) == 0 && len(flags.Files) == 0 && len(flags.Directories) == 0 {
		if cfg.Files.SearchParent {
			result, err := discoverer.FindInCurrentAndParents()
			if err != nil {
				return nil, err
			}
			files = append(files, result.Files...)
		}
	}

	// Print any discovery errors in verbose mode
	if cfg.Output.Verbose {
		for _, err := range allErrors {
			formatter := output.NewFormatter(cfg.Output.Color, true)
			formatter.FormatMessage(os.Stderr, "warning", err.Error())
		}
	}

	return files, nil
}

// validateFile validates a single file and returns the result
func validateFile(cmd *cobra.Command, filePath string, cfg *config.Config, formatter *output.Formatter, gitlabClient *GitLabClient) *validator.FileResult {
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
	localValidator := validator.NewLocalValidator(cfg.Validation.Strict)
	localResult := localValidator.Validate(content)
	result.Stages = append(result.Stages, localResult)

	// If local validation failed, mark as invalid but continue to API if configured
	if !localResult.Valid {
		result.Valid = false
	}

	// Stage 2: API validation (unless skipped)
	if !cfg.Validation.SkipAPI && gitlabClient != nil && gitlabClient.client != nil {
		// Run API validation using shared client
		apiValidator := validator.NewAPIValidator(gitlabClient.client, cfg.Validation.Project)
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
		}
	} else {
		cfg = config.GetDefaults()
	}

	// Step 1: GitLab Instance URL
	fmt.Println("\n--- GitLab Instance Configuration ---")
	var instanceURL string
	instancePrompt := &survey.Input{
		Message: "GitLab instance URL:",
		Default: cfg.GitLab.Instance,
		Help:    "Enter your GitLab instance URL (e.g., https://gitlab.com or https://gitlab.example.com)",
	}
	if err := survey.AskOne(instancePrompt, &instanceURL, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	cfg.GitLab.Instance = instanceURL

	// Detect instance type
	fmt.Print("Detecting instance type...")
	instanceInfo, err := setup.DetectInstanceType(instanceURL)
	if err != nil {
		fmt.Printf("\nWarning: Could not detect instance: %v\n", err)
	} else {
		fmt.Printf("\n  Detected: %s", instanceInfo.InstanceName)
		if instanceInfo.Version != "" {
			fmt.Printf(" (version: %s)", instanceInfo.Version)
		}
		fmt.Println()
	}

	// Step 2: Authentication
	fmt.Println("\n--- Authentication Configuration ---")

	var useToken bool
	useTokenPrompt := &survey.Confirm{
		Message: "Do you want to use a personal access token?",
		Default: cfg.Auth.Token != "",
	}
	if err := survey.AskOne(useTokenPrompt, &useToken); err != nil {
		return err
	}

	if useToken {
		maxRetries := 3
		for attempt := 0; attempt < maxRetries; attempt++ {
			var token string
			tokenPrompt := &survey.Password{
				Message: "Personal access token:",
				Help:    "Enter your GitLab personal access token (requires 'read_api' scope)",
			}
			if err := survey.AskOne(tokenPrompt, &token); err != nil {
				return err
			}

			if token == "" {
				// User entered empty token, skip
				break
			}

			// Validate token
			fmt.Print("Validating token...")
			result, err := setup.ValidateToken(instanceURL, token)
			if err != nil {
				return fmt.Errorf("token validation error: %w", err)
			}

			if result.Valid {
				fmt.Printf("\n  ✓ Token valid! Authenticated as: %s\n", result.Username)
				cfg.Auth.Token = token
				break // Success, exit retry loop
			} else {
				fmt.Printf("\n  ✗ Token validation failed: %s\n", result.Error)

				// Ask user if they want to retry (unless this was the last attempt)
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
						cfg.Auth.Token = ""
						break // User chose not to retry
					}
					// Otherwise, continue to next attempt
				} else {
					fmt.Println("  Maximum retry attempts reached.")
					cfg.Auth.Token = ""
				}
			}
		}
	}

	// Step 3: Default project (optional)
	fmt.Println("\n--- Project Configuration ---")
	var useProject bool
	useProjectPrompt := &survey.Confirm{
		Message: "Do you want to set a default project?",
		Default: cfg.Validation.Project != "",
		Help:    "Setting a default project enables project-specific validation (e.g., job references)",
	}
	if err := survey.AskOne(useProjectPrompt, &useProject); err != nil {
		return err
	}

	if useProject {
		var project string
		projectPrompt := &survey.Input{
			Message: "Default project (ID or path):",
			Default: cfg.Validation.Project,
			Help:    "Enter project ID (e.g., '123') or path (e.g., 'group/project')",
		}
		if err := survey.AskOne(projectPrompt, &project); err != nil {
			return err
		}
		if strings.TrimSpace(project) != "" {
			cfg.Validation.Project = strings.TrimSpace(project)
		}
	}

	// Step 4: Output preferences
	fmt.Println("\n--- Output Configuration ---")

	var outputFormat string
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

	var verbose bool
	verbosePrompt := &survey.Confirm{
		Message: "Enable verbose output by default?",
		Default: cfg.Output.Verbose,
	}
	if err := survey.AskOne(verbosePrompt, &verbose); err != nil {
		return err
	}
	cfg.Output.Verbose = verbose

	// Step 5: Review configuration
	fmt.Println("\n--- Configuration Summary ---")
	fmt.Printf("  GitLab Instance: %s\n", cfg.GitLab.Instance)
	if cfg.Auth.Token != "" {
		fmt.Printf("  Authentication: Token configured (glpat-***%s)\n", safeTruncate(cfg.Auth.Token, 4))
	} else {
		fmt.Printf("  Authentication: None (will use .netrc or prompt)")
	}
	if cfg.Validation.Project != "" {
		fmt.Printf("  Default Project: %s\n", cfg.Validation.Project)
	}
	fmt.Printf("  Output Format: %s\n", cfg.Output.Format)
	fmt.Printf("  Verbose: %t\n", cfg.Output.Verbose)
	fmt.Printf("  Config Location: %s\n", configPath)

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
		if err := testConfiguration(&cfg); err != nil {
			fmt.Printf("\nWarning: Configuration test failed: %v\n", err)
			fmt.Println("You may need to adjust your configuration.")
		} else {
			fmt.Println("\n✓ Configuration test passed!")
		}
	}

	fmt.Println("\nSetup complete!")
	fmt.Println("You can now use 'gitlab-ci-lint' to validate your CI configurations.")

	return nil
}

func testConfiguration(cfg *config.Config) error {
	fmt.Println("Testing connection to GitLab instance...")

	ctx := context.Background()
	if err := setup.TestConnection(ctx, cfg.GitLab.Instance); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	fmt.Println("✓ Connection successful")

	if cfg.Auth.Token != "" {
		fmt.Println("Testing authentication...")
		result, err := setup.ValidateToken(cfg.GitLab.Instance, cfg.Auth.Token)
		if err != nil {
			return fmt.Errorf("authentication test failed: %w", err)
		}
		if !result.Valid {
			return fmt.Errorf("authentication failed: %s", result.Error)
		}
		fmt.Printf("✓ Authentication successful (user: %s)\n", result.Username)
	}

	return nil
}

// safeTruncate returns the last n characters of a string, or the whole string if shorter
func safeTruncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
