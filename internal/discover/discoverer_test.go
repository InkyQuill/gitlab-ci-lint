package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDiscoverer(t *testing.T) {
	d := NewDiscoverer()

	if d == nil {
		t.Fatal("NewDiscoverer returned nil")
	}

	if len(d.filenames) == 0 {
		t.Error("filenames not initialized")
	}

	if len(d.ignorePatterns) == 0 {
		t.Error("ignorePatterns not initialized")
	}

	if d.maxDepth == 0 {
		t.Error("maxDepth not initialized")
	}
}

func TestSetFilenames(t *testing.T) {
	d := NewDiscoverer()
	customFilenames := []string{"custom.yml", "other.yaml"}

	d.SetFilenames(customFilenames)

	if len(d.filenames) != 2 {
		t.Errorf("expected 2 filenames, got %d", len(d.filenames))
	}
}

func TestSetIgnorePatterns(t *testing.T) {
	d := NewDiscoverer()
	customPatterns := []string{"custom1", "custom2"}

	d.SetIgnorePatterns(customPatterns)

	if len(d.ignorePatterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(d.ignorePatterns))
	}
}

func TestSetMaxDepth(t *testing.T) {
	d := NewDiscoverer()
	d.SetMaxDepth(5)

	if d.maxDepth != 5 {
		t.Errorf("expected maxDepth 5, got %d", d.maxDepth)
	}
}

func TestFindInCurrentAndParents(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create .gitlab-ci.yml in temp directory
	ciFile := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(ciFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file: %v", err)
	}

	// Change to subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change to subdirectory: %v", err)
	}

	// Test finding the file
	d := NewDiscoverer()
	result, err := d.FindInCurrentAndParents()

	if err != nil {
		t.Fatalf("FindInCurrentAndParents failed: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	if result.Files[0] != ciFile {
		t.Errorf("expected file %s, got %s", ciFile, result.Files[0])
	}
}

func TestFindInCurrentAndParents_NotFound(t *testing.T) {
	// Create a temporary directory without CI files
	tmpDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	d := NewDiscoverer()
	d.SetMaxDepth(2) // Reduce depth for faster test

	_, err = d.FindInCurrentAndParents()

	if err == nil {
		t.Error("expected error when file not found, got nil")
	}
}

func TestFindFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create nested directories: %v", err)
	}

	// Create CI file in temp directory
	ciFile := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(ciFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file: %v", err)
	}

	d := NewDiscoverer()
	result, err := d.FindFromDirectory(subDir)

	if err != nil {
		t.Fatalf("FindFromDirectory failed: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	if result.Files[0] != ciFile {
		t.Errorf("expected file %s, got %s", ciFile, result.Files[0])
	}
}

func TestFindInDirectoryTree(t *testing.T) {
	// Create test directory structure
	tmpDir := t.TempDir()

	// Create files
	ciFile1 := filepath.Join(tmpDir, ".gitlab-ci.yml")
	ciFile2 := filepath.Join(tmpDir, "subdir", ".gitlab-ci.yml")
	ciFile3 := filepath.Join(tmpDir, "deep", "nested", ".gitlab-ci.yml")

	if err := os.WriteFile(ciFile1, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file 1: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(ciFile2), 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(ciFile2, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file 2: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(ciFile3), 0755); err != nil {
		t.Fatalf("failed to create deep dir: %v", err)
	}
	if err := os.WriteFile(ciFile3, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file 3: %v", err)
	}

	d := NewDiscoverer()
	result, err := d.FindInDirectoryTree(tmpDir)

	if err != nil {
		t.Fatalf("FindInDirectoryTree failed: %v", err)
	}

	if len(result.Files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(result.Files))
	}
}

func TestFindInDirectoryTree_IgnoresCommonDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create CI files in various locations
	ciFile1 := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if err := os.WriteFile(ciFile1, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file 1: %v", err)
	}

	// Create ignored directories with CI files
	for _, dirName := range DefaultIgnorePatterns {
		ignoredDir := filepath.Join(tmpDir, dirName)
		if err := os.Mkdir(ignoredDir, 0755); err != nil {
			t.Fatalf("failed to create ignored dir: %v", err)
		}
		ignoredFile := filepath.Join(ignoredDir, ".gitlab-ci.yml")
		if err := os.WriteFile(ignoredFile, []byte("test: true"), 0644); err != nil {
			t.Fatalf("failed to create CI file in ignored dir: %v", err)
		}
	}

	d := NewDiscoverer()
	result, err := d.FindInDirectoryTree(tmpDir)

	if err != nil {
		t.Fatalf("FindInDirectoryTree failed: %v", err)
	}

	// Should only find the root CI file
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file (others should be ignored), got %d", len(result.Files))
	}

	if result.Files[0] != ciFile1 {
		t.Errorf("expected file %s, got %s", ciFile1, result.Files[0])
	}
}

func TestFindInDirectoryTree_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file instead of directory
	filePath := filepath.Join(tmpDir, "notadir.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	d := NewDiscoverer()
	_, err := d.FindInDirectoryTree(filePath)

	if err == nil {
		t.Error("expected error for non-directory, got nil")
	}
}

func TestFindInDirectoryTree_DirectoryNotFound(t *testing.T) {
	d := NewDiscoverer()
	_, err := d.FindInDirectoryTree("/nonexistent/directory/that/does/not/exist")

	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}

func TestValidateAndExpandPaths_Files(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.yml")
	file2 := filepath.Join(tmpDir, "file2.yml")

	if err := os.WriteFile(file1, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	d := NewDiscoverer()
	result, err := d.ValidateAndExpandPaths([]string{file1, file2})

	if err != nil {
		t.Fatalf("ValidateAndExpandPaths failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}

	if len(result.Errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidateAndExpandPaths_Directories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	ciFile := filepath.Join(subDir, ".gitlab-ci.yml")
	if err := os.WriteFile(ciFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file: %v", err)
	}

	d := NewDiscoverer()
	result, err := d.ValidateAndExpandPaths([]string{tmpDir})

	if err != nil {
		t.Fatalf("ValidateAndExpandPaths failed: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	if result.Files[0] != ciFile {
		t.Errorf("expected file %s, got %s", ciFile, result.Files[0])
	}
}

func TestValidateAndExpandPaths_Mixed(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	file1 := filepath.Join(tmpDir, "file1.yml")
	if err := os.WriteFile(file1, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}

	// Create a directory with a CI file
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	ciFile := filepath.Join(subDir, ".gitlab-ci.yml")
	if err := os.WriteFile(ciFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create CI file: %v", err)
	}

	d := NewDiscoverer()
	result, err := d.ValidateAndExpandPaths([]string{file1, subDir})

	if err != nil {
		t.Fatalf("ValidateAndExpandPaths failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
}

func TestValidateAndExpandPaths_NonExistent(t *testing.T) {
	d := NewDiscoverer()
	result, err := d.ValidateAndExpandPaths([]string{"/nonexistent/file.yml"})

	if err != nil {
		t.Fatalf("ValidateAndExpandPaths failed: %v", err)
	}

	if len(result.Files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(result.Files))
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestDeduplicate(t *testing.T) {
	d := NewDiscoverer()

	paths := []string{
		"/path/to/file.yml",
		"/path/to/file.yml",
		"/other/file.yml",
		"/path/to/../to/file.yml", // Same as first after normalization
	}

	result := d.deduplicate(paths)

	// After deduplication, should have 2 unique files
	if len(result) != 2 {
		t.Errorf("expected 2 unique files, got %d", len(result))
	}
}

func TestShouldIgnore(t *testing.T) {
	d := NewDiscoverer()

	testCases := []struct {
		name     string
		expected bool
	}{
		{".git", true},
		{"node_modules", true},
		{"vendor", true},
		{".venv", true},
		{"__pycache__", true},
		{"target", true},
		{"build", true},
		{"dist", true},
		{"src", false},
		{"test", false},
		{"docs", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := d.shouldIgnore(tc.name)
			if result != tc.expected {
				t.Errorf("shouldIgnore(%s) = %v, expected %v", tc.name, result, tc.expected)
			}
		})
	}
}

func TestFindInDirectoryTree_YamlExtension(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both .yml and .yaml files
	ymlFile := filepath.Join(tmpDir, ".gitlab-ci.yml")
	yamlFile := filepath.Join(tmpDir, ".gitlab-ci.yaml")

	if err := os.WriteFile(ymlFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create yml file: %v", err)
	}
	if err := os.WriteFile(yamlFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("failed to create yaml file: %v", err)
	}

	d := NewDiscoverer()
	result, err := d.FindInDirectoryTree(tmpDir)

	if err != nil {
		t.Fatalf("FindInDirectoryTree failed: %v", err)
	}

	// Should find both .yml and .yaml files
	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
}
