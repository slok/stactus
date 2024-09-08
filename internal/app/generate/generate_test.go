package generate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/stactus/internal/app/generate"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/storagemock"
)

func TestGenerate(t *testing.T) {
	tests := map[string]struct {
		mock    func(msg *storagemock.SystemGetter, muc *storagemock.UICreator)
		req     generate.GenerateReq
		expResp generate.GenerateResp
		expErr  bool
	}{

		"If listing systems returns an error, it should fail.": {
			mock: func(msg *storagemock.SystemGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return(nil, fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"If UI generation returns an error, it should fail.": {
			mock: func(msg *storagemock.SystemGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{}, nil)
				muc.On("CreateUI", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("something"))
			},
			req:     generate.GenerateReq{},
			expResp: generate.GenerateResp{},
			expErr:  true,
		},

		"Generating correct UI should generate the UI without errors.": {
			mock: func(msg *storagemock.SystemGetter, muc *storagemock.UICreator) {
				msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{
					{ID: "test1", Name: "Test 1", Description: "Something 1"},
					{ID: "test2", Name: "Test 2", Description: "Something 2"},
					{ID: "test3", Name: "Test 3", Description: "Something 3"},
				}, nil)

				exp := model.UI{
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			msg := storagemock.NewSystemGetter(t)
			muc := storagemock.NewUICreator(t)
			test.mock(msg, muc)

			// Exec.
			svc, err := generate.NewService(generate.ServiceConfig{
				SystemGetter: msg,
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
			muc.AssertExpectations(t)
		})
	}
}
