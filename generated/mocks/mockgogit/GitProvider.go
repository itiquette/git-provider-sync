// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "itiquette/git-provider-sync/internal/model"
)

// GitProvider is an autogenerated mock type for the GitProvider type
type GitProvider struct {
	mock.Mock
}

type GitProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *GitProvider) EXPECT() *GitProvider_Expecter {
	return &GitProvider_Expecter{mock: &_m.Mock}
}

// CreateProject provides a mock function with given fields: ctx, opt
func (_m *GitProvider) CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
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

// GitProvider_CreateProject_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateProject'
type GitProvider_CreateProject_Call struct {
	*mock.Call
}

// CreateProject is a helper method to define mock.On call
//   - ctx context.Context
//   - opt model.CreateProjectOption
func (_e *GitProvider_Expecter) CreateProject(ctx interface{}, opt interface{}) *GitProvider_CreateProject_Call {
	return &GitProvider_CreateProject_Call{Call: _e.mock.On("CreateProject", ctx, opt)}
}

func (_c *GitProvider_CreateProject_Call) Run(run func(ctx context.Context, opt model.CreateProjectOption)) *GitProvider_CreateProject_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.CreateProjectOption))
	})
	return _c
}

func (_c *GitProvider_CreateProject_Call) Return(_a0 string, _a1 error) *GitProvider_CreateProject_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GitProvider_CreateProject_Call) RunAndReturn(run func(context.Context, model.CreateProjectOption) (string, error)) *GitProvider_CreateProject_Call {
	_c.Call.Return(run)
	return _c
}

// GetProjectInfos provides a mock function with given fields: ctx, providerOpt, filtering
func (_m *GitProvider) GetProjectInfos(ctx context.Context, providerOpt model.ProviderOption, filtering bool) ([]model.ProjectInfo, error) {
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

// GitProvider_GetProjectInfos_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProjectInfos'
type GitProvider_GetProjectInfos_Call struct {
	*mock.Call
}

// GetProjectInfos is a helper method to define mock.On call
//   - ctx context.Context
//   - providerOpt model.ProviderOption
//   - filtering bool
func (_e *GitProvider_Expecter) GetProjectInfos(ctx interface{}, providerOpt interface{}, filtering interface{}) *GitProvider_GetProjectInfos_Call {
	return &GitProvider_GetProjectInfos_Call{Call: _e.mock.On("GetProjectInfos", ctx, providerOpt, filtering)}
}

func (_c *GitProvider_GetProjectInfos_Call) Run(run func(ctx context.Context, providerOpt model.ProviderOption, filtering bool)) *GitProvider_GetProjectInfos_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(model.ProviderOption), args[2].(bool))
	})
	return _c
}

func (_c *GitProvider_GetProjectInfos_Call) Return(_a0 []model.ProjectInfo, _a1 error) *GitProvider_GetProjectInfos_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GitProvider_GetProjectInfos_Call) RunAndReturn(run func(context.Context, model.ProviderOption, bool) ([]model.ProjectInfo, error)) *GitProvider_GetProjectInfos_Call {
	_c.Call.Return(run)
	return _c
}

// IsValidProjectName provides a mock function with given fields: ctx, name
func (_m *GitProvider) IsValidProjectName(ctx context.Context, name string) bool {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for IsValidProjectName")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GitProvider_IsValidProjectName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsValidProjectName'
type GitProvider_IsValidProjectName_Call struct {
	*mock.Call
}

// IsValidProjectName is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *GitProvider_Expecter) IsValidProjectName(ctx interface{}, name interface{}) *GitProvider_IsValidProjectName_Call {
	return &GitProvider_IsValidProjectName_Call{Call: _e.mock.On("IsValidProjectName", ctx, name)}
}

func (_c *GitProvider_IsValidProjectName_Call) Run(run func(ctx context.Context, name string)) *GitProvider_IsValidProjectName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *GitProvider_IsValidProjectName_Call) Return(_a0 bool) *GitProvider_IsValidProjectName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GitProvider_IsValidProjectName_Call) RunAndReturn(run func(context.Context, string) bool) *GitProvider_IsValidProjectName_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with no fields
func (_m *GitProvider) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GitProvider_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type GitProvider_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *GitProvider_Expecter) Name() *GitProvider_Name_Call {
	return &GitProvider_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *GitProvider_Name_Call) Run(run func()) *GitProvider_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GitProvider_Name_Call) Return(_a0 string) *GitProvider_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GitProvider_Name_Call) RunAndReturn(run func() string) *GitProvider_Name_Call {
	_c.Call.Return(run)
	return _c
}

// ProjectExists provides a mock function with given fields: ctx, owner, repo
func (_m *GitProvider) ProjectExists(ctx context.Context, owner string, repo string) (bool, string, error) {
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

// GitProvider_ProjectExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProjectExists'
type GitProvider_ProjectExists_Call struct {
	*mock.Call
}

// ProjectExists is a helper method to define mock.On call
//   - ctx context.Context
//   - owner string
//   - repo string
func (_e *GitProvider_Expecter) ProjectExists(ctx interface{}, owner interface{}, repo interface{}) *GitProvider_ProjectExists_Call {
	return &GitProvider_ProjectExists_Call{Call: _e.mock.On("ProjectExists", ctx, owner, repo)}
}

func (_c *GitProvider_ProjectExists_Call) Run(run func(ctx context.Context, owner string, repo string)) *GitProvider_ProjectExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *GitProvider_ProjectExists_Call) Return(_a0 bool, _a1 string, _a2 error) *GitProvider_ProjectExists_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *GitProvider_ProjectExists_Call) RunAndReturn(run func(context.Context, string, string) (bool, string, error)) *GitProvider_ProjectExists_Call {
	_c.Call.Return(run)
	return _c
}

// Protect provides a mock function with given fields: ctx, owner, defaultBranch, projectIDstr
func (_m *GitProvider) Protect(ctx context.Context, owner string, defaultBranch string, projectIDstr string) error {
	ret := _m.Called(ctx, owner, defaultBranch, projectIDstr)

	if len(ret) == 0 {
		panic("no return value specified for Protect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, owner, defaultBranch, projectIDstr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GitProvider_Protect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Protect'
type GitProvider_Protect_Call struct {
	*mock.Call
}

// Protect is a helper method to define mock.On call
//   - ctx context.Context
//   - owner string
//   - defaultBranch string
//   - projectIDstr string
func (_e *GitProvider_Expecter) Protect(ctx interface{}, owner interface{}, defaultBranch interface{}, projectIDstr interface{}) *GitProvider_Protect_Call {
	return &GitProvider_Protect_Call{Call: _e.mock.On("Protect", ctx, owner, defaultBranch, projectIDstr)}
}

func (_c *GitProvider_Protect_Call) Run(run func(ctx context.Context, owner string, defaultBranch string, projectIDstr string)) *GitProvider_Protect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *GitProvider_Protect_Call) Return(_a0 error) *GitProvider_Protect_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GitProvider_Protect_Call) RunAndReturn(run func(context.Context, string, string, string) error) *GitProvider_Protect_Call {
	_c.Call.Return(run)
	return _c
}

// SetDefaultBranch provides a mock function with given fields: ctx, owner, name, branch
func (_m *GitProvider) SetDefaultBranch(ctx context.Context, owner string, name string, branch string) error {
	ret := _m.Called(ctx, owner, name, branch)

	if len(ret) == 0 {
		panic("no return value specified for SetDefaultBranch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, owner, name, branch)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GitProvider_SetDefaultBranch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetDefaultBranch'
type GitProvider_SetDefaultBranch_Call struct {
	*mock.Call
}

// SetDefaultBranch is a helper method to define mock.On call
//   - ctx context.Context
//   - owner string
//   - name string
//   - branch string
func (_e *GitProvider_Expecter) SetDefaultBranch(ctx interface{}, owner interface{}, name interface{}, branch interface{}) *GitProvider_SetDefaultBranch_Call {
	return &GitProvider_SetDefaultBranch_Call{Call: _e.mock.On("SetDefaultBranch", ctx, owner, name, branch)}
}

func (_c *GitProvider_SetDefaultBranch_Call) Run(run func(ctx context.Context, owner string, name string, branch string)) *GitProvider_SetDefaultBranch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *GitProvider_SetDefaultBranch_Call) Return(_a0 error) *GitProvider_SetDefaultBranch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GitProvider_SetDefaultBranch_Call) RunAndReturn(run func(context.Context, string, string, string) error) *GitProvider_SetDefaultBranch_Call {
	_c.Call.Return(run)
	return _c
}

// Unprotect provides a mock function with given fields: ctx, defaultBranch, projectIDStr
func (_m *GitProvider) Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error {
	ret := _m.Called(ctx, defaultBranch, projectIDStr)

	if len(ret) == 0 {
		panic("no return value specified for Unprotect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, defaultBranch, projectIDStr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GitProvider_Unprotect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Unprotect'
type GitProvider_Unprotect_Call struct {
	*mock.Call
}

// Unprotect is a helper method to define mock.On call
//   - ctx context.Context
//   - defaultBranch string
//   - projectIDStr string
func (_e *GitProvider_Expecter) Unprotect(ctx interface{}, defaultBranch interface{}, projectIDStr interface{}) *GitProvider_Unprotect_Call {
	return &GitProvider_Unprotect_Call{Call: _e.mock.On("Unprotect", ctx, defaultBranch, projectIDStr)}
}

func (_c *GitProvider_Unprotect_Call) Run(run func(ctx context.Context, defaultBranch string, projectIDStr string)) *GitProvider_Unprotect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *GitProvider_Unprotect_Call) Return(_a0 error) *GitProvider_Unprotect_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GitProvider_Unprotect_Call) RunAndReturn(run func(context.Context, string, string) error) *GitProvider_Unprotect_Call {
	_c.Call.Return(run)
	return _c
}

// NewGitProvider creates a new instance of GitProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGitProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *GitProvider {
	mock := &GitProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
