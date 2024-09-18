package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
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

type memoryFileManager struct {
	fs fstest.MapFS
}

func NewMemoryFSFileManager(fs fstest.MapFS) FileManager {
	return memoryFileManager{fs: fs}
}

func (m memoryFileManager) WriteFile(ctx context.Context, path string, data []byte) error {
	m.fs[path] = &fstest.MapFile{Data: data}
	return nil
}

var trimSpaceMultilineRegexp = regexp.MustCompile(`(?m)(^\s+|\s+$)`)

// TestFileManager returns a file manager that knows how to assert written files in tests.
type TestFileManager struct {
	files map[string]string
}

func NewTestFileManager() TestFileManager {
	return TestFileManager{files: map[string]string{}}
}

func (f TestFileManager) WriteFile(ctx context.Context, path string, data []byte) error {
	f.files[path] = string(data)
	return nil
}

func (f TestFileManager) AssertContains(t *testing.T, path string, exp []string) {
	got, ok := f.files[path]
	if !ok {
		assert.Fail(t, "path missing", path)
		return
	}

	// Sanitize got HTML so we make easier to check content.
	got = trimSpaceMultilineRegexp.ReplaceAllString(got, "")
	got = strings.Replace(got, "\n", " ", -1)

	// Check each expected snippet.
	for _, e := range exp {
		assert.Contains(t, got, e)
	}
}
