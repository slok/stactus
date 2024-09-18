package model

import (
	"fmt"
	"strings"
)

type StatusPageSettings struct {
	Name string // E.g: GitHub.
	URL  string // E.g: https://statusgithub.com/.
}

func (s *StatusPageSettings) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}

	s.URL = strings.TrimSpace(s.URL)
	s.URL = strings.TrimSuffix(s.URL, "/")

	return nil
}
