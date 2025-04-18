// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GroupWikisServiceInterface is an autogenerated mock type for the GroupWikisServiceInterface type
type GroupWikisServiceInterface struct {
	mock.Mock
}

type GroupWikisServiceInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *GroupWikisServiceInterface) EXPECT() *GroupWikisServiceInterface_Expecter {
	return &GroupWikisServiceInterface_Expecter{mock: &_m.Mock}
}

// CreateGroupWikiPage provides a mock function with given fields: gid, opt, options
func (_m *GroupWikisServiceInterface) CreateGroupWikiPage(gid interface{}, opt *gitlab.CreateGroupWikiPageOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, gid, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateGroupWikiPage")
	}

	var r0 *gitlab.GroupWiki
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.CreateGroupWikiPageOptions, ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error)); ok {
		return rf(gid, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.CreateGroupWikiPageOptions, ...gitlab.RequestOptionFunc) *gitlab.GroupWiki); ok {
		r0 = rf(gid, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.GroupWiki)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, *gitlab.CreateGroupWikiPageOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(gid, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, *gitlab.CreateGroupWikiPageOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(gid, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GroupWikisServiceInterface_CreateGroupWikiPage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateGroupWikiPage'
type GroupWikisServiceInterface_CreateGroupWikiPage_Call struct {
	*mock.Call
}

// CreateGroupWikiPage is a helper method to define mock.On call
//   - gid interface{}
//   - opt *gitlab.CreateGroupWikiPageOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupWikisServiceInterface_Expecter) CreateGroupWikiPage(gid interface{}, opt interface{}, options ...interface{}) *GroupWikisServiceInterface_CreateGroupWikiPage_Call {
	return &GroupWikisServiceInterface_CreateGroupWikiPage_Call{Call: _e.mock.On("CreateGroupWikiPage",
		append([]interface{}{gid, opt}, options...)...)}
}

func (_c *GroupWikisServiceInterface_CreateGroupWikiPage_Call) Run(run func(gid interface{}, opt *gitlab.CreateGroupWikiPageOptions, options ...gitlab.RequestOptionFunc)) *GroupWikisServiceInterface_CreateGroupWikiPage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(*gitlab.CreateGroupWikiPageOptions), variadicArgs...)
	})
	return _c
}

func (_c *GroupWikisServiceInterface_CreateGroupWikiPage_Call) Return(_a0 *gitlab.GroupWiki, _a1 *gitlab.Response, _a2 error) *GroupWikisServiceInterface_CreateGroupWikiPage_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *GroupWikisServiceInterface_CreateGroupWikiPage_Call) RunAndReturn(run func(interface{}, *gitlab.CreateGroupWikiPageOptions, ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error)) *GroupWikisServiceInterface_CreateGroupWikiPage_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteGroupWikiPage provides a mock function with given fields: gid, slug, options
func (_m *GroupWikisServiceInterface) DeleteGroupWikiPage(gid interface{}, slug string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, gid, slug)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteGroupWikiPage")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}, string, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(gid, slug, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, string, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(gid, slug, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, string, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(gid, slug, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GroupWikisServiceInterface_DeleteGroupWikiPage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteGroupWikiPage'
type GroupWikisServiceInterface_DeleteGroupWikiPage_Call struct {
	*mock.Call
}

// DeleteGroupWikiPage is a helper method to define mock.On call
//   - gid interface{}
//   - slug string
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupWikisServiceInterface_Expecter) DeleteGroupWikiPage(gid interface{}, slug interface{}, options ...interface{}) *GroupWikisServiceInterface_DeleteGroupWikiPage_Call {
	return &GroupWikisServiceInterface_DeleteGroupWikiPage_Call{Call: _e.mock.On("DeleteGroupWikiPage",
		append([]interface{}{gid, slug}, options...)...)}
}

func (_c *GroupWikisServiceInterface_DeleteGroupWikiPage_Call) Run(run func(gid interface{}, slug string, options ...gitlab.RequestOptionFunc)) *GroupWikisServiceInterface_DeleteGroupWikiPage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *GroupWikisServiceInterface_DeleteGroupWikiPage_Call) Return(_a0 *gitlab.Response, _a1 error) *GroupWikisServiceInterface_DeleteGroupWikiPage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GroupWikisServiceInterface_DeleteGroupWikiPage_Call) RunAndReturn(run func(interface{}, string, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *GroupWikisServiceInterface_DeleteGroupWikiPage_Call {
	_c.Call.Return(run)
	return _c
}

// EditGroupWikiPage provides a mock function with given fields: gid, slug, opt, options
func (_m *GroupWikisServiceInterface) EditGroupWikiPage(gid interface{}, slug string, opt *gitlab.EditGroupWikiPageOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, gid, slug, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for EditGroupWikiPage")
	}

	var r0 *gitlab.GroupWiki
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, string, *gitlab.EditGroupWikiPageOptions, ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error)); ok {
		return rf(gid, slug, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, string, *gitlab.EditGroupWikiPageOptions, ...gitlab.RequestOptionFunc) *gitlab.GroupWiki); ok {
		r0 = rf(gid, slug, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.GroupWiki)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, string, *gitlab.EditGroupWikiPageOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(gid, slug, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, string, *gitlab.EditGroupWikiPageOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(gid, slug, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GroupWikisServiceInterface_EditGroupWikiPage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EditGroupWikiPage'
type GroupWikisServiceInterface_EditGroupWikiPage_Call struct {
	*mock.Call
}

// EditGroupWikiPage is a helper method to define mock.On call
//   - gid interface{}
//   - slug string
//   - opt *gitlab.EditGroupWikiPageOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupWikisServiceInterface_Expecter) EditGroupWikiPage(gid interface{}, slug interface{}, opt interface{}, options ...interface{}) *GroupWikisServiceInterface_EditGroupWikiPage_Call {
	return &GroupWikisServiceInterface_EditGroupWikiPage_Call{Call: _e.mock.On("EditGroupWikiPage",
		append([]interface{}{gid, slug, opt}, options...)...)}
}

func (_c *GroupWikisServiceInterface_EditGroupWikiPage_Call) Run(run func(gid interface{}, slug string, opt *gitlab.EditGroupWikiPageOptions, options ...gitlab.RequestOptionFunc)) *GroupWikisServiceInterface_EditGroupWikiPage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(string), args[2].(*gitlab.EditGroupWikiPageOptions), variadicArgs...)
	})
	return _c
}

func (_c *GroupWikisServiceInterface_EditGroupWikiPage_Call) Return(_a0 *gitlab.GroupWiki, _a1 *gitlab.Response, _a2 error) *GroupWikisServiceInterface_EditGroupWikiPage_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *GroupWikisServiceInterface_EditGroupWikiPage_Call) RunAndReturn(run func(interface{}, string, *gitlab.EditGroupWikiPageOptions, ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error)) *GroupWikisServiceInterface_EditGroupWikiPage_Call {
	_c.Call.Return(run)
	return _c
}

// GetGroupWikiPage provides a mock function with given fields: gid, slug, opt, options
func (_m *GroupWikisServiceInterface) GetGroupWikiPage(gid interface{}, slug string, opt *gitlab.GetGroupWikiPageOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, gid, slug, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetGroupWikiPage")
	}

	var r0 *gitlab.GroupWiki
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, string, *gitlab.GetGroupWikiPageOptions, ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error)); ok {
		return rf(gid, slug, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, string, *gitlab.GetGroupWikiPageOptions, ...gitlab.RequestOptionFunc) *gitlab.GroupWiki); ok {
		r0 = rf(gid, slug, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.GroupWiki)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, string, *gitlab.GetGroupWikiPageOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(gid, slug, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, string, *gitlab.GetGroupWikiPageOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(gid, slug, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GroupWikisServiceInterface_GetGroupWikiPage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGroupWikiPage'
type GroupWikisServiceInterface_GetGroupWikiPage_Call struct {
	*mock.Call
}

// GetGroupWikiPage is a helper method to define mock.On call
//   - gid interface{}
//   - slug string
//   - opt *gitlab.GetGroupWikiPageOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupWikisServiceInterface_Expecter) GetGroupWikiPage(gid interface{}, slug interface{}, opt interface{}, options ...interface{}) *GroupWikisServiceInterface_GetGroupWikiPage_Call {
	return &GroupWikisServiceInterface_GetGroupWikiPage_Call{Call: _e.mock.On("GetGroupWikiPage",
		append([]interface{}{gid, slug, opt}, options...)...)}
}

func (_c *GroupWikisServiceInterface_GetGroupWikiPage_Call) Run(run func(gid interface{}, slug string, opt *gitlab.GetGroupWikiPageOptions, options ...gitlab.RequestOptionFunc)) *GroupWikisServiceInterface_GetGroupWikiPage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(string), args[2].(*gitlab.GetGroupWikiPageOptions), variadicArgs...)
	})
	return _c
}

func (_c *GroupWikisServiceInterface_GetGroupWikiPage_Call) Return(_a0 *gitlab.GroupWiki, _a1 *gitlab.Response, _a2 error) *GroupWikisServiceInterface_GetGroupWikiPage_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *GroupWikisServiceInterface_GetGroupWikiPage_Call) RunAndReturn(run func(interface{}, string, *gitlab.GetGroupWikiPageOptions, ...gitlab.RequestOptionFunc) (*gitlab.GroupWiki, *gitlab.Response, error)) *GroupWikisServiceInterface_GetGroupWikiPage_Call {
	_c.Call.Return(run)
	return _c
}

// ListGroupWikis provides a mock function with given fields: gid, opt, options
func (_m *GroupWikisServiceInterface) ListGroupWikis(gid interface{}, opt *gitlab.ListGroupWikisOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.GroupWiki, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, gid, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListGroupWikis")
	}

	var r0 []*gitlab.GroupWiki
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.ListGroupWikisOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.GroupWiki, *gitlab.Response, error)); ok {
		return rf(gid, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.ListGroupWikisOptions, ...gitlab.RequestOptionFunc) []*gitlab.GroupWiki); ok {
		r0 = rf(gid, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.GroupWiki)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, *gitlab.ListGroupWikisOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(gid, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, *gitlab.ListGroupWikisOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(gid, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GroupWikisServiceInterface_ListGroupWikis_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListGroupWikis'
type GroupWikisServiceInterface_ListGroupWikis_Call struct {
	*mock.Call
}

// ListGroupWikis is a helper method to define mock.On call
//   - gid interface{}
//   - opt *gitlab.ListGroupWikisOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupWikisServiceInterface_Expecter) ListGroupWikis(gid interface{}, opt interface{}, options ...interface{}) *GroupWikisServiceInterface_ListGroupWikis_Call {
	return &GroupWikisServiceInterface_ListGroupWikis_Call{Call: _e.mock.On("ListGroupWikis",
		append([]interface{}{gid, opt}, options...)...)}
}

func (_c *GroupWikisServiceInterface_ListGroupWikis_Call) Run(run func(gid interface{}, opt *gitlab.ListGroupWikisOptions, options ...gitlab.RequestOptionFunc)) *GroupWikisServiceInterface_ListGroupWikis_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(*gitlab.ListGroupWikisOptions), variadicArgs...)
	})
	return _c
}

func (_c *GroupWikisServiceInterface_ListGroupWikis_Call) Return(_a0 []*gitlab.GroupWiki, _a1 *gitlab.Response, _a2 error) *GroupWikisServiceInterface_ListGroupWikis_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *GroupWikisServiceInterface_ListGroupWikis_Call) RunAndReturn(run func(interface{}, *gitlab.ListGroupWikisOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.GroupWiki, *gitlab.Response, error)) *GroupWikisServiceInterface_ListGroupWikis_Call {
	_c.Call.Return(run)
	return _c
}

// NewGroupWikisServiceInterface creates a new instance of GroupWikisServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGroupWikisServiceInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *GroupWikisServiceInterface {
	mock := &GroupWikisServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
