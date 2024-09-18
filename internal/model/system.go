package model

import "fmt"

type System struct {
	ID          string
	Name        string
	Description string
}

func (s *System) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("id is required")
	}

	if s.Name == "" {
		s.Name = s.ID
	}

	return nil
}
