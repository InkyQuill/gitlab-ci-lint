package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/InkyQuill/gitlab-ci-lint/internal/config"
	"github.com/InkyQuill/gitlab-ci-lint/internal/setup"
	"github.com/spf13/cobra"
)

var (
	forceOverwrite bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gitlab-ci-lint-setup",
		Short: "Interactive setup for gitlab-ci-lint",
		Long: `Interactive configuration wizard for gitlab-ci-lint.

This command guides you through setting up your GitLab CI Lint configuration,
including GitLab instance connection, authentication, and preferences.`,
		RunE: runSetup,
	}

	rootCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Overwrite existing configuration without asking")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
	if config.ConfigExists(configPath) && !forceOverwrite {
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
			forceOverwrite = true
		}
	}

	// Load existing config if we're modifying
	var cfg config.Config
	if existingConfig && !forceOverwrite {
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
		var token string
		tokenPrompt := &survey.Password{
			Message: "Personal access token:",
			Help:    "Enter your GitLab personal access token (requires 'read_api' scope)",
		}
		if err := survey.AskOne(tokenPrompt, &token); err != nil {
			return err
		}

		if token != "" {
			// Validate token
			fmt.Print("Validating token...")
			result, err := setup.ValidateToken(instanceURL, token)
			if err != nil {
				return fmt.Errorf("token validation error: %w", err)
			}

			if result.Valid {
				fmt.Printf("\n  ✓ Token valid! Authenticated as: %s\n", result.Username)
				cfg.Auth.Token = token
			} else {
				fmt.Printf("\n  ✗ Token validation failed: %s\n", result.Error)
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
				} else {
					// Retry token input
					return runSetup(cmd, args)
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
