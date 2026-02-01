package gitlab

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectProjectFromGitRepo attempts to detect the GitLab project path from a git repository
// It reads .git/config and extracts the project path from remote.origin.url
// Returns empty string if not a git repo or cannot determine the project
func DetectProjectFromGitRepo(dir string) (string, error) {
	// For git submodules, .git is a file pointing to the real git dir
	// We need to handle both cases
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		// .git doesn't exist
		return "", nil
	}

	var gitDir string
	if info.IsDir() {
		// Regular git repo
		gitDir = gitPath
	} else {
		// .git is a file - it's a submodule or worktree
		// Read the file to find the real git directory
		content, err := os.ReadFile(gitPath)
		if err != nil {
			return "", fmt.Errorf("failed to read .git file: %w", err)
		}

		// Parse gitdir: /path/to/parent/.git/modules/submodule
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "gitdir: ") {
				gitDirPath := strings.TrimSpace(strings.TrimPrefix(line, "gitdir: "))
				if filepath.IsAbs(gitDirPath) {
					gitDir = gitDirPath
				} else {
					// Relative path - resolve from current directory
					gitDir = filepath.Join(dir, gitDirPath)
				}
				break
			}
		}

		if gitDir == "" {
			return "", fmt.Errorf("invalid .git file format")
		}
	}

	// Try to read .git/config from the git directory
	gitConfigPath := filepath.Join(gitDir, "config")
	if _, err := os.Stat(gitConfigPath); os.IsNotExist(err) {
		// No config file found
		return "", nil
	}

	// Read .git/config file
	file, err := os.Open(gitConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read git config: %w", err)
	}
	defer file.Close()

	// Parse git config to find remote.origin.url
	scanner := bufio.NewScanner(file)
	inRemoteOrigin := false
	var remoteURL string

	// Regex to match remote "origin" section
	remoteOriginRegex := regexp.MustCompile(`^\s*\[remote\s+"origin"\]`)
	// Regex to match url = line
	urlRegex := regexp.MustCompile(`^\s*url\s*=\s*(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're in a [remote "origin"] section
		if remoteOriginRegex.MatchString(line) {
			inRemoteOrigin = true
			continue
		}

		// If we hit another section, we're done with remote origin
		if strings.HasPrefix(line, "[") && inRemoteOrigin {
			break
		}

		// Look for url = line
		if inRemoteOrigin {
			matches := urlRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				remoteURL = strings.TrimSpace(matches[1])
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read git config: %w", err)
	}

	// Extract project path from URL
	projectPath, err := extractProjectPath(remoteURL)
	if err != nil {
		return "", err
	}

	return projectPath, nil
}

// extractProjectPath extracts the project path from various GitLab URL formats
// Supported formats:
//   - https://gitlab.com/group/project.git
//   - git@gitlab.com:group/project.git
//   - http://gitlab.com/group/project
//   - git@gitlab.com:group/project
//   - ssh://git@gitlab.com/group/project.git
//   - git@gitlab.com:port:group/project.git
func extractProjectPath(remoteURL string) (string, error) {
	if remoteURL == "" {
		return "", nil
	}

	var projectPath string

	// Check for SSH format: git@host:path OR ssh://git@host/path
	// This must be checked before HTTPS format
	if strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "ssh://git@") {
		// Remove ssh:// prefix if present
		url := strings.TrimPrefix(remoteURL, "ssh://")

		// Remove git@ user part
		// Find the @ and take everything after it
		atIdx := strings.Index(url, "@")
		if atIdx != -1 {
			url = url[atIdx+1:]
		}

		// Check if it uses : (normal SSH) or / (ssh:// format)
		if strings.Contains(url, ":") && !strings.HasPrefix(url, "/") {
			// git@host:port:path format - split by : and skip host and port
			parts := strings.Split(url, ":")
			if len(parts) >= 3 {
				// Join everything after the port (parts[2:])
				projectPath = strings.Join(parts[2:], ":")
			} else {
				// git@host:path format - take second part
				projectPath = parts[1]
			}
		} else {
			// ssh://git@host/path format - extract path after host
			// Split by / and join everything after the host
			parts := strings.Split(url, "/")
			if len(parts) > 1 {
				projectPath = strings.Join(parts[1:], "/")
			}
		}
	} else {
		// HTTPS/HTTP format: https://host/path or http://host:port/path
		// Remove protocol
		re := regexp.MustCompile(`^(https?://)([^/]+)`)
		match := re.FindStringSubmatch(remoteURL)
		if len(match) > 2 {
			// Get everything after the host
			idx := len(match[1]) + len(match[2])
			projectPath = remoteURL[idx:]
			// Trim leading slash
			projectPath = strings.TrimPrefix(projectPath, "/")
		} else {
			projectPath = remoteURL
		}
	}

	// Remove .git suffix if present
	projectPath = strings.TrimSuffix(projectPath, ".git")

	// Trim any remaining leading/trailing slashes
	projectPath = strings.Trim(projectPath, "/")

	// Validate that it looks like a project path (contains /)
	if !strings.Contains(projectPath, "/") {
		return "", fmt.Errorf("invalid project path format: %s", remoteURL)
	}

	return projectPath, nil
}

// DetectProjectForFile attempts to detect the GitLab project for a specific file
// It searches upward from the file's directory to find the git repository root
// Handles both regular git repos and git submodules
func DetectProjectForFile(filePath string) (string, error) {
	// Get the directory containing the file
	dir := filepath.Dir(filePath)

	// First, try to detect from current directory (handles submodules)
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Stat(gitPath)

	if err == nil {
		// .git exists
		if info.IsDir() {
			// Regular git repository
			return DetectProjectFromGitRepo(dir)
		}

		// .git is a file - this is a git submodule or worktree
		// Try to detect project from the submodule's own config
		project, err := DetectProjectFromGitRepo(dir)
		if err == nil && project != "" {
			// Submodule has its own origin, use it
			return project, nil
		}

		// Submodule doesn't have its own origin or detection failed
		// Fall through to search upward for parent repo
	}

	// Not a git repo or submodule without own origin
	// Search upward for parent git repository
	for {
		// Go up one directory
		parentDir := filepath.Dir(dir)

		// Check if we've reached the root
		if parentDir == dir {
			// Reached filesystem root, no git repo found
			return "", nil
		}

		dir = parentDir

		// Check if parent has .git directory
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			// Found parent git repository
			return DetectProjectFromGitRepo(dir)
		}
	}
}
