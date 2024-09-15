package model_test

import (
	"testing"
	"time"

	"github.com/slok/stactus/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	t0 = time.Date(2024, 9, 13, 5, 42, 0, 0, time.UTC)
	t1 = time.Date(2024, 9, 13, 5, 59, 0, 0, time.UTC)
)

func getBaseIncidentReport() model.IncidentReport {
	return model.IncidentReport{
		ID:        "test-id",
		Name:      "Test 1",
		Start:     t0,
		End:       t1,
		SystemIDs: []string{"system1", "system2"},
		Impact:    model.IncidentImpactCritical,
		Timeline: []model.IncidentReportEvent{
			{Description: "desc1", Kind: model.IncidentUpdateKindResolved, TS: t1},
			{Description: "desc2", Kind: model.IncidentUpdateKindInvestigating, TS: t0},
		},
	}
}

func TestIncidentReportValidate(t *testing.T) {
	tests := map[string]struct {
		ir     func() model.IncidentReport
		expIR  func() model.IncidentReport
		expErr bool
	}{
		"A correct system should validate correctly.": {
			ir:    getBaseIncidentReport,
			expIR: getBaseIncidentReport,
		},

		"Not having a timeline should fail.": {
			ir: func() model.IncidentReport {
				ir := getBaseIncidentReport()
				ir.Timeline = nil
				return ir
			},
			expErr: true,
		},

		"Not having an ID should fail.": {
			ir: func() model.IncidentReport {
				ir := getBaseIncidentReport()
				ir.ID = ""
				return ir
			},
			expErr: true,
		},

		"Not having a name should fail.": {
			ir: func() model.IncidentReport {
				ir := getBaseIncidentReport()
				ir.Name = ""
				return ir
			},
			expErr: true,
		},

		"Not having start TS, should be set from the events.": {
			ir: func() model.IncidentReport {
				ir := getBaseIncidentReport()
				ir.Start = time.Time{}
				return ir
			},
			expIR: getBaseIncidentReport,
		},

		"Not having end TS, should be set from the events.": {
			ir: func() model.IncidentReport {
				ir := getBaseIncidentReport()
				ir.End = time.Time{}
				return ir
			},
			expIR: getBaseIncidentReport,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			ir := test.ir()
			err := ir.Validate()
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expIR(), ir)
			}
		})
	}
}
