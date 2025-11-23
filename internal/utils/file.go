package utils

import (
	"errors"
	"fmt"
	"os"
	"path"
)

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
