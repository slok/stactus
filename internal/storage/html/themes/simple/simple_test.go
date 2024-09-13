package simple_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/themes/simple"
	"github.com/slok/stactus/internal/storage/html/util"
)

func TestCreateUI(t *testing.T) {

	tests := map[string]struct {
		themeCustomization simple.ThemeCustomization
		ui                 model.UI
		expectHTML         map[string][]string
		expErr             bool
	}{
		// "The static files have been rendered correctly.": {
		// 	themeCustomization: simple.ThemeCustomization{
		// 		BrandTitle: "MonkeyIsland",
		// 		BrandURL:   "https://monkeyisland.slok.dev",
		// 	},
		// 	ui: model.UI{},
		// 	expectHTML: map[string][]string{
		// 		"./static/main.css": {},
		// 	},
		// },
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			fm := util.NewTestFileManager()
			gen, err := simple.NewGenerator(simple.GeneratorConfig{
				FileManager:        fm,
				OutPath:            "./",
				ThemeCustomization: test.themeCustomization,
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
