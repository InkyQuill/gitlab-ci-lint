package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultFilenames are the default CI configuration filenames to search for
var DefaultFilenames = []string{".gitlab-ci.yml", ".gitlab-ci.yaml"}

// DefaultIgnorePatterns are directories to skip during recursive search
var DefaultIgnorePatterns = []string{
	".git",
	"node_modules",
	"vendor",
	".venv",
	"__pycache__",
	"venv",
	"env",
	".tox",
	"target",
	"build",
	"dist",
	"out",
	".next",
	".nuxt",
	"coverage",
	".cache",
}

// Discoverer handles file discovery operations
type Discoverer struct {
	filenames      []string
	ignorePatterns []string
	maxDepth       int
}

// FileDiscovery contains discovered files and any errors encountered
type FileDiscovery struct {
	Files  []string
	Errors []error
}

// NewDiscoverer creates a new file discoverer with default settings
func NewDiscoverer() *Discoverer {
	return &Discoverer{
		filenames:      DefaultFilenames,
		ignorePatterns: DefaultIgnorePatterns,
		maxDepth:       10,
	}
}

// SetFilenames sets the filenames to search for
func (d *Discoverer) SetFilenames(filenames []string) {
	d.filenames = filenames
}

// SetIgnorePatterns sets the directory ignore patterns
func (d *Discoverer) SetIgnorePatterns(patterns []string) {
	d.ignorePatterns = patterns
}

// SetMaxDepth sets the maximum depth for parent directory search
func (d *Discoverer) SetMaxDepth(depth int) {
	d.maxDepth = depth
}

// FindInCurrentAndParents walks up the directory tree looking for CI files
// Starts from the current directory and searches up to maxDepth levels
func (d *Discoverer) FindInCurrentAndParents() (*FileDiscovery, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	return d.FindFromDirectory(dir)
}

// FindFromDirectory walks up from a given directory looking for CI files
func (d *Discoverer) FindFromDirectory(startDir string) (*FileDiscovery, error) {
	result := &FileDiscovery{
		Files: make([]string, 0),
	}

	currentDir := startDir
	depth := 0

	for depth <= d.maxDepth {
		// Check each filename in the current directory
		for _, filename := range d.filenames {
			path := filepath.Join(currentDir, filename)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				result.Files = append(result.Files, path)
				return result, nil
			}
		}

		// Move to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root
			break
		}
		currentDir = parentDir
		depth++
	}

	return nil, fmt.Errorf("no CI configuration file found in %s or parent directories (searched %d levels)", startDir, d.maxDepth)
}

// FindInDirectoryTree recursively searches a directory tree for all CI files.
//
// Depth Limiting:
//   - Respects the maxDepth setting set via SetMaxDepth() (default: 10 levels)
//   - Depth is calculated relative to the root directory provided
//   - For example, with maxDepth=2 and root="/home/user/project":
//     * /home/user/project/.gitlab-ci.yml (depth 0) ✓
//     * /home/user/project/subdir/.gitlab-ci.yml (depth 1) ✓
//     * /home/user/project/subdir/nested/.gitlab-ci.yml (depth 2) ✓
//     * /home/user/project/subdir/nested/deep/.gitlab-ci.yml (depth 3) ✗
//
// Ignore Patterns:
//   - Respects ignore patterns set via SetIgnorePatterns()
//   - Default ignores: .git, node_modules, vendor, build, dist, *.tar.gz
func (d *Discoverer) FindInDirectoryTree(root string) (*FileDiscovery, error) {
	result := &FileDiscovery{
		Files:  make([]string, 0),
		Errors: make([]error, 0),
	}

	// Check if root exists
	if info, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("directory not found: %s: %w", root, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", root)
	}

	// Clean root path for consistent depth calculation
	root = filepath.Clean(root)

	// Walk the directory tree
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			// Check if it's a permission error
			if os.IsPermission(err) {
				result.Errors = append(result.Errors, fmt.Errorf("permission denied: %s", path))
				return filepath.SkipDir
			}
			return err
		}

		// Check depth limit by counting path separators
		relPath, err := filepath.Rel(root, path)
		if err == nil {
			depth := 0
			if relPath != "." {
				depth = len(strings.Split(filepath.ToSlash(relPath), "/"))
			}
			// For directories, check if we've exceeded maxDepth
			if entry.IsDir() && depth > d.maxDepth {
				return filepath.SkipDir
			}
		}

		// Skip directories and ignored patterns
		if entry.IsDir() {
			if d.shouldIgnore(filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if filename matches any of our target filenames
		baseName := filepath.Base(path)
		for _, filename := range d.filenames {
			if baseName == filename {
				result.Files = append(result.Files, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}

	return result, nil
}

// ValidateAndExpandPaths validates files and expands directories recursively
// Returns a list of valid file paths and any errors encountered
func (d *Discoverer) ValidateAndExpandPaths(paths []string) (*FileDiscovery, error) {
	result := &FileDiscovery{
		Files:  make([]string, 0),
		Errors: make([]error, 0),
	}

	for _, path := range paths {
		// Get file info
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				result.Errors = append(result.Errors, fmt.Errorf("file not found: %s", path))
			} else {
				result.Errors = append(result.Errors, fmt.Errorf("error accessing %s: %w", path, err))
			}
			continue
		}

		// Handle directories
		if info.IsDir() {
			dirResult, err := d.FindInDirectoryTree(path)
			if err != nil {
				result.Errors = append(result.Errors, err)
				continue
			}
			result.Files = append(result.Files, dirResult.Files...)
			result.Errors = append(result.Errors, dirResult.Errors...)
			continue
		}

		// Regular file - add it directly
		result.Files = append(result.Files, path)
	}

	// Deduplicate files
	result.Files = d.deduplicate(result.Files)

	return result, nil
}

// shouldIgnore checks if a directory name should be ignored
func (d *Discoverer) shouldIgnore(name string) bool {
	for _, pattern := range d.ignorePatterns {
		if strings.EqualFold(name, pattern) {
			return true
		}
	}
	return false
}

// deduplicate removes duplicate file paths while preserving order
func (d *Discoverer) deduplicate(paths []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(paths))

	for _, path := range paths {
		// Normalize path for comparison
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}

		if !seen[absPath] {
			seen[absPath] = true
			result = append(result, path)
		}
	}

	return result
}
