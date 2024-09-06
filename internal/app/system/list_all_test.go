package system_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slok/stactus/internal/app/system"
	"github.com/slok/stactus/internal/log"
	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/storagemock"
)

func TestListAllSystems(t *testing.T) {
	tests := map[string]struct {
		mock    func(msg *storagemock.SystemGetter)
		req     system.ListAllSystemsReq
		expResp system.ListAllSystemsResp
		expErr  bool
	}{

		"If listing systems returns an error, it should fail.": {
			mock: func(msg *storagemock.SystemGetter) {
				msg.On("ListAllSystems", mock.Anything).Once().Return(nil, fmt.Errorf("something"))
			},
			req:     system.ListAllSystemsReq{},
			expResp: system.ListAllSystemsResp{},
			expErr:  true,
		},

		"Listing systems correctly should return the systems.": {
			mock: func(msg *storagemock.SystemGetter) {
				msg.On("ListAllSystems", mock.Anything).Once().Return([]model.System{
					{ID: "test1", Name: "Test 1", Description: "Something 1"},
					{ID: "test2", Name: "Test 2", Description: "Something 2"},
					{ID: "test3", Name: "Test 3", Description: "Something 3"},
				}, nil)
			},
			req: system.ListAllSystemsReq{},
			expResp: system.ListAllSystemsResp{
				Systems: []system.SystemAndOngoingIncident{
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
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			// Mocks.
			msg := storagemock.NewSystemGetter(t)
			test.mock(msg)

			// Exec.
			svc, err := system.NewService(system.ServiceConfig{
				SystemGetter: msg,
				Logger:       log.Noop,
			})
			require.NoError(err)

			resp, err := svc.ListAllSystems(context.TODO(), test.req)

			// Check.
			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expResp, resp)
			msg.AssertExpectations(t)
		})
	}
}
