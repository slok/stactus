package model_test

import (
	"testing"

	"github.com/slok/stactus/internal/model"
	"github.com/stretchr/testify/assert"
)

func getBaseSystem() model.System {
	return model.System{
		ID:          "test-id",
		Name:        "Test 1",
		Description: "test system",
	}
}

func TestSystemValidate(t *testing.T) {
	tests := map[string]struct {
		system    func() model.System
		expSystem func() model.System
		expErr    bool
	}{
		"A correct system should validate correctly.": {
			system:    getBaseSystem,
			expSystem: getBaseSystem,
		},

		"A missing id should fail.": {
			system: func() model.System {
				s := getBaseSystem()
				s.ID = ""
				return s
			},
			expErr: true,
		},

		"A missing name should default to ID.": {
			system: func() model.System {
				s := getBaseSystem()
				s.Name = ""
				return s
			},
			expSystem: func() model.System {
				s := getBaseSystem()
				s.Name = "test-id"
				return s
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			s := test.system()
			err := s.Validate()
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expSystem(), s)
			}
		})
	}
}
