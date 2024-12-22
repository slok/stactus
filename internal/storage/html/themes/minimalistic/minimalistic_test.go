package minimalistic_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/themes/minimalistic"
	utilfs "github.com/slok/stactus/internal/util/fs"
)

func TestCreateUI(t *testing.T) {
	//t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		ui         model.UI
		expectHTML map[string][]string
		expErr     bool
	}{
		"The static files have been rendered correctly.": {
			ui: model.UI{
				Settings: model.StatusPageSettings{
					Name: "MonkeyIsland",
					URL:  "https://monkeyisland.slok.dev",
				},
			},
			expectHTML: map[string][]string{
				"./static/main.css": {},
				"./static/main.js":  {},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			fm := utilfs.NewTestFileManager()
			gen, err := minimalistic.NewGenerator(minimalistic.GeneratorConfig{
				FileManager: fm,
				OutPath:     "./",
			})
			require.NoError(err)
			err = gen.CreateUI(context.TODO(), test.ui)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				for file, exp := range test.expectHTML {
					fm.AssertContains(t, file, exp)
				}
			}
		})
	}
}
