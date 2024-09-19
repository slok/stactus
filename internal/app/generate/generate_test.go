package generate_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/storagemock"
)

func TestGenerate(t *testing.T) {
	type mocks struct {
		mstg *storagemock.StatusPageSettingsGetter
		msg  *storagemock.SystemGetter
		mig  *storagemock.IncidentReportGetter
		muc  *storagemock.UICreator
		mpc  *storagemock.PromMetricsCreator
	}

	t0 := time.Now()

	tests := map[string]struct {
		mock    func(m mocks)
		req     generate.GenerateReq
		expResp generate.GenerateResp
		expErr  bool
	}{

		"If getting settings returns an error, it should fail.": {
			mock: func(m mocks) {
				m.mstg.On("GetStatusPageSettings", mock.Anything).Once().Return(nil, fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"If listing systems returns an error, it should fail.": {
			mock: func(m mocks) {
				m.mstg.On("GetStatusPageSettings", mock.Anything).Once().Return(&model.StatusPageSettings{Name: "test1", URL: "https://test.io"}, nil)
				m.msg.On("ListAllSystems", mock.Anything).Once().Return(nil, fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"If listing IRs returns an error, it should fail.": {
			mock: func(m mocks) {
				m.mstg.On("GetStatusPageSettings", mock.Anything).Once().Return(&model.StatusPageSettings{Name: "test1", URL: "https://test.io"}, nil)
				m.msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{}, nil)
				m.mig.On("ListAllIncidentReports", mock.Anything).Return(nil, fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"If UI generation returns an error, it should fail.": {
			mock: func(m mocks) {
				m.mstg.On("GetStatusPageSettings", mock.Anything).Once().Return(&model.StatusPageSettings{Name: "test1", URL: "https://test.io"}, nil)
				m.msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{}, nil)
				m.mig.On("ListAllIncidentReports", mock.Anything).Return([]model.IncidentReport{}, nil)
				m.muc.On("CreateUI", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"Creating the UI correctly should generate the UI (No IRs).": {
			mock: func(m mocks) {
				m.mstg.On("GetStatusPageSettings", mock.Anything).Once().Return(&model.StatusPageSettings{Name: "test1", URL: "https://test.io"}, nil)
				m.msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{
					{ID: "test1", Name: "Test 1", Description: "Something 1"},
					{ID: "test2", Name: "Test 2", Description: "Something 2"},
					{ID: "test3", Name: "Test 3", Description: "Something 3"},
				}, nil)

				m.mig.On("ListAllIncidentReports", mock.Anything).Return([]model.IncidentReport{}, nil)

				exp := model.UI{
					Stats: model.UIStats{
						TotalSystems: 3,
					},
					Settings: model.StatusPageSettings{
						Name: "test1",
						URL:  "https://test.io",
					},
					OpenedIRs: []*model.IncidentReport{},
					History:   []*model.IncidentReport{},
					SystemDetails: []model.SystemDetails{
						{
							System: model.System{ID: "test1", Name: "Test 1", Description: "Something 1"},
						},
						{
							System: model.System{ID: "test2", Name: "Test 2", Description: "Something 2"},
						},
						{
							System: model.System{ID: "test3", Name: "Test 3", Description: "Something 3"},
						},
					},
				}
				m.muc.On("CreateUI", mock.Anything, exp).Once().Return(nil)

				m.mpc.On("CreatePromMetrics", mock.Anything, exp).Once().Return(nil)
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
		},

		"Creating the UI correctly should generate the UI (service with IRs).": {
			mock: func(m mocks) {
				m.mstg.On("GetStatusPageSettings", mock.Anything).Once().Return(&model.StatusPageSettings{Name: "test1", URL: "https://test.io"}, nil)
				m.msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{
					{ID: "test1", Name: "Test 1", Description: "Something 1"},
					{ID: "test2", Name: "Test 2", Description: "Something 2"},
					{ID: "test3", Name: "Test 3", Description: "Something 3"},
				}, nil)

				m.mig.On("ListAllIncidentReports", mock.Anything).Return([]model.IncidentReport{
					{
						ID:        "ir1",
						SystemIDs: []string{"test2"},
						Name:      "IR 1",
						Start:     t0,
						Timeline: []model.IncidentReportEvent{
							{Description: "desc1"},
						},
					},
					{
						ID:        "ir3",
						SystemIDs: []string{"test3"},
						Name:      "IR 3",
						Start:     t0.Add(-3 * time.Hour),
						End:       t0.Add(-2 * time.Hour),
						Duration:  1 * time.Hour,
					},
					{
						ID:        "ir2",
						SystemIDs: []string{"test2", "test3"},
						Name:      "IR 2",
						Start:     t0.Add(-10 * time.Hour),
						End:       t0.Add(-4 * time.Hour),
						Duration:  6 * time.Hour,
					},
				}, nil)

				exp := model.UI{
					Stats: model.UIStats{
						TotalSystems: 3,
						TotalOpenIRs: 1,
						TotalIRs:     3,
						MTTR:         210 * time.Minute,
					},
					Settings: model.StatusPageSettings{
						Name: "test1",
						URL:  "https://something-new.io",
					},
					OpenedIRs: []*model.IncidentReport{
						{ID: "ir1", SystemIDs: []string{"test2"}, Name: "IR 1", Start: t0, Timeline: []model.IncidentReportEvent{{Description: "desc1"}}},
					},
					History: []*model.IncidentReport{
						{ID: "ir1", SystemIDs: []string{"test2"}, Name: "IR 1", Start: t0, Timeline: []model.IncidentReportEvent{{Description: "desc1"}}},
						{ID: "ir3", SystemIDs: []string{"test3"}, Name: "IR 3", Duration: 1 * time.Hour, Start: t0.Add(-3 * time.Hour), End: t0.Add(-2 * time.Hour)},
						{ID: "ir2", SystemIDs: []string{"test2", "test3"}, Name: "IR 2", Duration: 6 * time.Hour, Start: t0.Add(-10 * time.Hour), End: t0.Add(-4 * time.Hour)},
					},
					SystemDetails: []model.SystemDetails{
						{
							System: model.System{ID: "test1", Name: "Test 1", Description: "Something 1"},
						},
						{
							System:   model.System{ID: "test2", Name: "Test 2", Description: "Something 2"},
							LatestIR: &model.IncidentReport{ID: "ir1", SystemIDs: []string{"test2"}, Name: "IR 1", Start: t0, Timeline: []model.IncidentReportEvent{{Description: "desc1"}}},
							IRs: []*model.IncidentReport{
								{ID: "ir1", SystemIDs: []string{"test2"}, Name: "IR 1", Start: t0, Timeline: []model.IncidentReportEvent{{Description: "desc1"}}},
								{ID: "ir2", SystemIDs: []string{"test2", "test3"}, Name: "IR 2", Duration: 6 * time.Hour, Start: t0.Add(-10 * time.Hour), End: t0.Add(-4 * time.Hour)},
							},
						},
						{
							System:   model.System{ID: "test3", Name: "Test 3", Description: "Something 3"},
							LatestIR: &model.IncidentReport{ID: "ir3", SystemIDs: []string{"test3"}, Name: "IR 3", Duration: 1 * time.Hour, Start: t0.Add(-3 * time.Hour), End: t0.Add(-2 * time.Hour)},
							IRs: []*model.IncidentReport{
								{ID: "ir3", SystemIDs: []string{"test3"}, Name: "IR 3", Duration: 1 * time.Hour, Start: t0.Add(-3 * time.Hour), End: t0.Add(-2 * time.Hour)},
								{ID: "ir2", SystemIDs: []string{"test2", "test3"}, Name: "IR 2", Duration: 6 * time.Hour, Start: t0.Add(-10 * time.Hour), End: t0.Add(-4 * time.Hour)},
							},
						},
					},
				}
				m.muc.On("CreateUI", mock.Anything, exp).Once().Return(nil)

				m.mpc.On("CreatePromMetrics", mock.Anything, exp).Once().Return(nil)
			},
			req:     generate.GenerateReq{OverrideSiteURL: "https://something-new.io"},
			expResp: generate.GenerateResp{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			m := mocks{
				mstg: storagemock.NewStatusPageSettingsGetter(t),
				msg:  storagemock.NewSystemGetter(t),
				mig:  storagemock.NewIncidentReportGetter(t),
				muc:  storagemock.NewUICreator(t),
				mpc:  storagemock.NewPromMetricsCreator(t),
			}

			test.mock(m)

			// Exec.
			svc, err := generate.NewService(generate.ServiceConfig{
				SettingsGetter:     m.mstg,
				SystemGetter:       m.msg,
				IRGetter:           m.mig,
				UICreator:          m.muc,
				PromMetricsCreator: m.mpc,
				Logger:             log.Noop,
			})
			require.NoError(err)

			resp, err := svc.Generate(context.TODO(), test.req)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expResp, resp)
			m.mstg.AssertExpectations(t)
			m.msg.AssertExpectations(t)
			m.mig.AssertExpectations(t)
			m.muc.AssertExpectations(t)
			m.mpc.AssertExpectations(t)
		})
	}
}
