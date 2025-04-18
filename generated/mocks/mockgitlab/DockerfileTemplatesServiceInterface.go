// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// DockerfileTemplatesServiceInterface is an autogenerated mock type for the DockerfileTemplatesServiceInterface type
type DockerfileTemplatesServiceInterface struct {
	mock.Mock
}

type DockerfileTemplatesServiceInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *DockerfileTemplatesServiceInterface) EXPECT() *DockerfileTemplatesServiceInterface_Expecter {
	return &DockerfileTemplatesServiceInterface_Expecter{mock: &_m.Mock}
}

// GetTemplate provides a mock function with given fields: key, options
func (_m *DockerfileTemplatesServiceInterface) GetTemplate(key string, options ...gitlab.RequestOptionFunc) (*gitlab.DockerfileTemplate, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, key)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetTemplate")
	}

	var r0 *gitlab.DockerfileTemplate
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(string, ...gitlab.RequestOptionFunc) (*gitlab.DockerfileTemplate, *gitlab.Response, error)); ok {
		return rf(key, options...)
	}
	if rf, ok := ret.Get(0).(func(string, ...gitlab.RequestOptionFunc) *gitlab.DockerfileTemplate); ok {
		r0 = rf(key, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.DockerfileTemplate)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(key, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(string, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(key, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DockerfileTemplatesServiceInterface_GetTemplate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTemplate'
type DockerfileTemplatesServiceInterface_GetTemplate_Call struct {
	*mock.Call
}

// GetTemplate is a helper method to define mock.On call
//   - key string
//   - options ...gitlab.RequestOptionFunc
func (_e *DockerfileTemplatesServiceInterface_Expecter) GetTemplate(key interface{}, options ...interface{}) *DockerfileTemplatesServiceInterface_GetTemplate_Call {
	return &DockerfileTemplatesServiceInterface_GetTemplate_Call{Call: _e.mock.On("GetTemplate",
		append([]interface{}{key}, options...)...)}
}

func (_c *DockerfileTemplatesServiceInterface_GetTemplate_Call) Run(run func(key string, options ...gitlab.RequestOptionFunc)) *DockerfileTemplatesServiceInterface_GetTemplate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *DockerfileTemplatesServiceInterface_GetTemplate_Call) Return(_a0 *gitlab.DockerfileTemplate, _a1 *gitlab.Response, _a2 error) *DockerfileTemplatesServiceInterface_GetTemplate_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *DockerfileTemplatesServiceInterface_GetTemplate_Call) RunAndReturn(run func(string, ...gitlab.RequestOptionFunc) (*gitlab.DockerfileTemplate, *gitlab.Response, error)) *DockerfileTemplatesServiceInterface_GetTemplate_Call {
	_c.Call.Return(run)
	return _c
}

// ListTemplates provides a mock function with given fields: opt, options
func (_m *DockerfileTemplatesServiceInterface) ListTemplates(opt *gitlab.ListDockerfileTemplatesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DockerfileTemplateListItem, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListTemplates")
	}

	var r0 []*gitlab.DockerfileTemplateListItem
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(*gitlab.ListDockerfileTemplatesOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.DockerfileTemplateListItem, *gitlab.Response, error)); ok {
		return rf(opt, options...)
	}
	if rf, ok := ret.Get(0).(func(*gitlab.ListDockerfileTemplatesOptions, ...gitlab.RequestOptionFunc) []*gitlab.DockerfileTemplateListItem); ok {
		r0 = rf(opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.DockerfileTemplateListItem)
		}
	}

	if rf, ok := ret.Get(1).(func(*gitlab.ListDockerfileTemplatesOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(*gitlab.ListDockerfileTemplatesOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DockerfileTemplatesServiceInterface_ListTemplates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListTemplates'
type DockerfileTemplatesServiceInterface_ListTemplates_Call struct {
	*mock.Call
}

// ListTemplates is a helper method to define mock.On call
//   - opt *gitlab.ListDockerfileTemplatesOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *DockerfileTemplatesServiceInterface_Expecter) ListTemplates(opt interface{}, options ...interface{}) *DockerfileTemplatesServiceInterface_ListTemplates_Call {
	return &DockerfileTemplatesServiceInterface_ListTemplates_Call{Call: _e.mock.On("ListTemplates",
		append([]interface{}{opt}, options...)...)}
}

func (_c *DockerfileTemplatesServiceInterface_ListTemplates_Call) Run(run func(opt *gitlab.ListDockerfileTemplatesOptions, options ...gitlab.RequestOptionFunc)) *DockerfileTemplatesServiceInterface_ListTemplates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(*gitlab.ListDockerfileTemplatesOptions), variadicArgs...)
	})
	return _c
}

func (_c *DockerfileTemplatesServiceInterface_ListTemplates_Call) Return(_a0 []*gitlab.DockerfileTemplateListItem, _a1 *gitlab.Response, _a2 error) *DockerfileTemplatesServiceInterface_ListTemplates_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *DockerfileTemplatesServiceInterface_ListTemplates_Call) RunAndReturn(run func(*gitlab.ListDockerfileTemplatesOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.DockerfileTemplateListItem, *gitlab.Response, error)) *DockerfileTemplatesServiceInterface_ListTemplates_Call {
	_c.Call.Return(run)
	return _c
}

// NewDockerfileTemplatesServiceInterface creates a new instance of DockerfileTemplatesServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDockerfileTemplatesServiceInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *DockerfileTemplatesServiceInterface {
	mock := &DockerfileTemplatesServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
