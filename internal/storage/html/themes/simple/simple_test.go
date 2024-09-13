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
				"./static/main.js":  {},
			},
		},

		"The nav should be rendered correctly.": {
			themeCustomization: simple.ThemeCustomization{
				BrandTitle: "MonkeyIsland",
				BrandURL:   "https://monkeyisland.slok.dev",
			},
			ui: model.UI{},
			expectHTML: map[string][]string{
				"./index.html": {
					`<h2>MonkeyIsland status </h2>`,
					`<li><a href="/history/0">History</a></li>`,
					`<li><a href="/">Status</a></li>`,
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
					`<nav>`, // We have the nav var.
					`Powered by <a href="https://github.com/slok/stactus">Stactus</a>.`, // We have the footer.

					// Status is ok.
					`<strong>All systems operational</strong>`,
					`<article> Test 1 <span data-tooltip="Something test 1"><i class="ph-thin ph-question"></i></span><span class="move-right"> <i style="font-size: 150%;" class="ph-fill ph-check-circle text-ok"></i> </span><div> <small> Normal </small> </div> </article>`,
					`<article> Test 2 <span data-tooltip="Something test 2"><i class="ph-thin ph-question"></i></span><span class="move-right"> <i style="font-size: 150%;" class="ph-fill ph-check-circle text-ok"></i> </span><div> <small> Normal </small> </div> </article>`,
					`<article> Test 3 <span data-tooltip="Something test 3"><i class="ph-thin ph-question"></i></span><span class="move-right"> <i style="font-size: 150%;" class="ph-fill ph-check-circle text-ok"></i> </span><div> <small> Normal </small> </div> </article>`,
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
						ID:        "ir42",
						Name:      "Oh snap!",
						SystemIDs: []string{"test1"},
						Start:     t0,
						Impact:    model.IncidentImpactCritical,
						Timeline: []model.IncidentReportEvent{
							{Description: "There is a problem 3", TS: t0.Add(33 * time.Minute)},
							{Description: "There is a problem 2", TS: t0.Add(22 * time.Minute)},
							{Description: "There is a problem 1", TS: t0.Add(11 * time.Minute)},
						},
					},

					{
						ID:        "ir99",
						Name:      "fuuuuuuuuuuu",
						SystemIDs: []string{"test2"},
						Start:     t0,
						Impact:    model.IncidentImpactMajor,
						Timeline: []model.IncidentReportEvent{
							{Description: "something **something** 9", TS: t0.Add(199 * time.Minute)},
							{Description: "something something 6", TS: t0.Add(87 * time.Minute)},
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
					// Ongoing incident 1 info.
					`<article class="box-impact-critical"> <header class="header-impact-critical">`,                   // We have the impact.
					`<h4><a href="/ir/ir42" class="incident-title"> Oh snap!</a></h4>`,                                // We have the title and link
					`<p>There is a problem 3</p>`,                                                                     // Regular plain description.
					`<small>Latest update at <span x-init="renderTSUnixPrettyNoYear($el)">-1815344697</span></small>`, // TS set for client JS libs.

					// Ongoing incident 2 info.
					`<article class="box-impact-major"> <header class="header-impact-major">`,                         // We have the impact.
					`<h4><a href="/ir/ir99" class="incident-title"> fuuuuuuuuuuu</a></h4>`,                            // We have the title and link
					`<p>something <strong>something</strong> 9</p>`,                                                   //  Markdown description.
					`<small>Latest update at <span x-init="renderTSUnixPrettyNoYear($el)">-1815334737</span></small>`, // TS set for client JS libs.

					// Systems status.
					`<article> Test 1 <span data-tooltip="Something test 1"><i class="ph-thin ph-question"></i></span><span class="move-right"> <i style="font-size: 150%;" class="ph-fill ph-check-circle text-ok"></i> </span><div> <small> Normal </small> </div> </article>`,
					`<article> Test 2 <span data-tooltip="Something test 2"><i class="ph-thin ph-question"></i></span><span class="move-right"> <i style="font-size: 150%;" class="ph-fill ph-check-circle text-ok"></i> </span><div> <small> Normal </small> </div> </article>`,
					`<article> Test 3 <span data-tooltip="Something test 3"><i class="ph-thin ph-question"></i></span><span class="move-right"> <i style="font-size: 150%;" class="ph-fill ph-warning-circle text-"></i> </span><div> <small> Degraded </small> </div> </article>`,
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
						Impact:    model.IncidentImpactCritical,
						Timeline: []model.IncidentReportEvent{
							{Description: "Some **detail** 12"},
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
					`<nav>`, // We have the nav var.
					`Powered by <a href="https://github.com/slok/stactus">Stactus</a>.`, // We have the footer.

					`<h1>Incident History</h1> `, // We have the title.

					// Incident 1.
					`<h4><a href="/ir/ir-1" class="incident-title-major"> Incident report 1</a>`, // We have the title with impact an details URL.
					`<p>Some detail 11</p>`, // We have plain text details.
					`<span x-init="renderTSUnixPrettyNoYear($el)">-1815346677</span> - <span x-init="renderTSUnixPrettyNoYear($el)">-1815339477</span>`, // Start and end TS.
					`<mark class="resolved">Resolved</mark>`, // Resolved mark.

					// Incident 2.
					`<h4><a href="/ir/ir-2" class="incident-title-critical"> Incident report 2</a></h4>`,
					`<p>Some <strong>detail</strong> 12</p>`,                          // We have markdown details.
					`<span x-init="renderTSUnixPrettyNoYear($el)">-1815310677</span>`, // Start.
					`<mark class="unresolved">Ongoing</mark>`,                         // Unresolved mark.

					// Pagination.
					`<a href="/history/1" role="button"> ⮜ Previous </a>`,
				},

				"./history/1.html": {
					`<nav>`, // We have the nav var.
					`Powered by <a href="https://github.com/slok/stactus">Stactus</a>.`, // We have the footer.

					`<h1>Incident History</h1> `, // We have the title.

					// Incident 3.
					`<h4><a href="/ir/ir-3" class="incident-title-minor"> Incident report 3</a></h4>`, // We have the title with impact an details URL.
					`<p>Some detail 13</p>`, // We have plain text details.
					`<span x-init="renderTSUnixPrettyNoYear($el)">-1815274677</span> - <span x-init="renderTSUnixPrettyNoYear($el)">-1815256677</span>`, // Start and end TS.
					`<mark class="resolved">Resolved</mark>`, // Resolved mark.

					// Pagination.
					`<a href="/history/0" role="button"> Next ⮞ </a> </span>`,
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
					`<nav>`, // We have the nav var.
					`Powered by <a href="https://github.com/slok/stactus">Stactus</a>.`, // We have the footer.

					`class="text-major">Incident report 1</h1>`,    // We have the IR title with impact.
					`<article class="incident-resolved">`,          // Resolved mark.
					`<strong>Incident resolved in 2h0m0s</strong>`, // We have the time took to be resolved.

					// Timeline.
					`<blockquote> <h4> Resolved </h4> <p>Some detail 13</p><footer> <cite x-init="renderTSUnixPrettyNoYear($el)">-1815346377</cite> </footer> </blockquote>`,
					`<blockquote> <h4> Update </h4> <p>Some detail 12</p><footer> <cite x-init="renderTSUnixPrettyNoYear($el)">-1815346497</cite> </footer> </blockquote>`,
					`<blockquote> <h4> Investigating </h4> <p>Some detail 13</p><footer> <cite x-init="renderTSUnixPrettyNoYear($el)">-1815346557</cite> </footer> </blockquote>`,
				},

				"./ir/0987654321.html": {
					`<nav>`, // We have the nav var.
					`Powered by <a href="https://github.com/slok/stactus">Stactus</a>.`, // We have the footer.

					`class="text-minor">Incident report 2</h1>`, // We have the IR title with impact.
					`<article class="incident-ongoing-minor">`,  // Not resolved mark with impact.
					`<strong>Incident ongoing</strong>`,         // We have the incident ongoing message.

					// Timeline.
					`<blockquote> <h4> Investigating </h4> <p>Some detail 23</p><footer> <cite x-init="renderTSUnixPrettyNoYear($el)">-1815345777</cite> </footer> </blockquote>`,
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
