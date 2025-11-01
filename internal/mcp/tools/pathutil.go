package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathValidationResult contains the validated path and any warnings
type PathValidationResult struct {
	Path     string
	Warning  string
	UsedHome bool
}

// ValidateAndFixOutputPath validates an output path and fixes it if necessary.
// It tries the following in order:
// 1. If path is provided, expand tilde and validate writability
// 2. If path is empty or not writable, try current directory
// 3. If current directory is read-only, fall back to home directory
// 4. Returns error if no writable location found
func ValidateAndFixOutputPath(path, defaultFilename string) (*PathValidationResult, error) {
	result := &PathValidationResult{}

	// If path is provided, try to use it
	if path != "" {
		expanded := expandTilde(path)
		dir := filepath.Dir(expanded)

		// Check if directory exists and is writable
		if isDirectoryWritable(dir) {
			result.Path = expanded
			return result, nil
		}

		// Directory not writable, will try fallbacks
		result.Warning = fmt.Sprintf("Specified directory '%s' is not writable, trying fallback locations", dir)
	}

	// Try current directory
	cwd, err := os.Getwd()
	if err == nil {
		testPath := filepath.Join(cwd, defaultFilename)
		if isDirectoryWritable(cwd) {
			result.Path = testPath
			if result.Warning != "" {
				result.Warning += fmt.Sprintf("; using current directory: %s", cwd)
			}
			return result, nil
		}
	}

	// Current directory is read-only, try home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory and current directory is read-only: %w", err)
	}

	if !isDirectoryWritable(homeDir) {
		return nil, fmt.Errorf("home directory is not writable: %s", homeDir)
	}

	result.Path = filepath.Join(homeDir, defaultFilename)
	result.UsedHome = true
	if result.Warning == "" {
		result.Warning = fmt.Sprintf("Current directory is read-only, using home directory: %s", homeDir)
	} else {
		result.Warning += fmt.Sprintf("; using home directory: %s", homeDir)
	}

	return result, nil
}

// ValidateInputPath validates that an input file exists and is readable
func ValidateInputPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("input path cannot be empty")
	}

	expanded := expandTilde(path)

	// Check if file exists
	info, err := os.Stat(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file does not exist: %s", path)
		}
		return "", fmt.Errorf("cannot access file %s: %w", path, err)
	}

	// Check if it's a regular file (not a directory)
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", path)
	}

	return expanded, nil
}

// ValidateDirectoryPath validates that a directory exists and is accessible
// If createIfMissing is true, creates the directory if it doesn't exist
func ValidateDirectoryPath(path string, createIfMissing bool) (string, error) {
	if path == "" {
		return "", fmt.Errorf("directory path cannot be empty")
	}

	expanded := expandTilde(path)

	// Check if directory exists
	info, err := os.Stat(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			if createIfMissing {
				// Create directory with standard permissions
				if err := os.MkdirAll(expanded, 0755); err != nil {
					return "", fmt.Errorf("failed to create directory %s: %w", path, err)
				}
				return expanded, nil
			}
			return "", fmt.Errorf("directory does not exist: %s", path)
		}
		return "", fmt.Errorf("cannot access directory %s: %w", path, err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return "", fmt.Errorf("path is a file, not a directory: %s", path)
	}

	// Check if directory is writable
	if !isDirectoryWritable(expanded) {
		return "", fmt.Errorf("directory is not writable: %s", path)
	}

	return expanded, nil
}

// isDirectoryWritable checks if a directory is writable by attempting to create a temp file
func isDirectoryWritable(dir string) bool {
	// Create a test file
	testFile := filepath.Join(dir, ".gimage_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()

	// Clean up test file
	os.Remove(testFile)
	return true
}

// expandTilde expands ~ to the user's home directory
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path // Return original if we can't get home dir
	}

	if path == "~" {
		return homeDir
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}

	return path
}
