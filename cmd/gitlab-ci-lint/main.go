package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/InkyQuill/gitlab-ci-lint/internal/config"
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
		Use:   "gitlab-ci-lint [flags] <file>",
		Short: "GitLab CI/CD configuration linter",
		Long: `GitLab CI/CD configuration linter validates .gitlab-ci.yml files
in two stages:
1. Local YAML syntax validation
2. GitLab API validation (can be disabled with --skip-api)`,
		Args: cobra.ExactArgs(1),
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
	filePath := args[0]

	// Load configuration
	loader := config.NewLoader(flags)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(exit.ExitGeneralError)
	}

	// Read CI config file
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file '%s': %v\n", filePath, err)
		os.Exit(exit.ExitGeneralError)
	}

	// Stage 1: Local YAML validation
	localValidator := validator.NewLocalValidator(cfg.Validation.Strict)
	localResult := localValidator.Validate(content)

	// Create formatter
	formatter := output.NewFormatter(cfg.Output.Color, cfg.Output.Verbose)

	// If local validation failed, print errors and exit
	if !localResult.Valid {
		formatter.FormatResult(os.Stdout, cfg.Output.Format, []validator.Result{localResult}, filepath.Base(filePath))
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
		formatter.FormatResult(os.Stdout, cfg.Output.Format, results, filepath.Base(filePath))

		// Exit with appropriate code
		if !apiResult.Valid {
			os.Exit(exit.ExitValidationError)
		}
	} else {
		// Only local validation
		formatter.FormatResult(os.Stdout, cfg.Output.Format, results, filepath.Base(filePath))
	}

	// Success
	os.Exit(exit.ExitSuccess)
}
