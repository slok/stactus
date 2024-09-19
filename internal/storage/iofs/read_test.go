package iofs_test

import (
	"context"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/iofs"
)

var (
	testStatusFile = `
version: stactus/v1
name: SomethingIO
url: https://something.test.test.somethingdsadsadsad.com
systems:
  - id: system1
    name: System 1
    description: This is a description of system1
  - id: system2
    name: System 2
    description: This is a description of system2
`
	testSystems = []model.System{
		{ID: "system1", Name: "System 1", Description: "This is a description of system1"},
		{ID: "system2", Name: "System 2", Description: "This is a description of system2"},
	}

	testSettings = model.StatusPageSettings{Name: "SomethingIO", URL: "https://something.test.test.somethingdsadsadsad.com"}
)

func getIRForTSFormats(ts1, ts2 string) []byte {
	return []byte(fmt.Sprintf(`
version: incident/v1
id: test-0001
name: incident 1
impact: minor
systems: ["system1"]
timeline:
  - ts: %s
    investigating: true
    description: ts1

  - ts: %s
    resolved: true
    description: ts2
`, ts1, ts2))
}

func getIRForTSFormatsResult(ts1, ts2 time.Time) model.IncidentReport {
	return model.IncidentReport{ID: "test-0001", Name: "incident 1", SystemIDs: []string{"system1"}, Impact: "minor",
		Start:    time.Date(2024, 9, 13, 5, 42, 0, 0, time.UTC),
		End:      time.Date(2024, 9, 13, 5, 59, 0, 0, time.UTC),
		Duration: ts2.Sub(ts1),
		Timeline: []model.IncidentReportEvent{
			{Description: "ts2", Kind: model.IncidentUpdateKindResolved, TS: ts2},
			{Description: "ts1", Kind: model.IncidentUpdateKindInvestigating, TS: ts1},
		},
	}
}

func TestReadRepository(t *testing.T) {
	t0 := time.Date(2024, 9, 13, 5, 42, 0, 0, time.UTC)
	t1 := time.Date(2024, 9, 13, 5, 59, 0, 0, time.UTC)

	tests := map[string]struct {
		fs          func() fs.FS
		stactusFile string
		expSettings model.StatusPageSettings
		expSystems  []model.System
		expIRs      []model.IncidentReport
		expErr      bool
	}{
		"An empty stactus file should fail.": {
			fs:          func() fs.FS { return fstest.MapFS{} },
			stactusFile: "",
			expSettings: model.StatusPageSettings{},
			expSystems:  []model.System{},
			expIRs:      []model.IncidentReport{},
			expErr:      true,
		},

		"Having no systems should fail.": {
			fs:          func() fs.FS { return fstest.MapFS{} },
			stactusFile: "api: stactus/v1\nsystems: []",
			expSystems:  []model.System{},
			expIRs:      []model.IncidentReport{},
			expErr:      true,
		},

		"A correct stactus file should return the systems and settings.": {
			fs:          func() fs.FS { return fstest.MapFS{} },
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expIRs:      []model.IncidentReport{},
		},

		"Incident reports should be loaded correctly.": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: []byte(`
version: incident/v1
id: test-0001
name: incident 1
impact: minor
systems: ["system1"]
timeline:
  - ts: 2024/09/13 05:42
    investigating: true
    description: desc 1

  - ts: 2024/09/13 05:48
    description: desc 2

  - ts: 2024/09/13 05:59
    resolved: true
    description: desc 3
`)}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expIRs: []model.IncidentReport{
				{ID: "test-0001", Name: "incident 1", SystemIDs: []string{"system1"}, Impact: "minor",
					Start:    time.Date(2024, 9, 13, 5, 42, 0, 0, time.UTC),
					End:      time.Date(2024, 9, 13, 5, 59, 0, 0, time.UTC),
					Duration: 17 * time.Minute,
					Timeline: []model.IncidentReportEvent{
						{Description: "desc 3", Kind: model.IncidentUpdateKindResolved, TS: time.Date(2024, 9, 13, 5, 59, 0, 0, time.UTC)},
						{Description: "desc 2", Kind: model.IncidentUpdateKindUpdate, TS: time.Date(2024, 9, 13, 5, 48, 0, 0, time.UTC)},
						{Description: "desc 1", Kind: model.IncidentUpdateKindInvestigating, TS: time.Date(2024, 9, 13, 5, 42, 0, 0, time.UTC)},
					},
				},
			},
		},

		"Different TS formats should be loaded correctly (pretty format).": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("  2024/09/13 05:42", "2024-09-13 05:59   ")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expIRs:      []model.IncidentReport{getIRForTSFormatsResult(t0, t1)},
		},

		"Different TS formats should be loaded correctly (RFC3339).": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("2024-09-13T05:42:00Z", "2024-09-13T05:59:00.000000000Z")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expIRs:      []model.IncidentReport{getIRForTSFormatsResult(t0, t1)},
		},

		"Different TS formats should be loaded correctly (Relative duration).": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("2024/09/13 05:42", "+17m")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expIRs:      []model.IncidentReport{getIRForTSFormatsResult(t0, t1)},
		},

		"Different relative TS formats should fail (Relative duration bad format).": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("2024/09/13 05:42", "+17d")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expErr:      true,
		},

		"Different relative TS formats should fail if not previous date (Relative duration).": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("+17m", "2024/09/13 05:42")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expErr:      true,
		},

		"Different TS formats should be loaded correctly (Relative without day).": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("2024/09/13 05:42", "05:59")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expIRs:      []model.IncidentReport{getIRForTSFormatsResult(t0, t1)},
		},

		"Different relative TS formats should be loaded fail if not previous date (Relative without day).": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("05:59", "2024/09/13 05:42")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expErr:      true,
		},

		"Unsupported TS formats should fail.": {
			fs: func() fs.FS {
				fs := fstest.MapFS{}
				fs["ir.yaml"] = &fstest.MapFile{Data: getIRForTSFormats("Mon, 02 Jan 2006 15:04:05 MST", "2024/09/13 05:42")}
				return fs
			},
			stactusFile: testStatusFile,
			expSettings: testSettings,
			expSystems:  testSystems,
			expErr:      true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo, err := iofs.NewReadRepository(context.TODO(), iofs.ReadRepositoryConfig{
				IncidentsFS:     test.fs(),
				StactusFileData: test.stactusFile,
			})
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				gotSettings, _ := repo.GetStatusPageSettings(context.TODO())
				assert.Equal(test.expSettings, *gotSettings)
				gotSystems, _ := repo.ListAllSystems(context.TODO())
				assert.Equal(test.expSystems, gotSystems)
				gotIRs, _ := repo.ListAllIncidentReports(context.TODO())
				assert.Equal(test.expIRs, gotIRs)
			}
		})
	}
}
