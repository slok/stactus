package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/slok/stactus/internal/model"
	"github.com/slok/stactus/internal/storage/memory"
)

func TestRepositoryListAllSystems(t *testing.T) {
	tests := map[string]struct {
		repo       func() memory.Repository
		expSystems []model.System
		expErr     bool
	}{
		"Having no systems should return empty list of systems.": {
			repo: func() memory.Repository {
				r := memory.NewRepository(nil, nil)
				return r
			},
			expSystems: []model.System{},
		},

		"Having multiple systems should be returned ordered by name.": {
			repo: func() memory.Repository {
				r := memory.NewRepository([]model.System{
					{ID: "test2", Name: "Test 2", Description: "something 2"},
					{ID: "test3", Name: "Test 3", Description: "something 3"},
					{ID: "test1", Name: "Test 1", Description: "something 1"},
					{ID: "test4", Name: "Test 4", Description: "something 4"},
				}, nil)
				return r
			},
			expSystems: []model.System{
				{ID: "test1", Name: "Test 1", Description: "something 1"},
				{ID: "test2", Name: "Test 2", Description: "something 2"},
				{ID: "test3", Name: "Test 3", Description: "something 3"},
				{ID: "test4", Name: "Test 4", Description: "something 4"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotSystems, err := r.ListAllSystems(context.TODO())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expSystems, gotSystems)
			}
		})
	}
}

func TestRepositoryListAllIncidentReports(t *testing.T) {
	t0 := time.Now().UTC()

	tests := map[string]struct {
		repo   func() memory.Repository
		expIRs []model.IncidentReport
		expErr bool
	}{
		"Having no IRs should return empty list of IRs.": {
			repo: func() memory.Repository {
				r := memory.NewRepository(nil, nil)
				return r
			},
			expIRs: []model.IncidentReport{},
		},

		"Having multiple IRs should be returned ordered by created descending.": {
			repo: func() memory.Repository {
				r := memory.NewRepository(nil, []model.IncidentReport{
					{ID: "test2", Name: "Test 2", Start: t0.Add(242 * time.Minute)},
					{ID: "test1", Name: "Test 1", Start: t0.Add(342 * time.Minute)},
					{ID: "test4", Name: "Test 4", Start: t0.Add(42 * time.Minute)},
					{ID: "test3", Name: "Test 3", Start: t0.Add(142 * time.Minute)},
				})
				return r
			},
			expIRs: []model.IncidentReport{
				{ID: "test1", Name: "Test 1", Start: t0.Add(342 * time.Minute)},
				{ID: "test2", Name: "Test 2", Start: t0.Add(242 * time.Minute)},
				{ID: "test3", Name: "Test 3", Start: t0.Add(142 * time.Minute)},
				{ID: "test4", Name: "Test 4", Start: t0.Add(42 * time.Minute)},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			r := test.repo()
			gotIRs, err := r.ListAllIncidentReports(context.TODO())

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expIRs, gotIRs)
			}
		})
	}
}
