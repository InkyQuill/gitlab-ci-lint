package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	mainBinary string
)

func TestMain(m *testing.M) {
	// Build binaries before running tests
	if err := buildBinaries(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func buildBinaries() error {
	// Find project root
	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	// Set binary path as absolute path
	buildDir := filepath.Join(root, "build")
	mainBinary = filepath.Join(buildDir, "gitlab-ci-lint")

	// Create build directory
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return err
	}

	// Build main binary (with integrated setup)
	mainCmd := exec.Command("go", "build", "-o", mainBinary, "./cmd/gitlab-ci-lint")
	mainCmd.Dir = root
	mainCmd.Stdout = os.Stdout
	mainCmd.Stderr = os.Stderr
	return mainCmd.Run()
}

// skipInCI skips interactive tests in CI environment
func skipInCI(t *testing.T) {
	t.Helper()
	if os.Getenv("CI") != "" {
		t.Skip("Skipping interactive test in CI environment")
	}
}

func TestSetupCommand_CreatesConfigFile(t *testing.T) {
	skipInCI(t)
	// Create temp directory for config
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")
	configFile := filepath.Join(configDir, "config.yaml")

	// Prepare input for interactive prompts
	input := "https://gitlab.com\n" + // GitLab instance
		"no\n" + // Use personal access token
		"no\n" + // Set default project
		"text\n" + // Output format
		"no\n" + // Enable verbose
		"yes\n" + // Save configuration
		"no\n" // Test configuration

	// Run setup command with environment variable for config path
	cmd := exec.Command(mainBinary, "setup")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	// Verify config file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configFile)
	}

	// Verify config file permissions
	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check file permissions (should be 0600 or similar - owner read/write only)
	mode := info.Mode()
	if mode.Perm()&0077 != 0 {
		t.Logf("Warning: Config file has loose permissions: %v", mode)
	}

	// Verify config file content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "gitlab.com") {
		t.Errorf("Config file does not contain expected instance URL")
	}
}

func TestSetupCommand_ModifyExistingConfig(t *testing.T) {
	skipInCI(t)
	// Create temp directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create existing config
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	existingConfig := `gitlab:
  instance: https://gitlab.example.com
  timeout: 30s
auth:
  token: ""
  netrc: false
validation:
  skip_api: false
  strict: true
  project: ""
output:
  format: text
  verbose: false
  color: auto
`
	if err := os.WriteFile(configFile, []byte(existingConfig), 0600); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	// Input: Modify existing, change output format to json
	input := "Modify existing configuration\n" +
		"json\n" + // Change output format
		"no\n" + // Verbose
		"yes\n" + // Save
		"no\n" // Test

	cmd := exec.Command(mainBinary, "setup")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	// Verify config was updated
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "format: json") {
		t.Errorf("Config format was not updated to json")
	}

	// Original instance should be preserved
	if !strings.Contains(contentStr, "gitlab.example.com") {
		t.Errorf("Original instance URL was not preserved")
	}
}

func TestSetupCommand_OverwriteExistingConfig(t *testing.T) {
	skipInCI(t)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create existing config
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	existingConfig := `gitlab:
  instance: https://old.example.com
`
	if err := os.WriteFile(configFile, []byte(existingConfig), 0600); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	// Input: Create new (overwrite), use new instance
	input := "Create new configuration (overwrite)\n" +
		"https://new.example.com\n" + // New instance
		"no\n" + // Token
		"no\n" + // Project
		"text\n" + // Format
		"no\n" + // Verbose
		"yes\n" + // Save
		"no\n" // Test

	cmd := exec.Command(mainBinary, "setup")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	// Verify config was overwritten
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "old.example.com") {
		t.Errorf("Old instance URL still present after overwrite")
	}

	if !strings.Contains(contentStr, "new.example.com") {
		t.Errorf("New instance URL not found after overwrite")
	}
}

func TestSetupCommand_CancelSetup(t *testing.T) {
	skipInCI(t)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create existing config
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	existingConfig := `gitlab:
  instance: https://preserve-me.com
`
	if err := os.WriteFile(configFile, []byte(existingConfig), 0600); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	// Input: Cancel
	input := "Cancel\n"

	cmd := exec.Command(mainBinary, "setup")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	// Verify existing config was not modified
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "preserve-me.com") {
		t.Errorf("Config was modified despite cancel")
	}

	if !strings.Contains(string(output), "Setup cancelled") {
		t.Error("Expected cancellation message in output")
	}
}

func TestSetupCommand_RejectSummary(t *testing.T) {
	skipInCI(t)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")
	configFile := filepath.Join(configDir, "config.yaml")

	// Input: Complete setup but reject at summary
	input := "https://gitlab.com\n" +
		"no\n" + // Token
		"no\n" + // Project
		"text\n" + // Format
		"no\n" + // Verbose
		"no\n" // Don't save

	cmd := exec.Command(mainBinary, "setup")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	// Verify config was NOT created
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Errorf("Config file was created despite rejection")
	}

	if !strings.Contains(string(output), "Configuration not saved") {
		t.Error("Expected 'not saved' message in output")
	}
}

func TestSetupCommand_ForceFlag(t *testing.T) {
	skipInCI(t)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create existing config
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	existingConfig := `gitlab:
  instance: https://old.example.com
`
	if err := os.WriteFile(configFile, []byte(existingConfig), 0600); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	// Input: No prompts needed with -f flag, just provide defaults
	input := "text\n" + // Format
		"no\n" + // Verbose
		"yes\n" + // Save
		"no\n" // Test

	cmd := exec.Command(mainBinary, "setup", "--force")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	// Should skip the modify/overwrite prompt
	if strings.Contains(string(output), "Configuration file already exists") {
		t.Error("Should skip modify/overwrite prompt with --force flag")
	}
}

func TestSetupCommand_HelpFlag(t *testing.T) {
	cmd := exec.Command(mainBinary, "setup", "--help")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup --help failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Interactive configuration wizard") {
		t.Error("Expected help text in output")
	}

	if !strings.Contains(outputStr, "Usage") {
		t.Error("Expected usage information in output")
	}
}

func TestSetupCommand_OutputContainsSummary(t *testing.T) {
	skipInCI(t)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")

	// Input: Complete setup
	input := "https://custom.gitlab.com\n" +
		"no\n" + // Token
		"yes\n" + // Project
		"my-group/my-project\n" + // Project path
		"yaml\n" + // Output format
		"yes\n" + // Verbose
		"yes\n" + // Save
		"no\n" // Test

	cmd := exec.Command(mainBinary, "setup")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify summary is shown
	if !strings.Contains(outputStr, "Configuration Summary") {
		t.Error("Expected configuration summary in output")
	}

	if !strings.Contains(outputStr, "custom.gitlab.com") {
		t.Error("Expected instance URL in summary")
	}

	if !strings.Contains(outputStr, "my-group/my-project") {
		t.Error("Expected project in summary")
	}

	if !strings.Contains(outputStr, "Output Format: yaml") {
		t.Error("Expected output format in summary")
	}

	if !strings.Contains(outputStr, "Verbose: true") {
		t.Error("Expected verbose setting in summary")
	}
}

func TestSetupCommand_CreatesDirectory(t *testing.T) {
	skipInCI(t)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "nested", ".gitlab-ci-lint")

	// Ensure parent directory doesn't exist
	_ = os.RemoveAll(filepath.Join(tempDir, "nested"))

	// Input: Minimal setup
	input := "https://gitlab.com\n" +
		"no\n" + // Token
		"no\n" + // Project
		"text\n" + // Format
		"no\n" + // Verbose
		"yes\n" + // Save
		"no\n" // Test

	cmd := exec.Command(mainBinary, "setup")
	cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Setup command failed: %v\nOutput: %s", err, output)
	}

	// Verify directory was created
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}
}

// BenchmarkSetupCommand benchmarks the setup command performance
func BenchmarkSetupCommand(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	tempDir := b.TempDir()
	configDir := filepath.Join(tempDir, ".gitlab-ci-lint")

	// Pre-build binary
	if err := buildBinaries(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create unique config dir for each iteration
		iterConfigDir := filepath.Join(configDir, "iter", string(rune(i)))

		input := "https://gitlab.com\n" +
			"no\n" + // Token
			"no\n" + // Project
			"text\n" + // Format
			"no\n" + // Verbose
			"yes\n" + // Save
			"no\n" // Test

		cmd := exec.Command(mainBinary, "setup")
		cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+iterConfigDir)
		cmd.Stdin = strings.NewReader(input)

		if err := cmd.Run(); err != nil {
			b.Fatalf("Setup failed: %v", err)
		}
	}
}

// TestSetupCommand_Concurrent tests concurrent setup operations
func TestSetupCommand_Concurrent(t *testing.T) {
	skipInCI(t)
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	concurrency := 5
	results := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(iteration int) {
			tempDir := t.TempDir()
			configDir := filepath.Join(tempDir, ".gitlab-ci-lint", string(rune(iteration)))

			input := "https://gitlab.com\n" +
				"no\n" + // Token
				"no\n" + // Project
				"text\n" + // Format
				"no\n" + // Verbose
				"yes\n" + // Save
				"no\n" // Test

			cmd := exec.Command(mainBinary, "setup")
			cmd.Env = append(os.Environ(), "GCL_CONFIG_DIR="+configDir)
			cmd.Stdin = strings.NewReader(input)

			results <- cmd.Run()
		}(i)
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent setup failed: %v", err)
		}
	}
}
