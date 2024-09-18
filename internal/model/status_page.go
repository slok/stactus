package model

import "fmt"

type StatusPageSettings struct {
	Name string // E.g: GitHub.
	URL  string // E.g: https://statusgithub.com/.
}

func (s *StatusPageSettings) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}

	return nil
}
