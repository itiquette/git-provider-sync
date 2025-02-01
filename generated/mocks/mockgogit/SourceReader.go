// SPDX-FileCopyrightText: 2025 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "itiquette/git-provider-sync/internal/model"
)

// SourceReader is an autogenerated mock type for the SourceReader type
type SourceReader struct {
	mock.Mock
}

type SourceReader_Expecter struct {
	mock *mock.Mock
}

func (_m *SourceReader) EXPECT() *SourceReader_Expecter {
	return &SourceReader_Expecter{mock: &_m.Mock}
}

// Clone provides a mock function with given fields: ctx, option
func (_m *SourceReader) Clone(ctx context.Context, option model.CloneOption) (model.Repository, error) {
	ret := _m.Called(ctx, option)

	if len(ret) == 0 {
		panic("no return value specified for Clone")
	}

	var r0 model.Repository
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.CloneOption) (model.Repository, error)); ok {
		return rf(ctx, option)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.CloneOption) model.Repository); ok {
		r0 = rf(ctx, option)
	} else {
		r0 = ret.Get(0).(model.Repository)
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.CloneOption) error); ok {
		r1 = rf(ctx, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SourceReader_Clone_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Clone'
type SourceReader_Clone_Call struct {
	*mock.Call
}

// Clone is a helper method to define mock.On call
//   - ctx context.Context
//   - option model.CloneOption
func (_e *SourceReader_Expecter) Clone(ctx interface{}, option interface{}) *SourceReader_Clone_Call {
	return &SourceReader_Clone_Call{Call: _e.mock.On("Clone", ctx, option)}
}

func (_c *SourceReader_Clone_Call) Run(run func(ctx context.Context, option model.CloneOption)) *SourceReader_Clone_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.CloneOption))
	})
	return _c
}

func (_c *SourceReader_Clone_Call) Return(_a0 model.Repository, _a1 error) *SourceReader_Clone_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SourceReader_Clone_Call) RunAndReturn(run func(context.Context, model.CloneOption) (model.Repository, error)) *SourceReader_Clone_Call {
	_c.Call.Return(run)
	return _c
}

// NewSourceReader creates a new instance of SourceReader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSourceReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *SourceReader {
	mock := &SourceReader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
