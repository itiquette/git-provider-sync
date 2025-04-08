// SPDX-FileCopyrightText: 2025 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Code generated by mockery v2.46.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// BranchManager is an autogenerated mock type for the BranchManager type
type BranchManager struct {
	mock.Mock
}

type BranchManager_Expecter struct {
	mock *mock.Mock
}

func (_m *BranchManager) EXPECT() *BranchManager_Expecter {
	return &BranchManager_Expecter{mock: &_m.Mock}
}

// CreateTrackingBranches provides a mock function with given fields: ctx, repoPath
func (_m *BranchManager) CreateTrackingBranches(ctx context.Context, repoPath string) error {
	ret := _m.Called(ctx, repoPath)

	if len(ret) == 0 {
		panic("no return value specified for CreateTrackingBranches")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, repoPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BranchManager_CreateTrackingBranches_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateTrackingBranches'
type BranchManager_CreateTrackingBranches_Call struct {
	*mock.Call
}

// CreateTrackingBranches is a helper method to define mock.On call
//   - ctx context.Context
//   - repoPath string
func (_e *BranchManager_Expecter) CreateTrackingBranches(ctx interface{}, repoPath interface{}) *BranchManager_CreateTrackingBranches_Call {
	return &BranchManager_CreateTrackingBranches_Call{Call: _e.mock.On("CreateTrackingBranches", ctx, repoPath)}
}

func (_c *BranchManager_CreateTrackingBranches_Call) Run(run func(ctx context.Context, repoPath string)) *BranchManager_CreateTrackingBranches_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *BranchManager_CreateTrackingBranches_Call) Return(_a0 error) *BranchManager_CreateTrackingBranches_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BranchManager_CreateTrackingBranches_Call) RunAndReturn(run func(context.Context, string) error) *BranchManager_CreateTrackingBranches_Call {
	_c.Call.Return(run)
	return _c
}

// Fetch provides a mock function with given fields: ctx, workingDirPath
func (_m *BranchManager) Fetch(ctx context.Context, workingDirPath string) error {
	ret := _m.Called(ctx, workingDirPath)

	if len(ret) == 0 {
		panic("no return value specified for Fetch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, workingDirPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BranchManager_Fetch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Fetch'
type BranchManager_Fetch_Call struct {
	*mock.Call
}

// Fetch is a helper method to define mock.On call
//   - ctx context.Context
//   - workingDirPath string
func (_e *BranchManager_Expecter) Fetch(ctx interface{}, workingDirPath interface{}) *BranchManager_Fetch_Call {
	return &BranchManager_Fetch_Call{Call: _e.mock.On("Fetch", ctx, workingDirPath)}
}

func (_c *BranchManager_Fetch_Call) Run(run func(ctx context.Context, workingDirPath string)) *BranchManager_Fetch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *BranchManager_Fetch_Call) Return(_a0 error) *BranchManager_Fetch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BranchManager_Fetch_Call) RunAndReturn(run func(context.Context, string) error) *BranchManager_Fetch_Call {
	_c.Call.Return(run)
	return _c
}

// ProcessTrackingBranches provides a mock function with given fields: ctx, targetPath, input
func (_m *BranchManager) ProcessTrackingBranches(ctx context.Context, targetPath string, input []byte) error {
	ret := _m.Called(ctx, targetPath, input)

	if len(ret) == 0 {
		panic("no return value specified for ProcessTrackingBranches")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) error); ok {
		r0 = rf(ctx, targetPath, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BranchManager_ProcessTrackingBranches_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProcessTrackingBranches'
type BranchManager_ProcessTrackingBranches_Call struct {
	*mock.Call
}

// ProcessTrackingBranches is a helper method to define mock.On call
//   - ctx context.Context
//   - targetPath string
//   - input []byte
func (_e *BranchManager_Expecter) ProcessTrackingBranches(ctx interface{}, targetPath interface{}, input interface{}) *BranchManager_ProcessTrackingBranches_Call {
	return &BranchManager_ProcessTrackingBranches_Call{Call: _e.mock.On("ProcessTrackingBranches", ctx, targetPath, input)}
}

func (_c *BranchManager_ProcessTrackingBranches_Call) Run(run func(ctx context.Context, targetPath string, input []byte)) *BranchManager_ProcessTrackingBranches_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]byte))
	})
	return _c
}

func (_c *BranchManager_ProcessTrackingBranches_Call) Return(_a0 error) *BranchManager_ProcessTrackingBranches_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BranchManager_ProcessTrackingBranches_Call) RunAndReturn(run func(context.Context, string, []byte) error) *BranchManager_ProcessTrackingBranches_Call {
	_c.Call.Return(run)
	return _c
}

// NewBranchManager creates a new instance of BranchManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBranchManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *BranchManager {
	mock := &BranchManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
