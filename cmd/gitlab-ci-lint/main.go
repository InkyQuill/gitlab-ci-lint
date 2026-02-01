package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/InkyQuill/gitlab-ci-lint/internal/config"
	"github.com/InkyQuill/gitlab-ci-lint/internal/discover"
	"github.com/InkyQuill/gitlab-ci-lint/internal/exit"
	"github.com/InkyQuill/gitlab-ci-lint/internal/gitlab"
	"github.com/InkyQuill/gitlab-ci-lint/internal/output"
	"github.com/InkyQuill/gitlab-ci-lint/internal/validator"
	"github.com/InkyQuill/gitlab-ci-lint/pkg/version"
	"github.com/spf13/cobra"
)

var (
	flags = &config.ConfigFlags{}
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
		Run: func(cmd *cobra.Command, args []string) {
			// Run the setup binary
			setupBin, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: unable to determine executable path: %v\n", err)
				os.Exit(exit.ExitGeneralError)
			}

			// Try to run gitlab-ci-lint-setup if it exists
			setupPath := filepath.Join(filepath.Dir(setupBin), "gitlab-ci-lint-setup")
			if _, err := os.Stat(setupPath); err == nil {
				// Setup binary exists, run it
				setupCmd := exec.Command(setupPath)
				setupCmd.Stdin = os.Stdin
				setupCmd.Stdout = os.Stdout
				setupCmd.Stderr = os.Stderr
				if err := setupCmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Setup command failed: %v\n", err)
					os.Exit(exit.ExitGeneralError)
				}
			} else {
				fmt.Println("Setup command not found. Please build the setup binary:")
				fmt.Println("  go build -o gitlab-ci-lint-setup ./cmd/setup")
				os.Exit(exit.ExitGeneralError)
			}
		},
	}

	rootCmd.AddCommand(setupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}
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
		fileResult := validateFile(cmd, filePath, cfg, formatter)
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
		// Normalize instance URL
		instance := gitlab.NormalizeInstanceURL(cfg.GitLab.Instance)

		// Get token from config or environment
		token := cfg.Auth.Token
		if token == "" && cfg.Auth.Netrc {
			token, err = gitlab.ExtractTokenFromNetrc(instance)
			if err != nil && cfg.Output.Verbose {
				formatter.FormatMessage(os.Stdout, "warning", fmt.Sprintf("Failed to read .netrc: %v", err))
			}
		}

		// Create GitLab client
		client := gitlab.NewClient(instance, token, cfg.GitLab.Timeout)

		// Validate token (if provided)
		if token != "" {
			if cfg.Output.Verbose {
				formatter.FormatMessage(os.Stdout, "info", "Validating GitLab token...")
			}
			if err := client.ValidateToken(cmd.Context()); err != nil {
				fmt.Fprintf(os.Stderr, "Token validation failed: %v\n", err)
				os.Exit(exit.ExitGeneralError)
			}
		}

		// Run API validation
		apiValidator := validator.NewAPIValidator(client, cfg.Validation.Project)
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
func validateFile(cmd *cobra.Command, filePath string, cfg *config.Config, formatter *output.Formatter) *validator.FileResult {
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
	if !cfg.Validation.SkipAPI {
		// Normalize instance URL
		instance := gitlab.NormalizeInstanceURL(cfg.GitLab.Instance)

		// Get token from config or environment
		token := cfg.Auth.Token
		if token == "" && cfg.Auth.Netrc {
			var err error
			token, err = gitlab.ExtractTokenFromNetrc(instance)
			if err != nil && cfg.Output.Verbose {
				formatter.FormatMessage(os.Stderr, "warning", fmt.Sprintf("Failed to read .netrc: %v", err))
			}
		}

		// Create GitLab client
		client := gitlab.NewClient(instance, token, cfg.GitLab.Timeout)

		// Validate token (if provided) - only once
		// Note: Token validation is done in runLint for the first file, skipped for subsequent files

		// Run API validation
		apiValidator := validator.NewAPIValidator(client, cfg.Validation.Project)
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
