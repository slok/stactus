package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type FileManager interface {
	WriteFile(ctx context.Context, path string, data []byte) error
}

var StdFileManager = stdFileManager(false)

type stdFileManager bool

func (stdFileManager) WriteFile(ctx context.Context, path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	err = os.WriteFile(path, data, 0666)
	if err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}
