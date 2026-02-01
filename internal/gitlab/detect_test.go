package gitlab

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractProjectPath(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
		hasError bool
	}{
		{
			name:     "HTTPS URL with .git",
			url:      "https://gitlab.com/group/project.git",
			expected: "group/project",
			hasError: false,
		},
		{
			name:     "HTTPS URL without .git",
			url:      "https://gitlab.com/group/project",
			expected: "group/project",
			hasError: false,
		},
		{
			name:     "SSH URL with .git",
			url:      "git@gitlab.com:group/project.git",
			expected: "group/project",
			hasError: false,
		},
		{
			name:     "SSH URL without .git",
			url:      "git@gitlab.com:group/project",
			expected: "group/project",
			hasError: false,
		},
		{
			name:     "Custom GitLab instance",
			url:      "https://git.example.com/group/subgroup/project.git",
			expected: "group/subgroup/project",
			hasError: false,
		},
		{
			name:     "With port number",
			url:      "git@gitlab.com:2222:group/project.git",
			expected: "group/project",
			hasError: false,
		},
		{
			name:     "HTTP with port",
			url:      "http://gitlab.com:80/group/project.git",
			expected: "group/project",
			hasError: false,
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
			hasError: false,
		},
		{
			name:     "Invalid - no slash",
			url:      "https://gitlab.com/project",
			expected: "",
			hasError: true,
		},
		{
			name:     "SSH protocol",
			url:      "ssh://git@gitlab.com/group/project.git",
			expected: "group/project",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := extractProjectPath(tc.url)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestDetectProjectFromGitRepo_NotAGitRepo(t *testing.T) {
	// Create a temporary directory without .git
	tempDir := t.TempDir()

	project, err := DetectProjectFromGitRepo(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if project != "" {
		t.Errorf("Expected empty project for non-git directory, got '%s'", project)
	}
}

func TestDetectProjectFromGitRepo_WithGitConfig(t *testing.T) {
	// Create a temporary git repository structure
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")

	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create a .git/config file
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = https://gitlab.com/ai/aphrodite.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`

	configPath := filepath.Join(gitDir, "config")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	project, err := DetectProjectFromGitRepo(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := "ai/aphrodite"
	if project != expected {
		t.Errorf("Expected project '%s', got '%s'", expected, project)
	}
}

func TestDetectProjectFromGitRepo_SSHUrl(t *testing.T) {
	// Create a temporary git repository structure with SSH URL
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")

	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create a .git/config file with SSH URL
	configContent := `[remote "origin"]
	url = git@gitlab.com:ai/backbone.git
`

	configPath := filepath.Join(gitDir, "config")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	project, err := DetectProjectFromGitRepo(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := "ai/backbone"
	if project != expected {
		t.Errorf("Expected project '%s', got '%s'", expected, project)
	}
}

func TestDetectProjectFromGitRepo_NoRemoteOrigin(t *testing.T) {
	// Create a git config without remote "origin"
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")

	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	configContent := `[remote "upstream"]
	url = https://gitlab.com/someone/project.git
`

	configPath := filepath.Join(gitDir, "config")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	project, err := DetectProjectFromGitRepo(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if project != "" {
		t.Errorf("Expected empty project when no remote origin, got '%s'", project)
	}
}
