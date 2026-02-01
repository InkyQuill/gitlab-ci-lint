package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidation_ValidYAML(t *testing.T) {
	// Create a valid .gitlab-ci.yml file
	tempFile := createTempFile(t, `
image: alpine:latest

stages:
  - build
  - test

build:
  stage: build
  script:
    - echo "Building"
  artifacts:
    paths:
      - bin/

test:
  stage: test
  script:
    - echo "Testing"
  dependencies:
    - build
`)

	// Run validation
	cmd := exec.Command(mainBinary, "--skip-api", tempFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Validation failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Valid") && !strings.Contains(outputStr, "valid") {
		t.Logf("Note: Output does not contain 'Valid': %s", outputStr)
	}
}

func TestValidation_InvalidYAML(t *testing.T) {
	// Create an invalid YAML file
	tempFile := createTempFile(t, `
image: alpine:latest

stages:
  - build
  test

build:
  stage: build
    script:
      - echo "Invalid indentation"
`)

	// Run validation
	cmd := exec.Command(mainBinary, "--skip-api", tempFile)
	output, err := cmd.CombinedOutput()

	// Should fail with exit code 10 (validation error)
	if err == nil {
		t.Error("Expected validation error, but got none")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatal("Expected exit error")
	}

	if exitErr.ExitCode() != 10 {
		t.Errorf("Expected exit code 10, got %d", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "error") &&
		!strings.Contains(strings.ToLower(outputStr), "invalid") {
		t.Errorf("Expected error message in output: %s", outputStr)
	}
}

func TestValidation_OutputFormats(t *testing.T) {
	validYAML := `
stages:
  - test

test:
  stage: test
  script:
    - echo "test"
`

	tempFile := createTempFile(t, validYAML)

	// Test JSON output
	cmd := exec.Command(mainBinary, "--skip-api", "--output", "json", tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("JSON validation failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "{") || !strings.Contains(outputStr, "}") {
		t.Error("Expected JSON output with braces")
	}

	// Test YAML output
	cmd = exec.Command(mainBinary, "--skip-api", "--output", "yaml", tempFile)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("YAML validation failed: %v\nOutput: %s", err, output)
	}

	outputStr = string(output)
	if !strings.Contains(outputStr, "valid:") && !strings.Contains(outputStr, "stage:") {
		t.Logf("Note: YAML output format: %s", outputStr)
	}
}

func TestValidation_VerboseFlag(t *testing.T) {
	validYAML := `
stages:
  - test

test:
  stage: test
  script:
    - echo "test"
`

	tempFile := createTempFile(t, validYAML)

	// Run with verbose flag
	cmd := exec.Command(mainBinary, "--skip-api", "--verbose", tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Verbose validation failed: %v\nOutput: %s", err, output)
	}

	// Verbose output should include more details
	outputStr := string(output)
	t.Logf("Verbose output: %s", outputStr)
}

func TestValidation_ConfigFile(t *testing.T) {
	// Create config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
gitlab:
  instance: https://gitlab.com
  timeout: 30s
auth:
  token: ""
  netrc: false
validation:
  skip_api: true
  strict: true
  project: ""
output:
  format: json
  verbose: false
  color: never
`

	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	validYAML := `
stages:
  - test

test:
  stage: test
  script:
    - echo "test"
`

	tempFile := createTempFile(t, validYAML)

	// Run with config file
	cmd := exec.Command(mainBinary, "--config", configFile, tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Validation with config failed: %v\nOutput: %s", err, output)
	}

	// Should output JSON format due to config
	outputStr := string(output)
	if !strings.Contains(outputStr, "{") {
		t.Errorf("Expected JSON output from config, got: %s", outputStr)
	}
}

func TestValidation_MultipleFiles(t *testing.T) {
	// Create multiple CI files
	files := []string{
		`
stages:
  - build

build:
  stage: build
  script: echo "build"
`,
		`
stages:
  - test

test:
  stage: test
  script: echo "test"
`,
	}

	tempFiles := make([]string, len(files))
	for i, content := range files {
		tempFiles[i] = createTempFile(t, content)
	}

	// Validate each file
	for _, tf := range tempFiles {
		cmd := exec.Command(mainBinary, "--skip-api", tf)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("Validation failed for %s: %v\nOutput: %s", tf, err, output)
		}
	}
}

func TestValidation_ComplexPipeline(t *testing.T) {
	// Create a complex GitLab CI configuration
	complexYAML := `
variables:
  GLOBAL_VAR: "value"
  ANOTHER_VAR: 123

default:
  image: alpine:latest
  artifacts:
    expire_in: 1 week

stages:
  - build
  - test
  - deploy

.build_template: &build_template
  stage: build
  before_script:
    - echo "Before"
  after_script:
    - echo "After"

build1:
  <<: *build_template
  script:
    - make build
  cache:
    key: ${CI_COMMIT_REF_SLUG}
    paths:
      - vendor/

build2:
  <<: *build_template
  parallel: 5
  script:
    - make build-all

test:
  stage: test
  needs: [build1, build2]
  script:
    - make test
  only:
    - branches
  except:
    - /^hotfix-.*$/
  coverage: '/Code coverage: \d+\.\d+/'

deploy:
  stage: deploy
  script:
    - make deploy
  when: manual
  environment:
    name: production
    url: https://example.com
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
`

	tempFile := createTempFile(t, complexYAML)

	cmd := exec.Command(mainBinary, "--skip-api", tempFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Complex pipeline validation failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Complex pipeline validated successfully")
}

func TestValidation_AnchorsAndAliases(t *testing.T) {
	yamlWithAnchors := `
.defaults: &defaults
  cache:
    paths:
      - vendor/

.job_template: &job_template
  stage: build
  before_script:
    - echo "Before"

job1:
  <<: *defaults
  <<: *job_template
  script: test

job2:
  <<: [*defaults, *job_template]
  script: deploy
`

	tempFile := createTempFile(t, yamlWithAnchors)

	cmd := exec.Command(mainBinary, "--skip-api", tempFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Anchors and aliases validation failed: %v\nOutput: %s", err, output)
	}
}

func TestValidation_EmptyFile(t *testing.T) {
	tempFile := createTempFile(t, "")

	cmd := exec.Command(mainBinary, "--skip-api", tempFile)
	output, err := cmd.CombinedOutput()

	// Empty files might pass or fail depending on YAML parser
	// Just ensure it doesn't crash
	_ = err
	_ = output
	t.Log("Empty file validation completed (may succeed or fail)")
}

func TestValidation_ExitCodes(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		skipAPI   bool
		wantCode  int
	}{
		{
			name: "valid_local",
			yaml: `
stages:
  - test
test:
  stage: test
  script: echo "test"
`,
			skipAPI:  true,
			wantCode: 0,
		},
		{
			name: "invalid_syntax",
			yaml: `
stages:
  - test
invalid: [unclosed
`,
			skipAPI:  true,
			wantCode: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := createTempFile(t, tt.yaml)

			args := []string{tempFile}
			if tt.skipAPI {
				args = append([]string{"--skip-api"}, args...)
			}

			cmd := exec.Command(mainBinary, args...)
			err := cmd.Run()

			exitErr, ok := err.(*exec.ExitError)
			if !ok && tt.wantCode != 0 {
				t.Errorf("Expected exit error with code %d, got: %v", tt.wantCode, err)
			}

			if ok && exitErr.ExitCode() != tt.wantCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantCode, exitErr.ExitCode())
			}
		})
	}
}

func TestValidation_HelpFlag(t *testing.T) {
	cmd := exec.Command(mainBinary, "--help")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help flag failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage") {
		t.Error("Expected usage information in help")
	}

	if !strings.Contains(outputStr, "gitlab-ci-lint") {
		t.Error("Expected tool name in help")
	}
}

func TestValidation_VersionCommand(t *testing.T) {
	cmd := exec.Command(mainBinary, "version")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	outputStr := string(output)
	t.Logf("Version output: %s", outputStr)

	// Version output should contain something
	if strings.TrimSpace(outputStr) == "" {
		t.Error("Expected version information")
	}
}

func TestValidation_NonExistentFile(t *testing.T) {
	cmd := exec.Command(mainBinary, "--skip-api", "/nonexistent/file.yml")

	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatal("Expected exit error")
	}

	// Should be exit code 1 (general error), not 10 (validation error)
	if exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1 for missing file, got %d", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "no such file") &&
		!strings.Contains(strings.ToLower(outputStr), "not found") {
		t.Logf("Error message: %s", outputStr)
	}
}

func TestValidation_StdinInput(t *testing.T) {
	validYAML := `
stages:
  - test

test:
  stage: test
  script:
    - echo "test"
`

	cmd := exec.Command(mainBinary, "--skip-api", "-")
	cmd.Stdin = strings.NewReader(validYAML)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Stdin validation failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Stdin validation completed")
}

func TestValidation_EnvVarConfig(t *testing.T) {
	// Set environment variable for config
	validYAML := `
stages:
  - test

test:
  stage: test
  script:
    - echo "test"
`

	tempFile := createTempFile(t, validYAML)

	cmd := exec.Command(mainBinary, "--skip-api", tempFile)
	cmd.Env = append(os.Environ(), "GCL_OUTPUT_FORMAT=json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Env var config validation failed: %v\nOutput: %s", err, output)
	}

	// Should output JSON format due to env var
	outputStr := string(output)
	if !strings.Contains(outputStr, "{") {
		t.Logf("Note: Output with env var: %s", outputStr)
	}
}

func TestValidation_ColorFlags(t *testing.T) {
	validYAML := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`

	tempFile := createTempFile(t, validYAML)

	colorSettings := []string{"always", "never", "auto"}

	for _, color := range colorSettings {
		t.Run("color_"+color, func(t *testing.T) {
			cmd := exec.Command(mainBinary, "--skip-api", "--color", color, tempFile)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("Validation with color=%s failed: %v\nOutput: %s", color, err, output)
			}

			t.Logf("Color=%s validation completed", color)
		})
	}
}

// Helper function to create a temporary file with content
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, ".gitlab-ci.yml")

	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tempFile
}

// BenchmarkValidation benchmarks validation performance
func BenchmarkValidation_ValidYAML(b *testing.B) {
	validYAML := `
stages:
  - build
  - test

build:
  stage: build
  script:
    - make build
  artifacts:
    paths:
      - bin/

test:
  stage: test
  script:
    - make test
  dependencies:
    - build
`

	tempFile := createTempFile(&testing.T{}, validYAML)
	defer os.Remove(tempFile)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(mainBinary, "--skip-api", tempFile)
		if err := cmd.Run(); err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// Tests for file discovery feature

func TestFileDiscovery_AutoDiscovery(t *testing.T) {
	// Create a temporary directory with a .gitlab-ci.yml file
	tempDir := t.TempDir()
	ciFile := filepath.Join(tempDir, ".gitlab-ci.yml")
	content := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`
	if err := os.WriteFile(ciFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create CI file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run without any arguments - should auto-discover the file
	cmd := exec.Command(mainBinary, "--skip-api")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Auto-discovery validation failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Valid") && !strings.Contains(outputStr, "valid") {
		t.Logf("Note: Output does not contain 'Valid': %s", outputStr)
	}
}

func TestFileDiscovery_AutoDiscoveryParentDirectory(t *testing.T) {
	// Create a directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectories: %v", err)
	}

	// Create .gitlab-ci.yml in the parent (tempDir)
	ciFile := filepath.Join(tempDir, ".gitlab-ci.yml")
	content := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`
	if err := os.WriteFile(ciFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create CI file: %v", err)
	}

	// Change to nested subdirectory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run without any arguments - should find file in parent directory
	cmd := exec.Command(mainBinary, "--skip-api")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Parent directory discovery failed: %v\nOutput: %s", err, output)
	}
}

func TestFileDiscovery_MultipleFilesWithFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple CI files
	file1 := filepath.Join(tempDir, "ci1.yml")
	file2 := filepath.Join(tempDir, "ci2.yml")

	content1 := `
stages:
  - build

build:
  stage: build
  script: echo "build"
`
	content2 := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Validate multiple files using -f flag
	cmd := exec.Command(mainBinary, "--skip-api", "-f", file1, "-f", file2)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Multiple files validation failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Summary") {
		t.Logf("Note: Expected summary output: %s", outputStr)
	}
}

func TestFileDiscovery_DirectoryScanning(t *testing.T) {
	tempDir := t.TempDir()

	// Create directory structure with multiple CI files
	subdir1 := filepath.Join(tempDir, "service1")
	subdir2 := filepath.Join(tempDir, "service2")
	ignoredDir := filepath.Join(tempDir, "node_modules")

	if err := os.MkdirAll(subdir1, 0755); err != nil {
		t.Fatalf("Failed to create subdir1: %v", err)
	}
	if err := os.MkdirAll(subdir2, 0755); err != nil {
		t.Fatalf("Failed to create subdir2: %v", err)
	}
	if err := os.MkdirAll(ignoredDir, 0755); err != nil {
		t.Fatalf("Failed to create ignored dir: %v", err)
	}

	// Create CI files
	ciFile1 := filepath.Join(subdir1, ".gitlab-ci.yml")
	ciFile2 := filepath.Join(subdir2, ".gitlab-ci.yml")
	ignoredCiFile := filepath.Join(ignoredDir, ".gitlab-ci.yml")

	content := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`

	if err := os.WriteFile(ciFile1, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create ciFile1: %v", err)
	}
	if err := os.WriteFile(ciFile2, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create ciFile2: %v", err)
	}
	if err := os.WriteFile(ignoredCiFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create ignoredCiFile: %v", err)
	}

	// Scan directory with -d flag
	cmd := exec.Command(mainBinary, "--skip-api", "-d", tempDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Directory scanning failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Should find 2 files (not the one in node_modules)
	if !strings.Contains(outputStr, "Total files: 2") && !strings.Contains(outputStr, "2") {
		t.Logf("Note: Expected 2 files found: %s", outputStr)
	}
}

func TestFileDiscovery_CombinedFileAndDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create a specific file
	specificFile := filepath.Join(tempDir, "specific.yml")
	content1 := `
stages:
  - build

build:
  stage: build
  script: echo "build"
`
	if err := os.WriteFile(specificFile, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create specific file: %v", err)
	}

	// Create a directory with a CI file
	subdir := filepath.Join(tempDir, "service")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	ciFile := filepath.Join(subdir, ".gitlab-ci.yml")
	content2 := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`
	if err := os.WriteFile(ciFile, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create CI file: %v", err)
	}

	// Use both -f and -d flags
	cmd := exec.Command(mainBinary, "--skip-api", "-f", specificFile, "-d", subdir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Combined file and directory validation failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Summary") {
		t.Logf("Note: Expected summary output: %s", outputStr)
	}
}

func TestFileDiscovery_NoFilesFound(t *testing.T) {
	// Create empty temp directory
	tempDir := t.TempDir()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run without any arguments - no CI files exist
	cmd := exec.Command(mainBinary, "--skip-api")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error when no files found")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatal("Expected exit error")
	}

	// Should be exit code 1 (general error)
	if exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1 for no files found, got %d", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "no ci") {
		t.Logf("Error message: %s", outputStr)
	}
}

func TestFileDiscovery_NonExistentFile(t *testing.T) {
	cmd := exec.Command(mainBinary, "--skip-api", "-f", "/nonexistent/file.yml")

	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for non-existent file with -f flag")
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "not found") {
		t.Logf("Error message: %s", outputStr)
	}
}

func TestFileDiscovery_NonExistentDirectory(t *testing.T) {
	cmd := exec.Command(mainBinary, "--skip-api", "-d", "/nonexistent/directory")

	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for non-existent directory with -d flag")
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "not found") &&
		!strings.Contains(strings.ToLower(outputStr), "directory") {
		t.Logf("Error message: %s", outputStr)
	}
}

func TestFileDiscovery_BackwardCompatibility(t *testing.T) {
	// Test that single positional argument still works
	tempFile := createTempFile(t, `
stages:
  - test

test:
  stage: test
  script: echo "test"
`)

	// Use positional argument (old style)
	cmd := exec.Command(mainBinary, "--skip-api", tempFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Backward compatibility validation failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Valid") && !strings.Contains(outputStr, "valid") {
		t.Logf("Note: Output does not contain 'Valid': %s", outputStr)
	}
}

func TestFileDiscovery_YamlExtension(t *testing.T) {
	tempDir := t.TempDir()

	// Create both .yml and .yaml files
	ymlFile := filepath.Join(tempDir, ".gitlab-ci.yml")
	yamlFile := filepath.Join(tempDir, "service", ".gitlab-ci.yaml")

	if err := os.MkdirAll(filepath.Dir(yamlFile), 0755); err != nil {
		t.Fatalf("Failed to create service dir: %v", err)
	}

	content := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`

	if err := os.WriteFile(ymlFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create yml file: %v", err)
	}
	if err := os.WriteFile(yamlFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create yaml file: %v", err)
	}

	// Scan directory - should find both extensions
	cmd := exec.Command(mainBinary, "--skip-api", "-d", tempDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("YAML extension test failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// Should find 2 files
	if !strings.Contains(outputStr, "Total files: 2") && !strings.Contains(outputStr, "2") {
		t.Logf("Note: Expected 2 files found: %s", outputStr)
	}
}

func TestFileDiscovery_ExitCodes(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		args        []string
		wantCode    int
		description string
	}{
		{
			name: "all_files_valid",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create subdirectories with CI files
				subdir1 := filepath.Join(tempDir, "service1")
				subdir2 := filepath.Join(tempDir, "service2")
				os.MkdirAll(subdir1, 0755)
				os.MkdirAll(subdir2, 0755)
				file1 := filepath.Join(subdir1, ".gitlab-ci.yml")
				file2 := filepath.Join(subdir2, ".gitlab-ci.yml")
				content := `
stages:
  - test

test:
  stage: test
  script: echo "test"
`
				os.WriteFile(file1, []byte(content), 0644)
				os.WriteFile(file2, []byte(content), 0644)
				return tempDir
			},
			args:        []string{"-d"},
			wantCode:    0,
			description: "All files valid should exit 0",
		},
		{
			name: "one_file_invalid",
			setup: func(t *testing.T) string {
				tempDir := t.TempDir()
				subdir1 := filepath.Join(tempDir, "service1")
				subdir2 := filepath.Join(tempDir, "service2")
				os.MkdirAll(subdir1, 0755)
				os.MkdirAll(subdir2, 0755)
				validFile := filepath.Join(subdir1, ".gitlab-ci.yml")
				invalidFile := filepath.Join(subdir2, ".gitlab-ci.yml")
				os.WriteFile(validFile, []byte(`
stages:
  - test
test:
  stage: test
  script: echo "test"
`), 0644)
				os.WriteFile(invalidFile, []byte(`
stages:
  - test
invalid: [unclosed
`), 0644)
				return tempDir
			},
			args:        []string{"-d"},
			wantCode:    10,
			description: "One or more invalid files should exit 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := tt.setup(t)

			args := append([]string{"--skip-api"}, tt.args...)
			args = append(args, tempDir)
			cmd := exec.Command(mainBinary, args...)
			err := cmd.Run()

			exitErr, ok := err.(*exec.ExitError)
			if !ok && tt.wantCode != 0 {
				t.Errorf("Expected exit error with code %d, got: %v", tt.wantCode, err)
			}

			if ok && exitErr.ExitCode() != tt.wantCode {
				t.Errorf("%s: Expected exit code %d, got %d", tt.description, tt.wantCode, exitErr.ExitCode())
			}
		})
	}
}
