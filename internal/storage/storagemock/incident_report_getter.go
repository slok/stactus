// Code generated by mockery v2.45.0. DO NOT EDIT.

package storagemock

import (
	context "context"

	model "github.com/slok/stactus/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// IncidentReportGetter is an autogenerated mock type for the IncidentReportGetter type
type IncidentReportGetter struct {
	mock.Mock
}

// ListAllIncidentReports provides a mock function with given fields: ctx
func (_m *IncidentReportGetter) ListAllIncidentReports(ctx context.Context) ([]model.IncidentReport, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListAllIncidentReports")
	}

	var r0 []model.IncidentReport
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]model.IncidentReport, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []model.IncidentReport); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.IncidentReport)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIncidentReportGetter creates a new instance of IncidentReportGetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIncidentReportGetter(t interface {
	mock.TestingT
	Cleanup(func())
}) *IncidentReportGetter {
	mock := &IncidentReportGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
