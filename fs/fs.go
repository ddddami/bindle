package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateDirIfNotExists creates a dir at a specified path if it does not exists
func CreateDirIfNotExists(path string) error {
	return CreateDirIfNotExistsWithPerm(path, 0o755)
}

// CreateDirIfNotExistsWithPerm creates a dir at a specified path and perm if it does not exists
func CreateDirIfNotExistsWithPerm(path string, perm os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, perm)
		if err != nil {
			return err
		}
	}
	return nil
}

// PathExists checks if a directory or a file exists at the given path
func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// DirExists Check if the path exists and is a directory
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// ListFilesWithExt returns a list of files in the given directory with the specified extension.
// The extension should include the dot, e.g. ".go"
func ListFilesWithExt(dir, ext string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ext {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

// SafeRemove removes a file or directory if it exists,
// and does nothing if it doesn't exist
func SafeRemove(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(path)
}

// CopyFile copies a file from src to dst.
// If dst already exists, it will be overwritten.
func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err = CreateDirIfNotExists(filepath.Dir(dst)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
