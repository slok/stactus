package prometheus_test

import (
	"context"
	"testing"
	"time"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/prometheus"
	utilfs "github.com/slok/stactus/internal/util/fs"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryCreatePromMetrics(t *testing.T) {
	tests := map[string]struct {
		ui         func() model.UI
		expMetrics string
		expErr     bool
	}{
		"Correct data should render correctly the metrics": {
			ui: func() model.UI {
				return model.UI{
					SystemDetails: []model.SystemDetails{
						{
							System: model.System{ID: "s1", Name: "System 1"},
							IRs: []*model.IncidentReport{
								{Impact: model.IncidentImpactMinor},
								{Impact: model.IncidentImpactMajor},
							},
						},
						{
							System: model.System{ID: "s2", Name: "System 2"},
							IRs: []*model.IncidentReport{
								{Impact: model.IncidentImpactMajor},
								{Impact: model.IncidentImpactCritical},
							},
						},
						{
							System: model.System{ID: "s3", Name: "System 3"},
							IRs: []*model.IncidentReport{
								{Impact: model.IncidentImpactMinor},
							},
						},
						{
							System: model.System{ID: "s4", Name: "System 4"},
							IRs: []*model.IncidentReport{
								{Impact: model.IncidentImpactMinor, End: time.Now()},
							},
						},
					},
					Settings: model.StatusPageSettings{
						Name: "test-SP",
					},
					Stats: model.UIStats{
						MTTR: 42 * time.Minute,
					},
					OpenedIRs: []*model.IncidentReport{
						{ID: "test1", Impact: model.IncidentImpactCritical},
						{ID: "test2", Impact: model.IncidentImpactMinor},
						{ID: "test3", Impact: model.IncidentImpactNone},
					},
				}
			},
			expMetrics: `
# HELP stactus_all_systems_status Tells if all systems are operational or not.
# TYPE stactus_all_systems_status gauge
stactus_all_systems_status{status_ok="false",status_page="test-SP"} 1
# HELP stactus_incident_mttr_seconds The MTTR based on all the incident history.
# TYPE stactus_incident_mttr_seconds gauge
stactus_incident_mttr_seconds{status_page="test-SP"} 2520
# HELP stactus_open_incident The details of open (not resolved) incidents.
# TYPE stactus_open_incident gauge
stactus_open_incident{id="test1",impact="critical",status_page="test-SP"} 1
stactus_open_incident{id="test2",impact="minor",status_page="test-SP"} 1
stactus_open_incident{id="test3",impact="none",status_page="test-SP"} 1
# HELP stactus_system_status Tells Systems are operational or not.
# TYPE stactus_system_status gauge
stactus_system_status{id="s1",impact="major",name="System 1",status_ok="false",status_page="test-SP"} 1
stactus_system_status{id="s2",impact="critical",name="System 2",status_ok="false",status_page="test-SP"} 1
stactus_system_status{id="s3",impact="minor",name="System 3",status_ok="false",status_page="test-SP"} 1
stactus_system_status{id="s4",impact="none",name="System 4",status_ok="true",status_page="test-SP"} 1
			`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := assert.New(t)
			assert := assert.New(t)

			fsm := utilfs.NewTestFileManager()
			repo, err := prometheus.NewFSRepository(prometheus.RepositoryConfig{
				FileManager:     fsm,
				MetricsFilePath: "test/metrics",
			})
			require.NoError(err)

			err = repo.CreatePromMetrics(context.TODO(), test.ui())
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				fsm.AssertEqual(t, "test/metrics", test.expMetrics)
			}
		})
	}
}
