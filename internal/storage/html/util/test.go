package util

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var trimSpaceMultilineRegexp = regexp.MustCompile(`(?m)(^\s+|\s+$)`)

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
