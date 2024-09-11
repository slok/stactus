package simple_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/html/themes/simple"
	"github.com/slok/stactus/internal/storage/html/util"
)

func TestCreateUI(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "1912-06-23T01:02:03Z")

	tests := map[string]struct {
		themeCustomization simple.ThemeCustomization
		ui                 model.UI
		expectHTML         map[string][]string
		expErr             bool
	}{
		"The static files have been rendered correctly.": {
			themeCustomization: simple.ThemeCustomization{
				BrandTitle: "MonkeyIsland",
				BrandURL:   "https://monkeyisland.slok.dev",
			},
			ui: model.UI{},
			expectHTML: map[string][]string{
				"./static/main.css": {},
			},
		},

		"The index dashboard shared components should render correctly.": {
			themeCustomization: simple.ThemeCustomization{
				BrandTitle: "MonkeyIsland",
				BrandURL:   "https://monkeyisland.slok.dev",
			},
			ui: model.UI{},
			expectHTML: map[string][]string{
				"./index.html": {
					`<title>MonkeyIsland status</title>`,                                        // We have the title.
					`<h1> <a href="https://monkeyisland.slok.dev">MonkeyIsland</a> status</h1>`, // We have the brand header.
					`<a href="/history/0">Incident history</a>`,                                 // We have history pagination.
				},
			},
		},

		"If all systems are ok it should be reflected.": {
			themeCustomization: simple.ThemeCustomization{
				BrandTitle: "MonkeyIsland",
				BrandURL:   "https://monkeyisland.slok.dev",
			},
			ui: model.UI{
				SystemDetails: []model.SystemDetails{
					{
						System: model.System{ID: "test1", Name: "Test 1", Description: "Something test 1"},
					},
					{
						System: model.System{ID: "test2", Name: "Test 2", Description: "Something test 2"},
					},
					{
						System: model.System{ID: "test3", Name: "Test 3", Description: "Something test 3"},
					},
				},
			},
			expectHTML: map[string][]string{
				"./index.html": {
					`<strong>All systems ok</strong>`,                      // We have the message that all systems are ok.
					`title="Something test 1"> <strong>Test 1</strong> OK`, // We have system 1 status.
					`title="Something test 2"> <strong>Test 2</strong> OK`, // We have system 2 status.
					`title="Something test 3"> <strong>Test 3</strong> OK`, // We have system 3 status.
				},
			},
		},

		"If any systems is not ok it should be reflected.": {
			themeCustomization: simple.ThemeCustomization{
				BrandTitle: "MonkeyIsland",
				BrandURL:   "https://monkeyisland.slok.dev",
			},
			ui: model.UI{
				OpenedIRs: []*model.IncidentReport{
					{
						Timeline: []model.IncidentReportEvent{
							{Description: "There is a problem"},
						},
					},
				},
				SystemDetails: []model.SystemDetails{
					{
						System: model.System{ID: "test1", Name: "Test 1", Description: "Something test 1"},
					},
					{
						System: model.System{ID: "test2", Name: "Test 2", Description: "Something test 2"},
					},
					{
						System:   model.System{ID: "test3", Name: "Test 3", Description: "Something test 3"},
						LatestIR: &model.IncidentReport{},
					},
				},
			},
			expectHTML: map[string][]string{
				"./index.html": {
					`There is a problem`, // We have the message that some system is not ok.
					`title="Something test 1"> <strong>Test 1</strong> OK`,       // We have system 1 status.
					`title="Something test 2"> <strong>Test 2</strong> OK`,       // We have system 2 status.
					`title="Something test 3"> <strong>Test 3</strong> Degraded`, // We have system 3 status.
				},
			},
		},

		"History pagination should be rendered correctly.": {
			themeCustomization: simple.ThemeCustomization{
				BrandTitle:       "MonkeyIsland",
				BrandURL:         "https://monkeyisland.slok.dev",
				HistoryIRPerPage: 2,
			},
			ui: model.UI{
				SystemDetails: []model.SystemDetails{
					{System: model.System{ID: "test1", Name: "Test 1", Description: "Something test 1"}},
				},
				History: []*model.IncidentReport{
					{
						ID:        "ir-1",
						Name:      "Incident report 1",
						SystemIDs: []string{"test1"},
						Start:     t0,
						End:       t0.Add(2 * time.Hour),
						Impact:    model.IncidentImpactMajor,
						Timeline: []model.IncidentReportEvent{
							{Description: "Some detail 11"},
							{Description: "Some detail 21"},
						},
					},
					{
						ID:        "ir-2",
						Name:      "Incident report 2",
						SystemIDs: []string{"test1"},
						Start:     t0.Add(10 * time.Hour),
						End:       t0.Add(15 * time.Hour),
						Impact:    model.IncidentImpactCritical,
						Timeline: []model.IncidentReportEvent{
							{Description: "Some detail 12"},
							{Description: "Some detail 22"},
						},
					},
					{
						ID:        "ir-3",
						Name:      "Incident report 3",
						SystemIDs: []string{"test1"},
						Start:     t0.Add(20 * time.Hour),
						End:       t0.Add(25 * time.Hour),
						Impact:    model.IncidentImpactMinor,
						Timeline: []model.IncidentReportEvent{
							{Description: "Some detail 13"},
							{Description: "Some detail 23"},
						},
					},
				},
			},
			expectHTML: map[string][]string{
				"./history/0.html": {
					`<h1>Incident History</h1> `, // We have the title.

					// Incident 1.
					`<div class="box impact-major">`,
					`<h2><a href="/ir/ir-1">Incident report 1</a></h2>`,
					`<p> Some detail 11 </p>`,
					`Jun 23, 01:02 - Jun 23, 03:02`,

					// Incident 2.
					`<div class="box impact-critical">`,
					`<h2><a href="/ir/ir-2">Incident report 2</a></h2>`,
					`<p> Some detail 12 </p>`,
					`Jun 23, 11:02 - Jun 23, 16:02`,

					// Pagination.
					`<a href="/history/1"> ⮜ Previous </a>`,
				},

				"./history/1.html": {
					`<h1>Incident History</h1> `, // We have the title.

					// Incident 3.
					`<div class="box impact-minor">`,
					`<h2><a href="/ir/ir-3">Incident report 3</a></h2>`,
					`<p> Some detail 13 </p>`,
					`Jun 23, 21:02 - Jun 24, 02:02`,

					// Pagination.
					`<a href="/history/0"> Next ⮞ </a>`,
				},
			},
		},

		"IR details should be rendered correctly.": {
			themeCustomization: simple.ThemeCustomization{
				BrandTitle: "MonkeyIsland",
				BrandURL:   "https://monkeyisland.slok.dev",
			},
			ui: model.UI{
				History: []*model.IncidentReport{
					{
						ID:        "1234567890",
						Name:      "Incident report 1",
						SystemIDs: []string{"test1"},
						Start:     t0,
						End:       t0.Add(2 * time.Hour),
						Impact:    model.IncidentImpactMajor,
						Timeline: []model.IncidentReportEvent{
							{TS: t0.Add(5 * time.Minute), Kind: model.IncidentUpdateKindResolved, Description: "Some detail 13"},
							{TS: t0.Add(3 * time.Minute), Kind: model.IncidentUpdateKindUpdate, Description: "Some detail 12"},
							{TS: t0.Add(2 * time.Minute), Kind: model.IncidentUpdateKindInvestigating, Description: "Some detail 13"},
						},
					},

					{
						ID:        "0987654321",
						Name:      "Incident report 2",
						SystemIDs: []string{"test1"},
						Start:     t0,
						Impact:    model.IncidentImpactMinor,
						Timeline: []model.IncidentReportEvent{
							{TS: t0.Add(15 * time.Minute), Kind: model.IncidentUpdateKindInvestigating, Description: "Some detail 23"},
						},
					},
				},
			},
			expectHTML: map[string][]string{
				"./ir/1234567890.html": {
					`<h1>Incident report 1</h1>`,  // We have the title page.
					`<a href="/index">Status</a>`, // We have the link to the status.

					// Details.
					`<h2> Details </h2>`,
					`<td><strong>ID</strong></td> <td>1234567890</td>`,
					`<td><strong>Severity</strong></td> <td>Major</td>`,
					`<td><strong>Started</strong></td> <td>Jun 23, 01:02</td>`,
					`<td><strong>End</strong></td> <td>Jun 23, 03:02</td>`,
					`<td><strong>Duration</strong></td> <td>2h0m0s</td>`,

					// Timeline.
					`<h2> Timeline </h2>`,
					`<h3> Resolved - Jun 23, 01:07 </h3> <p> Some detail 13 </p>`,
					`<h3> Update - Jun 23, 01:05 </h3> <p> Some detail 12 </p>`,
					`<h3> Investigating - Jun 23, 01:04 </h3> <p> Some detail 13 </p>`,
				},

				"./ir/0987654321.html": {
					`<h1>Incident report 2</h1>`,  // We have the title page.
					`<a href="/index">Status</a>`, // We have the link to the status.

					// Details.
					`<h2> Details </h2>`,
					`<td><strong>ID</strong></td> <td>0987654321</td>`,
					`<td><strong>Severity</strong></td> <td>Minor</td>`,
					`<td><strong>Started</strong></td> <td>Jun 23, 01:02</td>`,
					`<td><strong>End</strong></td> <td> Ongoing </td>`,

					// Timeline.
					`<h2> Timeline </h2>`,
					`<h3> Investigating - Jun 23, 01:17 </h3> <p> Some detail 23 </p> `,
				},
			},
		},
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
