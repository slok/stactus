package model_test

import (
	"testing"

	"github.com/slok/stactus/internal/model"
	"github.com/stretchr/testify/assert"
)

func getBaseSettings() model.StatusPageSettings {
	return model.StatusPageSettings{
		Name: "Test 1",
		URL:  "https://something.io",
	}
}

func TestStatusPageSettingsValidate(t *testing.T) {
	tests := map[string]struct {
		system                func() model.StatusPageSettings
		expStatusPageSettings func() model.StatusPageSettings
		expErr                bool
	}{
		"A correct system should validate correctly.": {
			system:                getBaseSettings,
			expStatusPageSettings: getBaseSettings,
		},

		"A missing name should fail.": {
			system: func() model.StatusPageSettings {
				s := getBaseSettings()
				s.Name = ""
				return s
			},
			expErr: true,
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
				assert.Equal(test.expStatusPageSettings(), s)
			}
		})
	}
}
