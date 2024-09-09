package gh_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/themes/gh"
	"github.com/slok/stactus/internal/storage/html/util"
)

func TestCreateUI(t *testing.T) {
	tests := map[string]struct {
		ui         model.UI
		expectHTML map[string][]string
		expErr     bool
	}{
		"The static files have been rendered correctly.": {
			ui: model.UI{},
			expectHTML: map[string][]string{
				"./static/main.css": {},
				"./static/main.js":  {},
			},
		},

		"The index dashboard should render correctly.": {
			ui: model.UI{},
			expectHTML: map[string][]string{
				"./index.html": {
					`<title>Stactus</title>`, // We have the title.
					`<a class="btn btn-ghost text-xl" href="https://github.com">Github</a>`,                                                   // We have the Company name and link.
					`<img src="https://user-images.githubusercontent.com/19292210/60553863-044dd200-9cea-11e9-987e-7db84449f215.png"> </img>`, // We have the banner.
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			fm := util.NewTestFileManager()
			gen, err := gh.NewGenerator(gh.GeneratorConfig{
				FileManager: fm,
				OutPath:     "./",
				ThemeCustomization: gh.ThemeCustomization{
					BrandTitle:     "Github",
					BrandURL:       "https://github.com",
					BannerImageURL: "https://user-images.githubusercontent.com/19292210/60553863-044dd200-9cea-11e9-987e-7db84449f215.png",
					LogoURL:        "https://raw.githubusercontent.com/gilbarbara/logos/main/logos/github-icon.svg",
				},
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
