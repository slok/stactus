package model

import "fmt"

type System struct {
	ID          string
	Name        string
	Description string
}

func (i *System) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("id is required")
	}

	if i.Name == "" {
		i.Name = i.ID
	}

	return nil
}
