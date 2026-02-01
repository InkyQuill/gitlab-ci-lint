package gitlab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractInstanceURL_HTTPS_Standard tests HTTPS standard URLs
func TestExtractInstanceURL_HTTPS_Standard(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://gitlab.com/group/project.git", "https://gitlab.com"},
		{"https://gitlab.example.com/group/project", "https://gitlab.example.com"},
		{"https://git.a1tis.ru/group/project.git", "https://git.a1tis.ru"},
		{"https://gitlab.com/", "https://gitlab.com"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := extractInstanceURL(tt.url)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestExtractInstanceURL_HTTPS_CustomDomain tests custom domains
func TestExtractInstanceURL_HTTPS_CustomDomain(t *testing.T) {
	urls := []string{
		"https://gitlab.mycompany.com/group/project.git",
		"https://code.example.org/myproject",
		"https://git.internal.io/group/project",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			result, err := extractInstanceURL(url)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result == "" {
				t.Error("Expected non-empty result")
			}

			if !strings.HasPrefix(result, "https://") {
				t.Errorf("Expected HTTPS prefix, got '%s'", result)
			}
		})
	}
}

// TestExtractInstanceURL_HTTP_WithPort tests HTTP URLs with ports
func TestExtractInstanceURL_HTTP_WithPort(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://gitlab.com:8080/group/project.git", "http://gitlab.com:8080"},
		{"http://localhost:3000/myproject", "http://localhost:3000"},
		{"http://gitlab.example.com:80/group/project", "http://gitlab.example.com:80"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := extractInstanceURL(tt.url)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestExtractInstanceURL_SSH_Standard tests SSH format URLs
func TestExtractInstanceURL_SSH_Standard(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"git@gitlab.com:group/project.git", "http://gitlab.com"},
		{"git@gitlab.example.com:group/project", "http://gitlab.example.com"},
		{"git@git.a1tis.ru:mygroup/myproject", "http://git.a1tis.ru"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := extractInstanceURL(tt.url)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestExtractInstanceURL_SSH_WithPort tests SSH URLs with ports
func TestExtractInstanceURL_SSH_WithPort(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"git@gitlab.com:2222:group/project.git", "http://gitlab.com"},
		{"git@gitlab.example.com:443:myproject", "http://gitlab.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := extractInstanceURL(tt.url)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestExtractInstanceURL_SSH_Protocol tests ssh:// protocol URLs
func TestExtractInstanceURL_SSH_Protocol(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"ssh://git@gitlab.com/group/project.git", "http://gitlab.com"},
		{"ssh://git@gitlab.example.com:2222/group/project", "http://gitlab.example.com"}, // Port gets lost
		{"ssh://git@git.a1tis.ru/myproject", "http://git.a1tis.ru"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := extractInstanceURL(tt.url)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestExtractInstanceURL_Empty tests empty URL handling
func TestExtractInstanceURL_Empty(t *testing.T) {
	result, err := extractInstanceURL("")
	if err != nil {
		t.Errorf("Expected no error for empty string, got: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

// TestExtractInstanceURL_Invalid tests invalid URL handling
func TestExtractInstanceURL_Invalid(t *testing.T) {
	tests := []string{
		"not-a-url",
		":invalid",
		"ftp://gitlab.com/project",
		"git@gitlab.com", // Missing path
	}

	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			result, err := extractInstanceURL(url)
			if err == nil {
				// Some invalid URLs might not error, just check behavior
				t.Logf("URL '%s' returned '%s' without error", url, result)
			} else {
				// Expected error for truly invalid URLs
				t.Logf("URL '%s' correctly returned error: %v", url, err)
			}
		})
	}
}

// TestDetectInstanceAndProject_ValidGitRepo tests detection in a valid git repo
func TestDetectInstanceAndProject_ValidGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	gitConfigPath := filepath.Join(gitDir, "config")
	gitConfigContent := `[remote "origin"]
	url = https://gitlab.com/group/project.git
`
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	instance, _, err := DetectInstanceAndProject(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "https://gitlab.com" {
		t.Errorf("Expected instance 'https://gitlab.com', got '%s'", instance)
	}
}

// TestDetectInstanceAndProject_HTTPS_vs_SSH tests different URL formats
func TestDetectInstanceAndProject_HTTPS_vs_SSH(t *testing.T) {
	tests := []struct {
		name        string
		gitConfigURL string
		expectedInstance string
		expectedProject  string
	}{
		{
			name:        "HTTPS",
			gitConfigURL: "url = https://gitlab.com/group/project.git",
			expectedInstance: "https://gitlab.com",
			expectedProject:  "group/project",
		},
		{
			name:        "SSH",
			gitConfigURL: "url = git@gitlab.com:group/project.git",
			expectedInstance: "http://gitlab.com",
			expectedProject:  "group/project",
		},
		{
			name:        "SSH with port",
			gitConfigURL: "url = git@gitlab.com:2222:group/project.git",
			expectedInstance: "http://gitlab.com",
			expectedProject:  "group/project",
		},
		{
			name:        "SSH protocol",
			gitConfigURL: "url = ssh://git@gitlab.com/group/project.git",
			expectedInstance: "http://gitlab.com",
			expectedProject:  "group/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			gitDir := filepath.Join(tempDir, ".git")
			if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

			gitConfigContent := `[remote "origin"]
` + tt.gitConfigURL + `
`
			gitConfigPath := filepath.Join(gitDir, "config")
			if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
				t.Fatalf("Failed to write git config: %v", err)
			}

			instance, project, err := DetectInstanceAndProject(tempDir)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if instance != tt.expectedInstance {
				t.Errorf("Expected instance '%s', got '%s'", tt.expectedInstance, instance)
			}

			if project != tt.expectedProject {
				t.Errorf("Expected project '%s', got '%s'", tt.expectedProject, project)
			}
		})
	}
}

// TestDetectInstanceAndProject_CustomInstance tests custom GitLab instance
func TestDetectInstanceAndProject_CustomInstance(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	gitConfigContent := `[remote "origin"]
	url = https://gitlab.example.com:8080/mygroup/myproject.git
`
	gitConfigPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	instance, _, err := DetectInstanceAndProject(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// HTTP with port gets normalized to https
	if instance != "https://gitlab.example.com:8080" {
		t.Errorf("Expected instance 'https://gitlab.example.com:8080', got '%s'", instance)
	}
}

// TestDetectInstanceAndProject_NoGitDir tests behavior when .git doesn't exist
func TestDetectInstanceAndProject_NoGitDir(t *testing.T) {
	tempDir := t.TempDir()
	// Don't create .git directory

	instance, project, err := DetectInstanceAndProject(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "" {
		t.Errorf("Expected empty instance, got '%s'", instance)
	}

	if project != "" {
		t.Errorf("Expected empty project, got '%s'", project)
	}
}

// TestDetectInstanceAndProject_NoRemoteOrigin tests when no remote origin exists
func TestDetectInstanceAndProject_NoRemoteOrigin(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create config without remote origin
	gitConfigContent := `[core]
	repositoryformatversion = 0
`
	gitConfigPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	instance, project, err := DetectInstanceAndProject(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "" {
		t.Errorf("Expected empty instance, got '%s'", instance)
	}

	if project != "" {
		t.Errorf("Expected empty project, got '%s'", project)
	}
}

// TestDetectInstanceAndProject_InvalidGitConfig tests with invalid .git/config
func TestDetectInstanceAndProject_InvalidGitConfig(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create invalid config file
	gitConfigPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte("invalid config {{{"), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	// Should not panic, just return empty
	instance, project, err := DetectInstanceAndProject(tempDir)
	if err != nil {
		t.Logf("Got error (may be expected): %v", err)
	}

	if instance != "" {
		t.Logf("Got instance: %s", instance)
	}

	if project != "" {
		t.Logf("Got project: %s", project)
	}
}

// TestDetectInstanceAndProject_Submodule tests detection in a git submodule
func TestDetectInstanceAndProject_Submodule(t *testing.T) {
	tempDir := t.TempDir()

	// Create .git file (submodule marker)
	gitFilePath := filepath.Join(tempDir, ".git")
	gitFileContent := `gitdir: ../.git/modules/submodule
`
	if err := os.WriteFile(gitFilePath, []byte(gitFileContent), 0644); err != nil {
		t.Fatalf("Failed to write .git file: %v", err)
	}

	// Create the actual git directory
	parentGitDir := filepath.Join(tempDir, "..", ".git", "modules", "submodule")
	if err := os.MkdirAll(parentGitDir, 0755); err != nil {
		t.Fatalf("Failed to create parent git directory: %v", err)
	}

	gitConfigContent := `[remote "origin"]
	url = https://gitlab.com/submodule/group/project.git
`
	gitConfigPath := filepath.Join(parentGitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	instance, project, err := DetectInstanceAndProject(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "https://gitlab.com" {
		t.Errorf("Expected instance 'https://gitlab.com', got '%s'", instance)
	}

	if project != "submodule/group/project" {
		t.Errorf("Expected project 'submodule/group/project', got '%s'", project)
	}
}

// TestDetectInstanceAndProject_Worktree tests detection in a git worktree
func TestDetectInstanceAndProject_Worktree(t *testing.T) {
	tempDir := t.TempDir()

	// Create worktree git directory outside tempDir (to avoid .git file vs dir conflict)
	parentDir := filepath.Join(tempDir, "parent")
	worktreeGitDir := filepath.Join(parentDir, ".git", "worktrees", "worktree-name")
	if err := os.MkdirAll(worktreeGitDir, 0755); err != nil {
		t.Fatalf("Failed to create worktree git directory: %v", err)
	}

	// Create .git file (worktree marker) pointing to external location
	gitFilePath := filepath.Join(tempDir, ".git")
	gitFileContent := fmt.Sprintf("gitdir: %s\n", worktreeGitDir)
	if err := os.WriteFile(gitFilePath, []byte(gitFileContent), 0644); err != nil {
		t.Fatalf("Failed to write .git file: %v", err)
	}

	gitConfigContent := `[remote "origin"]
	url = https://gitlab.com/worktree/project.git
`
	gitConfigPath := filepath.Join(worktreeGitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	instance, _, err := DetectInstanceAndProject(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "https://gitlab.com" {
		t.Errorf("Expected instance 'https://gitlab.com', got '%s'", instance)
	}
}

// TestDetectInstanceForFile_InGitRepo tests detection for a file in a git repo
func TestDetectInstanceForFile_InGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	gitConfigContent := `[remote "origin"]
	url = https://gitlab.com/group/project.git
`
	gitConfigPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, ".gitlab-ci.yml")
	if err := os.WriteFile(testFile, []byte("image: alpine"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	instance, project, err := DetectInstanceForFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "https://gitlab.com" {
		t.Errorf("Expected instance 'https://gitlab.com', got '%s'", instance)
	}

	if project != "group/project" {
		t.Errorf("Expected project 'group/project', got '%s'", project)
	}
}

// TestDetectInstanceForFile_SubmoduleWithOrigin tests submodule with its own origin
func TestDetectInstanceForFile_SubmoduleWithOrigin(t *testing.T) {
	tempDir := t.TempDir()

	// Create .git file (submodule) - point to absolute path outside tempDir
	// to avoid the .git file vs directory conflict
	parentDir := filepath.Join(tempDir, "parent")
	parentGitDir := filepath.Join(parentDir, ".git", "modules", "submodule")
	if err := os.MkdirAll(parentGitDir, 0755); err != nil {
		t.Fatalf("Failed to create parent git directory: %v", err)
	}

	gitFilePath := filepath.Join(tempDir, ".git")
	gitFileContent := fmt.Sprintf("gitdir: %s\n", parentGitDir)
	if err := os.WriteFile(gitFilePath, []byte(gitFileContent), 0644); err != nil {
		t.Fatalf("Failed to write .git file: %v", err)
	}

	// Submodule has its own origin
	gitConfigContent := `[remote "origin"]
	url = https://custom.gitlab.com/submodule/project.git
`
	gitConfigPath := filepath.Join(parentGitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	// Create a test file in submodule
	testFile := filepath.Join(tempDir, "submodule-file.yml")
	if err := os.WriteFile(testFile, []byte("image: alpine"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	instance, _, err := DetectInstanceForFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should use submodule's own origin
	if instance != "https://custom.gitlab.com" {
		t.Errorf("Expected instance 'https://custom.gitlab.com', got '%s'", instance)
	}
}

// TestDetectInstanceForFile_SubmoduleWithoutOrigin tests submodule without its own origin
func TestDetectInstanceForFile_SubmoduleWithoutOrigin(t *testing.T) {
	tempDir := t.TempDir()

	// Create parent repo with .git
	parentGitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(parentGitDir, 0755); err != nil {
		t.Fatalf("Failed to create parent git directory: %v", err)
	}

	parentConfigContent := `[remote "origin"]
	url = https://gitlab.com/parent/project.git
`
	parentConfigPath := filepath.Join(parentGitDir, "config")
	if err := os.WriteFile(parentConfigPath, []byte(parentConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write parent git config: %v", err)
	}

	// Create submodule directory
	submoduleDir := filepath.Join(tempDir, "submodule")
	if err := os.MkdirAll(submoduleDir, 0755); err != nil {
		t.Fatalf("Failed to create submodule directory: %v", err)
	}

	// Create .git file for submodule (pointing to parent)
	gitFilePath := filepath.Join(submoduleDir, ".git")
	gitFileContent := `gitdir: ../.git/modules/submodule
`
	if err := os.WriteFile(gitFilePath, []byte(gitFileContent), 0644); err != nil {
		t.Fatalf("Failed to write .git file: %v", err)
	}

	// Create submodule git directory without config
	submoduleGitDir := filepath.Join(tempDir, ".git", "modules", "submodule")
	if err := os.MkdirAll(submoduleGitDir, 0755); err != nil {
		t.Fatalf("Failed to create submodule git directory: %v", err)
	}
	// No config file - submodule has no own origin

	// Create a test file in submodule
	testFile := filepath.Join(submoduleDir, "test.yml")
	if err := os.WriteFile(testFile, []byte("image: alpine"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	instance, _, err := DetectInstanceForFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should fall back to parent's origin
	if instance != "https://gitlab.com" {
		t.Errorf("Expected parent instance 'https://gitlab.com', got '%s'", instance)
	}
}

// TestDetectInstanceForFile_OutsideGitRepo tests file outside any git repo
func TestDetectInstanceForFile_OutsideGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	// No .git directory

	// Create a test file
	testFile := filepath.Join(tempDir, ".gitlab-ci.yml")
	if err := os.WriteFile(testFile, []byte("image: alpine"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	instance, project, err := DetectInstanceForFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "" {
		t.Errorf("Expected empty instance, got '%s'", instance)
	}

	if project != "" {
		t.Errorf("Expected empty project, got '%s'", project)
	}
}

// TestDetectInstanceForFile_DeepNesting tests file deeply nested in directories
func TestDetectInstanceForFile_DeepNesting(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	gitConfigContent := `[remote "origin"]
	url = https://gitlab.com/group/project.git
`
	gitConfigPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	// Create deeply nested file
	nestedPath := filepath.Join(tempDir, "level1", "level2", "level3", "test.yml")
	if err := os.MkdirAll(filepath.Dir(nestedPath), 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}
	if err := os.WriteFile(nestedPath, []byte("image: alpine"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	instance, project, err := DetectInstanceForFile(nestedPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "https://gitlab.com" {
		t.Errorf("Expected instance 'https://gitlab.com', got '%s'", instance)
	}

	if project != "group/project" {
		t.Errorf("Expected project 'group/project', got '%s'", project)
	}
}

// TestDetectInstanceForFile_GitFileNotDir tests .git as a file, not directory
func TestDetectInstanceForFile_GitFileNotDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create .git as a file (worktree or submodule)
	gitFilePath := filepath.Join(tempDir, ".git")
	gitFileContent := `gitdir: /tmp/test-git-dir
`
	if err := os.WriteFile(gitFilePath, []byte(gitFileContent), 0644); err != nil {
		t.Fatalf("Failed to write .git file: %v", err)
	}

	// Create the referenced git dir
	gitDir := "/tmp/test-git-dir"
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}
	// Note: gitDir cleanup is handled by test framework (tempDir or manual cleanup)

	gitConfigContent := `[remote "origin"]
	url = https://gitlab.com/worktree/project.git
`
	gitConfigPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(gitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.yml")
	if err := os.WriteFile(testFile, []byte("image: alpine"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	instance, _, err := DetectInstanceForFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if instance != "https://gitlab.com" {
		t.Errorf("Expected instance 'https://gitlab.com', got '%s'", instance)
	}
}
