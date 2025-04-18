// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// IssueBoardsServiceInterface is an autogenerated mock type for the IssueBoardsServiceInterface type
type IssueBoardsServiceInterface struct {
	mock.Mock
}

type IssueBoardsServiceInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *IssueBoardsServiceInterface) EXPECT() *IssueBoardsServiceInterface_Expecter {
	return &IssueBoardsServiceInterface_Expecter{mock: &_m.Mock}
}

// CreateIssueBoard provides a mock function with given fields: pid, opt, options
func (_m *IssueBoardsServiceInterface) CreateIssueBoard(pid interface{}, opt *gitlab.CreateIssueBoardOptions, options ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateIssueBoard")
	}

	var r0 *gitlab.IssueBoard
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.CreateIssueBoardOptions, ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error)); ok {
		return rf(pid, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.CreateIssueBoardOptions, ...gitlab.RequestOptionFunc) *gitlab.IssueBoard); ok {
		r0 = rf(pid, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.IssueBoard)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, *gitlab.CreateIssueBoardOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, *gitlab.CreateIssueBoardOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_CreateIssueBoard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateIssueBoard'
type IssueBoardsServiceInterface_CreateIssueBoard_Call struct {
	*mock.Call
}

// CreateIssueBoard is a helper method to define mock.On call
//   - pid interface{}
//   - opt *gitlab.CreateIssueBoardOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) CreateIssueBoard(pid interface{}, opt interface{}, options ...interface{}) *IssueBoardsServiceInterface_CreateIssueBoard_Call {
	return &IssueBoardsServiceInterface_CreateIssueBoard_Call{Call: _e.mock.On("CreateIssueBoard",
		append([]interface{}{pid, opt}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_CreateIssueBoard_Call) Run(run func(pid interface{}, opt *gitlab.CreateIssueBoardOptions, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_CreateIssueBoard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(*gitlab.CreateIssueBoardOptions), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_CreateIssueBoard_Call) Return(_a0 *gitlab.IssueBoard, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_CreateIssueBoard_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_CreateIssueBoard_Call) RunAndReturn(run func(interface{}, *gitlab.CreateIssueBoardOptions, ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error)) *IssueBoardsServiceInterface_CreateIssueBoard_Call {
	_c.Call.Return(run)
	return _c
}

// CreateIssueBoardList provides a mock function with given fields: pid, board, opt, options
func (_m *IssueBoardsServiceInterface) CreateIssueBoardList(pid interface{}, board int, opt *gitlab.CreateIssueBoardListOptions, options ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateIssueBoardList")
	}

	var r0 *gitlab.BoardList
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.CreateIssueBoardListOptions, ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error)); ok {
		return rf(pid, board, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.CreateIssueBoardListOptions, ...gitlab.RequestOptionFunc) *gitlab.BoardList); ok {
		r0 = rf(pid, board, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.BoardList)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, *gitlab.CreateIssueBoardListOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, board, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, *gitlab.CreateIssueBoardListOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, board, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_CreateIssueBoardList_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateIssueBoardList'
type IssueBoardsServiceInterface_CreateIssueBoardList_Call struct {
	*mock.Call
}

// CreateIssueBoardList is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - opt *gitlab.CreateIssueBoardListOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) CreateIssueBoardList(pid interface{}, board interface{}, opt interface{}, options ...interface{}) *IssueBoardsServiceInterface_CreateIssueBoardList_Call {
	return &IssueBoardsServiceInterface_CreateIssueBoardList_Call{Call: _e.mock.On("CreateIssueBoardList",
		append([]interface{}{pid, board, opt}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_CreateIssueBoardList_Call) Run(run func(pid interface{}, board int, opt *gitlab.CreateIssueBoardListOptions, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_CreateIssueBoardList_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(*gitlab.CreateIssueBoardListOptions), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_CreateIssueBoardList_Call) Return(_a0 *gitlab.BoardList, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_CreateIssueBoardList_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_CreateIssueBoardList_Call) RunAndReturn(run func(interface{}, int, *gitlab.CreateIssueBoardListOptions, ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error)) *IssueBoardsServiceInterface_CreateIssueBoardList_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteIssueBoard provides a mock function with given fields: pid, board, options
func (_m *IssueBoardsServiceInterface) DeleteIssueBoard(pid interface{}, board int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteIssueBoard")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(pid, board, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(pid, board, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(pid, board, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IssueBoardsServiceInterface_DeleteIssueBoard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteIssueBoard'
type IssueBoardsServiceInterface_DeleteIssueBoard_Call struct {
	*mock.Call
}

// DeleteIssueBoard is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) DeleteIssueBoard(pid interface{}, board interface{}, options ...interface{}) *IssueBoardsServiceInterface_DeleteIssueBoard_Call {
	return &IssueBoardsServiceInterface_DeleteIssueBoard_Call{Call: _e.mock.On("DeleteIssueBoard",
		append([]interface{}{pid, board}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_DeleteIssueBoard_Call) Run(run func(pid interface{}, board int, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_DeleteIssueBoard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_DeleteIssueBoard_Call) Return(_a0 *gitlab.Response, _a1 error) *IssueBoardsServiceInterface_DeleteIssueBoard_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IssueBoardsServiceInterface_DeleteIssueBoard_Call) RunAndReturn(run func(interface{}, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *IssueBoardsServiceInterface_DeleteIssueBoard_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteIssueBoardList provides a mock function with given fields: pid, board, list, options
func (_m *IssueBoardsServiceInterface) DeleteIssueBoardList(pid interface{}, board int, list int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board, list)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteIssueBoardList")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(pid, board, list, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(pid, board, list, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(pid, board, list, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IssueBoardsServiceInterface_DeleteIssueBoardList_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteIssueBoardList'
type IssueBoardsServiceInterface_DeleteIssueBoardList_Call struct {
	*mock.Call
}

// DeleteIssueBoardList is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - list int
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) DeleteIssueBoardList(pid interface{}, board interface{}, list interface{}, options ...interface{}) *IssueBoardsServiceInterface_DeleteIssueBoardList_Call {
	return &IssueBoardsServiceInterface_DeleteIssueBoardList_Call{Call: _e.mock.On("DeleteIssueBoardList",
		append([]interface{}{pid, board, list}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_DeleteIssueBoardList_Call) Run(run func(pid interface{}, board int, list int, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_DeleteIssueBoardList_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(int), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_DeleteIssueBoardList_Call) Return(_a0 *gitlab.Response, _a1 error) *IssueBoardsServiceInterface_DeleteIssueBoardList_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IssueBoardsServiceInterface_DeleteIssueBoardList_Call) RunAndReturn(run func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *IssueBoardsServiceInterface_DeleteIssueBoardList_Call {
	_c.Call.Return(run)
	return _c
}

// GetIssueBoard provides a mock function with given fields: pid, board, options
func (_m *IssueBoardsServiceInterface) GetIssueBoard(pid interface{}, board int, options ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetIssueBoard")
	}

	var r0 *gitlab.IssueBoard
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error)); ok {
		return rf(pid, board, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, ...gitlab.RequestOptionFunc) *gitlab.IssueBoard); ok {
		r0 = rf(pid, board, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.IssueBoard)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, board, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, board, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_GetIssueBoard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIssueBoard'
type IssueBoardsServiceInterface_GetIssueBoard_Call struct {
	*mock.Call
}

// GetIssueBoard is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) GetIssueBoard(pid interface{}, board interface{}, options ...interface{}) *IssueBoardsServiceInterface_GetIssueBoard_Call {
	return &IssueBoardsServiceInterface_GetIssueBoard_Call{Call: _e.mock.On("GetIssueBoard",
		append([]interface{}{pid, board}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_GetIssueBoard_Call) Run(run func(pid interface{}, board int, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_GetIssueBoard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_GetIssueBoard_Call) Return(_a0 *gitlab.IssueBoard, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_GetIssueBoard_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_GetIssueBoard_Call) RunAndReturn(run func(interface{}, int, ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error)) *IssueBoardsServiceInterface_GetIssueBoard_Call {
	_c.Call.Return(run)
	return _c
}

// GetIssueBoardList provides a mock function with given fields: pid, board, list, options
func (_m *IssueBoardsServiceInterface) GetIssueBoardList(pid interface{}, board int, list int, options ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board, list)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetIssueBoardList")
	}

	var r0 *gitlab.BoardList
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error)); ok {
		return rf(pid, board, list, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) *gitlab.BoardList); ok {
		r0 = rf(pid, board, list, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.BoardList)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, board, list, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, board, list, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_GetIssueBoardList_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIssueBoardList'
type IssueBoardsServiceInterface_GetIssueBoardList_Call struct {
	*mock.Call
}

// GetIssueBoardList is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - list int
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) GetIssueBoardList(pid interface{}, board interface{}, list interface{}, options ...interface{}) *IssueBoardsServiceInterface_GetIssueBoardList_Call {
	return &IssueBoardsServiceInterface_GetIssueBoardList_Call{Call: _e.mock.On("GetIssueBoardList",
		append([]interface{}{pid, board, list}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_GetIssueBoardList_Call) Run(run func(pid interface{}, board int, list int, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_GetIssueBoardList_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(int), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_GetIssueBoardList_Call) Return(_a0 *gitlab.BoardList, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_GetIssueBoardList_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_GetIssueBoardList_Call) RunAndReturn(run func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error)) *IssueBoardsServiceInterface_GetIssueBoardList_Call {
	_c.Call.Return(run)
	return _c
}

// GetIssueBoardLists provides a mock function with given fields: pid, board, opt, options
func (_m *IssueBoardsServiceInterface) GetIssueBoardLists(pid interface{}, board int, opt *gitlab.GetIssueBoardListsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.BoardList, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetIssueBoardLists")
	}

	var r0 []*gitlab.BoardList
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.GetIssueBoardListsOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.BoardList, *gitlab.Response, error)); ok {
		return rf(pid, board, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.GetIssueBoardListsOptions, ...gitlab.RequestOptionFunc) []*gitlab.BoardList); ok {
		r0 = rf(pid, board, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.BoardList)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, *gitlab.GetIssueBoardListsOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, board, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, *gitlab.GetIssueBoardListsOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, board, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_GetIssueBoardLists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIssueBoardLists'
type IssueBoardsServiceInterface_GetIssueBoardLists_Call struct {
	*mock.Call
}

// GetIssueBoardLists is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - opt *gitlab.GetIssueBoardListsOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) GetIssueBoardLists(pid interface{}, board interface{}, opt interface{}, options ...interface{}) *IssueBoardsServiceInterface_GetIssueBoardLists_Call {
	return &IssueBoardsServiceInterface_GetIssueBoardLists_Call{Call: _e.mock.On("GetIssueBoardLists",
		append([]interface{}{pid, board, opt}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_GetIssueBoardLists_Call) Run(run func(pid interface{}, board int, opt *gitlab.GetIssueBoardListsOptions, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_GetIssueBoardLists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(*gitlab.GetIssueBoardListsOptions), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_GetIssueBoardLists_Call) Return(_a0 []*gitlab.BoardList, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_GetIssueBoardLists_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_GetIssueBoardLists_Call) RunAndReturn(run func(interface{}, int, *gitlab.GetIssueBoardListsOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.BoardList, *gitlab.Response, error)) *IssueBoardsServiceInterface_GetIssueBoardLists_Call {
	_c.Call.Return(run)
	return _c
}

// ListIssueBoards provides a mock function with given fields: pid, opt, options
func (_m *IssueBoardsServiceInterface) ListIssueBoards(pid interface{}, opt *gitlab.ListIssueBoardsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.IssueBoard, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListIssueBoards")
	}

	var r0 []*gitlab.IssueBoard
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.ListIssueBoardsOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.IssueBoard, *gitlab.Response, error)); ok {
		return rf(pid, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, *gitlab.ListIssueBoardsOptions, ...gitlab.RequestOptionFunc) []*gitlab.IssueBoard); ok {
		r0 = rf(pid, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.IssueBoard)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, *gitlab.ListIssueBoardsOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, *gitlab.ListIssueBoardsOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_ListIssueBoards_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListIssueBoards'
type IssueBoardsServiceInterface_ListIssueBoards_Call struct {
	*mock.Call
}

// ListIssueBoards is a helper method to define mock.On call
//   - pid interface{}
//   - opt *gitlab.ListIssueBoardsOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) ListIssueBoards(pid interface{}, opt interface{}, options ...interface{}) *IssueBoardsServiceInterface_ListIssueBoards_Call {
	return &IssueBoardsServiceInterface_ListIssueBoards_Call{Call: _e.mock.On("ListIssueBoards",
		append([]interface{}{pid, opt}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_ListIssueBoards_Call) Run(run func(pid interface{}, opt *gitlab.ListIssueBoardsOptions, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_ListIssueBoards_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(*gitlab.ListIssueBoardsOptions), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_ListIssueBoards_Call) Return(_a0 []*gitlab.IssueBoard, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_ListIssueBoards_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_ListIssueBoards_Call) RunAndReturn(run func(interface{}, *gitlab.ListIssueBoardsOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.IssueBoard, *gitlab.Response, error)) *IssueBoardsServiceInterface_ListIssueBoards_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateIssueBoard provides a mock function with given fields: pid, board, opt, options
func (_m *IssueBoardsServiceInterface) UpdateIssueBoard(pid interface{}, board int, opt *gitlab.UpdateIssueBoardOptions, options ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateIssueBoard")
	}

	var r0 *gitlab.IssueBoard
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.UpdateIssueBoardOptions, ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error)); ok {
		return rf(pid, board, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.UpdateIssueBoardOptions, ...gitlab.RequestOptionFunc) *gitlab.IssueBoard); ok {
		r0 = rf(pid, board, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.IssueBoard)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, *gitlab.UpdateIssueBoardOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, board, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, *gitlab.UpdateIssueBoardOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, board, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_UpdateIssueBoard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateIssueBoard'
type IssueBoardsServiceInterface_UpdateIssueBoard_Call struct {
	*mock.Call
}

// UpdateIssueBoard is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - opt *gitlab.UpdateIssueBoardOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) UpdateIssueBoard(pid interface{}, board interface{}, opt interface{}, options ...interface{}) *IssueBoardsServiceInterface_UpdateIssueBoard_Call {
	return &IssueBoardsServiceInterface_UpdateIssueBoard_Call{Call: _e.mock.On("UpdateIssueBoard",
		append([]interface{}{pid, board, opt}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_UpdateIssueBoard_Call) Run(run func(pid interface{}, board int, opt *gitlab.UpdateIssueBoardOptions, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_UpdateIssueBoard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(*gitlab.UpdateIssueBoardOptions), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_UpdateIssueBoard_Call) Return(_a0 *gitlab.IssueBoard, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_UpdateIssueBoard_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_UpdateIssueBoard_Call) RunAndReturn(run func(interface{}, int, *gitlab.UpdateIssueBoardOptions, ...gitlab.RequestOptionFunc) (*gitlab.IssueBoard, *gitlab.Response, error)) *IssueBoardsServiceInterface_UpdateIssueBoard_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateIssueBoardList provides a mock function with given fields: pid, board, list, opt, options
func (_m *IssueBoardsServiceInterface) UpdateIssueBoardList(pid interface{}, board int, list int, opt *gitlab.UpdateIssueBoardListOptions, options ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, board, list, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateIssueBoardList")
	}

	var r0 *gitlab.BoardList
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, int, *gitlab.UpdateIssueBoardListOptions, ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error)); ok {
		return rf(pid, board, list, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, int, *gitlab.UpdateIssueBoardListOptions, ...gitlab.RequestOptionFunc) *gitlab.BoardList); ok {
		r0 = rf(pid, board, list, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.BoardList)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, int, *gitlab.UpdateIssueBoardListOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, board, list, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, int, *gitlab.UpdateIssueBoardListOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, board, list, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// IssueBoardsServiceInterface_UpdateIssueBoardList_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateIssueBoardList'
type IssueBoardsServiceInterface_UpdateIssueBoardList_Call struct {
	*mock.Call
}

// UpdateIssueBoardList is a helper method to define mock.On call
//   - pid interface{}
//   - board int
//   - list int
//   - opt *gitlab.UpdateIssueBoardListOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *IssueBoardsServiceInterface_Expecter) UpdateIssueBoardList(pid interface{}, board interface{}, list interface{}, opt interface{}, options ...interface{}) *IssueBoardsServiceInterface_UpdateIssueBoardList_Call {
	return &IssueBoardsServiceInterface_UpdateIssueBoardList_Call{Call: _e.mock.On("UpdateIssueBoardList",
		append([]interface{}{pid, board, list, opt}, options...)...)}
}

func (_c *IssueBoardsServiceInterface_UpdateIssueBoardList_Call) Run(run func(pid interface{}, board int, list int, opt *gitlab.UpdateIssueBoardListOptions, options ...gitlab.RequestOptionFunc)) *IssueBoardsServiceInterface_UpdateIssueBoardList_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-4)
		for i, a := range args[4:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(int), args[3].(*gitlab.UpdateIssueBoardListOptions), variadicArgs...)
	})
	return _c
}

func (_c *IssueBoardsServiceInterface_UpdateIssueBoardList_Call) Return(_a0 *gitlab.BoardList, _a1 *gitlab.Response, _a2 error) *IssueBoardsServiceInterface_UpdateIssueBoardList_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *IssueBoardsServiceInterface_UpdateIssueBoardList_Call) RunAndReturn(run func(interface{}, int, int, *gitlab.UpdateIssueBoardListOptions, ...gitlab.RequestOptionFunc) (*gitlab.BoardList, *gitlab.Response, error)) *IssueBoardsServiceInterface_UpdateIssueBoardList_Call {
	_c.Call.Return(run)
	return _c
}

// NewIssueBoardsServiceInterface creates a new instance of IssueBoardsServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIssueBoardsServiceInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *IssueBoardsServiceInterface {
	mock := &IssueBoardsServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
