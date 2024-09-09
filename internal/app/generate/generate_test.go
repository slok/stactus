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
	t0 := time.Now()

	tests := map[string]struct {
		mock    func(msg *storagemock.SystemGetter, mig *storagemock.IncidentReportGetter, muc *storagemock.UICreator)
		req     generate.GenerateReq
		expResp generate.GenerateResp
		expErr  bool
	}{

		"If listing systems returns an error, it should fail.": {
			mock: func(msg *storagemock.SystemGetter, mig *storagemock.IncidentReportGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return(nil, fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"If listing IRs returns an error, it should fail.": {
			mock: func(msg *storagemock.SystemGetter, mig *storagemock.IncidentReportGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{}, nil)
				mig.On("ListAllIncidentReports", mock.Anything).Return(nil, fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"If UI generation returns an error, it should fail.": {
			mock: func(msg *storagemock.SystemGetter, mig *storagemock.IncidentReportGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{}, nil)
				mig.On("ListAllIncidentReports", mock.Anything).Return([]model.IncidentReport{}, nil)
				muc.On("CreateUI", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"Creating the UI correctly should generate the UI (No IRs).": {
			mock: func(msg *storagemock.SystemGetter, mig *storagemock.IncidentReportGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{
					{ID: "test1", Name: "Test 1", Description: "Something 1"},
					{ID: "test2", Name: "Test 2", Description: "Something 2"},
					{ID: "test3", Name: "Test 3", Description: "Something 3"},
				}, nil)

				mig.On("ListAllIncidentReports", mock.Anything).Return([]model.IncidentReport{}, nil)

				exp := model.UI{
					History: []*model.IncidentReport{},
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
				muc.On("CreateUI", mock.Anything, exp).Once().Return(nil)
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
		},

		"Creating the UI correctly should generate the UI (service with IRs).": {
			mock: func(msg *storagemock.SystemGetter, mig *storagemock.IncidentReportGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{
					{ID: "test1", Name: "Test 1", Description: "Something 1"},
					{ID: "test2", Name: "Test 2", Description: "Something 2"},
					{ID: "test3", Name: "Test 3", Description: "Something 3"},
				}, nil)

				mig.On("ListAllIncidentReports", mock.Anything).Return([]model.IncidentReport{
					{
						ID:       "ir1",
						SystemID: "test2",
						Name:     "IR 1",
						Start:    t0,
						Details: []model.IncidentReportDetail{
							{Description: "desc1"},
						},
					},
					{
						ID:       "ir3",
						SystemID: "test3",
						Name:     "IR 3",
						Start:    t0.Add(-3 * time.Hour),
						End:      t0.Add(-2 * time.Hour),
					},
					{
						ID:       "ir2",
						SystemID: "test2",
						Name:     "IR 2",
						Start:    t0.Add(-5 * time.Hour),
						End:      t0.Add(-4 * time.Hour),
					},
				}, nil)

				exp := model.UI{
					LatestUpdate: &model.IncidentReportDetail{
						Description: "desc1",
					},
					History: []*model.IncidentReport{
						{ID: "ir1", SystemID: "test2", Name: "IR 1", Start: t0, Details: []model.IncidentReportDetail{{Description: "desc1"}}},
						{ID: "ir3", SystemID: "test3", Name: "IR 3", Start: t0.Add(-3 * time.Hour), End: t0.Add(-2 * time.Hour)},
						{ID: "ir2", SystemID: "test2", Name: "IR 2", Start: t0.Add(-5 * time.Hour), End: t0.Add(-4 * time.Hour)},
					},
					SystemDetails: []model.SystemDetails{
						{
							System: model.System{ID: "test1", Name: "Test 1", Description: "Something 1"},
						},
						{
							System:   model.System{ID: "test2", Name: "Test 2", Description: "Something 2"},
							LatestIR: &model.IncidentReport{ID: "ir1", SystemID: "test2", Name: "IR 1", Start: t0, Details: []model.IncidentReportDetail{{Description: "desc1"}}},
							IRs: []*model.IncidentReport{
								{ID: "ir1", SystemID: "test2", Name: "IR 1", Start: t0, Details: []model.IncidentReportDetail{{Description: "desc1"}}},
								{ID: "ir2", SystemID: "test2", Name: "IR 2", Start: t0.Add(-5 * time.Hour), End: t0.Add(-4 * time.Hour)},
							},
						},
						{
							System:   model.System{ID: "test3", Name: "Test 3", Description: "Something 3"},
							LatestIR: &model.IncidentReport{ID: "ir3", SystemID: "test3", Name: "IR 3", Start: t0.Add(-3 * time.Hour), End: t0.Add(-2 * time.Hour)},
							IRs: []*model.IncidentReport{
								{ID: "ir3", SystemID: "test3", Name: "IR 3", Start: t0.Add(-3 * time.Hour), End: t0.Add(-2 * time.Hour)},
							},
						},
					},
				}
				muc.On("CreateUI", mock.Anything, exp).Once().Return(nil)
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			msg := storagemock.NewSystemGetter(t)
			mig := storagemock.NewIncidentReportGetter(t)
			muc := storagemock.NewUICreator(t)
			test.mock(msg, mig, muc)

			// Exec.
			svc, err := generate.NewService(generate.ServiceConfig{
				SystemGetter: msg,
				IRGetter:     mig,
				UICreator:    muc,
				Logger:       log.Noop,
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
			msg.AssertExpectations(t)
			mig.AssertExpectations(t)
			muc.AssertExpectations(t)
		})
	}
}
