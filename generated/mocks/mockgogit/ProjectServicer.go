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

// ProjectServicer is an autogenerated mock type for the ProjectServicer type
type ProjectServicer struct {
	mock.Mock
}

type ProjectServicer_Expecter struct {
	mock *mock.Mock
}

func (_m *ProjectServicer) EXPECT() *ProjectServicer_Expecter {
	return &ProjectServicer_Expecter{mock: &_m.Mock}
}

// CreateProject provides a mock function with given fields: ctx, opt
func (_m *ProjectServicer) CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
	ret := _m.Called(ctx, opt)

	if len(ret) == 0 {
		panic("no return value specified for CreateProject")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.CreateProjectOption) (string, error)); ok {
		return rf(ctx, opt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.CreateProjectOption) string); ok {
		r0 = rf(ctx, opt)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.CreateProjectOption) error); ok {
		r1 = rf(ctx, opt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectServicer_CreateProject_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateProject'
type ProjectServicer_CreateProject_Call struct {
	*mock.Call
}

// CreateProject is a helper method to define mock.On call
//   - ctx context.Context
//   - opt model.CreateProjectOption
func (_e *ProjectServicer_Expecter) CreateProject(ctx interface{}, opt interface{}) *ProjectServicer_CreateProject_Call {
	return &ProjectServicer_CreateProject_Call{Call: _e.mock.On("CreateProject", ctx, opt)}
}

func (_c *ProjectServicer_CreateProject_Call) Run(run func(ctx context.Context, opt model.CreateProjectOption)) *ProjectServicer_CreateProject_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.CreateProjectOption))
	})
	return _c
}

func (_c *ProjectServicer_CreateProject_Call) Return(_a0 string, _a1 error) *ProjectServicer_CreateProject_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectServicer_CreateProject_Call) RunAndReturn(run func(context.Context, model.CreateProjectOption) (string, error)) *ProjectServicer_CreateProject_Call {
	_c.Call.Return(run)
	return _c
}

// GetProjectInfos provides a mock function with given fields: ctx, providerOpt, filtering
func (_m *ProjectServicer) GetProjectInfos(ctx context.Context, providerOpt model.ProviderOption, filtering bool) ([]model.ProjectInfo, error) {
	ret := _m.Called(ctx, providerOpt, filtering)

	if len(ret) == 0 {
		panic("no return value specified for GetProjectInfos")
	}

	var r0 []model.ProjectInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.ProviderOption, bool) ([]model.ProjectInfo, error)); ok {
		return rf(ctx, providerOpt, filtering)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.ProviderOption, bool) []model.ProjectInfo); ok {
		r0 = rf(ctx, providerOpt, filtering)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.ProjectInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.ProviderOption, bool) error); ok {
		r1 = rf(ctx, providerOpt, filtering)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectServicer_GetProjectInfos_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProjectInfos'
type ProjectServicer_GetProjectInfos_Call struct {
	*mock.Call
}

// GetProjectInfos is a helper method to define mock.On call
//   - ctx context.Context
//   - providerOpt model.ProviderOption
//   - filtering bool
func (_e *ProjectServicer_Expecter) GetProjectInfos(ctx interface{}, providerOpt interface{}, filtering interface{}) *ProjectServicer_GetProjectInfos_Call {
	return &ProjectServicer_GetProjectInfos_Call{Call: _e.mock.On("GetProjectInfos", ctx, providerOpt, filtering)}
}

func (_c *ProjectServicer_GetProjectInfos_Call) Run(run func(ctx context.Context, providerOpt model.ProviderOption, filtering bool)) *ProjectServicer_GetProjectInfos_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.ProviderOption), args[2].(bool))
	})
	return _c
}

func (_c *ProjectServicer_GetProjectInfos_Call) Return(_a0 []model.ProjectInfo, _a1 error) *ProjectServicer_GetProjectInfos_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectServicer_GetProjectInfos_Call) RunAndReturn(run func(context.Context, model.ProviderOption, bool) ([]model.ProjectInfo, error)) *ProjectServicer_GetProjectInfos_Call {
	_c.Call.Return(run)
	return _c
}

// ProjectExists provides a mock function with given fields: ctx, owner, repo
func (_m *ProjectServicer) ProjectExists(ctx context.Context, owner string, repo string) (bool, string, error) {
	ret := _m.Called(ctx, owner, repo)

	if len(ret) == 0 {
		panic("no return value specified for ProjectExists")
	}

	var r0 bool
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, string, error)); ok {
		return rf(ctx, owner, repo)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, owner, repo)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) string); ok {
		r1 = rf(ctx, owner, repo)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, string) error); ok {
		r2 = rf(ctx, owner, repo)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ProjectServicer_ProjectExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProjectExists'
type ProjectServicer_ProjectExists_Call struct {
	*mock.Call
}

// ProjectExists is a helper method to define mock.On call
//   - ctx context.Context
//   - owner string
//   - repo string
func (_e *ProjectServicer_Expecter) ProjectExists(ctx interface{}, owner interface{}, repo interface{}) *ProjectServicer_ProjectExists_Call {
	return &ProjectServicer_ProjectExists_Call{Call: _e.mock.On("ProjectExists", ctx, owner, repo)}
}

func (_c *ProjectServicer_ProjectExists_Call) Run(run func(ctx context.Context, owner string, repo string)) *ProjectServicer_ProjectExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *ProjectServicer_ProjectExists_Call) Return(_a0 bool, _a1 string, _a2 error) *ProjectServicer_ProjectExists_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *ProjectServicer_ProjectExists_Call) RunAndReturn(run func(context.Context, string, string) (bool, string, error)) *ProjectServicer_ProjectExists_Call {
	_c.Call.Return(run)
	return _c
}

// SetDefaultBranch provides a mock function with given fields: ctx, owner, projectName, branch
func (_m *ProjectServicer) SetDefaultBranch(ctx context.Context, owner string, projectName string, branch string) error {
	ret := _m.Called(ctx, owner, projectName, branch)

	if len(ret) == 0 {
		panic("no return value specified for SetDefaultBranch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, owner, projectName, branch)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ProjectServicer_SetDefaultBranch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetDefaultBranch'
type ProjectServicer_SetDefaultBranch_Call struct {
	*mock.Call
}

// SetDefaultBranch is a helper method to define mock.On call
//   - ctx context.Context
//   - owner string
//   - projectName string
//   - branch string
func (_e *ProjectServicer_Expecter) SetDefaultBranch(ctx interface{}, owner interface{}, projectName interface{}, branch interface{}) *ProjectServicer_SetDefaultBranch_Call {
	return &ProjectServicer_SetDefaultBranch_Call{Call: _e.mock.On("SetDefaultBranch", ctx, owner, projectName, branch)}
}

func (_c *ProjectServicer_SetDefaultBranch_Call) Run(run func(ctx context.Context, owner string, projectName string, branch string)) *ProjectServicer_SetDefaultBranch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *ProjectServicer_SetDefaultBranch_Call) Return(_a0 error) *ProjectServicer_SetDefaultBranch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ProjectServicer_SetDefaultBranch_Call) RunAndReturn(run func(context.Context, string, string, string) error) *ProjectServicer_SetDefaultBranch_Call {
	_c.Call.Return(run)
	return _c
}

// NewProjectServicer creates a new instance of ProjectServicer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProjectServicer(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProjectServicer {
	mock := &ProjectServicer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
