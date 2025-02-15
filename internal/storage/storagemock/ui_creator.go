// Code generated by mockery v2.45.0. DO NOT EDIT.

package storagemock

import (
	context "context"

	model "github.com/slok/stactus/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// UICreator is an autogenerated mock type for the UICreator type
type UICreator struct {
	mock.Mock
}

// CreateUI provides a mock function with given fields: ctx, ui
func (_m *UICreator) CreateUI(ctx context.Context, ui model.UI) error {
	ret := _m.Called(ctx, ui)

	if len(ret) == 0 {
		panic("no return value specified for CreateUI")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.UI) error); ok {
		r0 = rf(ctx, ui)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewUICreator creates a new instance of UICreator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUICreator(t interface {
	mock.TestingT
	Cleanup(func())
}) *UICreator {
	mock := &UICreator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
