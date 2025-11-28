package utils

import (
	"errors"
	"fmt"
	"os"
	"path"
)

// CreateFile ensures that a file and its parent directories exist.
// If the directory structure does not exist, it is created.
// If the file does not exist, it is created with write-only permissions.
//
// Behavior:
//   - Returns an error if the path is empty
//   - Creates parent directories recursively (0755 permissions)
//   - Creates the file if missing (write-only, 0755)
//   - Returns a wrapped error for any filesystem operation failure
//
// Parameters:
//
//	filePath - absolute or relative path to the target file
//
// Returns:
//
//	error - nil on success; an error if the file or directories cannot be created
func CreateFile(filePath string) error {
	if filePath == "" {
		return errors.New("no file storage path was provided")
	}

	if err := os.MkdirAll(path.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("error creating directory for file: %w", err)
	}

	if _, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0755); err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	return nil
}

// AppendToFile appends a single line of data to the specified file.
// The file is created if it does not already exist.
//
// Behavior:
//   - Opens the file in write/append mode (0755)
//   - Appends the provided byte slice as a string, followed by a newline
//   - Ensures file descriptor is properly closed, returning close errors if applicable
//
// Parameters:
//
//	filePath - path to the file where data should be appended
//	data     - byte slice to append as a line
//
// Returns:
//
//	error - nil on success; an error if the file cannot be opened, written to, or closed
func AppendToFile(filePath string, data []byte) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "%s\n", data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
