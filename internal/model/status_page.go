package model

import (
	"fmt"
	"strings"
)

type StatusPageSettings struct {
	Name  string // E.g: GitHub.
	URL   string // E.g: https://statusgithub.com/.
	Theme Theme
}

func (s *StatusPageSettings) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}

	s.URL = strings.TrimSpace(s.URL)
	s.URL = strings.TrimSuffix(s.URL, "/")

	if s.Theme.Simple == nil {
		return fmt.Errorf("at least one theme must be selected")
	}

	return nil
}

type Theme struct {
	// Can override the templates of any theme.
	OverrideTPLPath string

	// Themes settings, the one that is not null, it's the one being used.
	Simple *ThemeSimple
}

type ThemeSimple struct{}
