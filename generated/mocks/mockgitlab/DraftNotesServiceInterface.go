// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// DraftNotesServiceInterface is an autogenerated mock type for the DraftNotesServiceInterface type
type DraftNotesServiceInterface struct {
	mock.Mock
}

type DraftNotesServiceInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *DraftNotesServiceInterface) EXPECT() *DraftNotesServiceInterface_Expecter {
	return &DraftNotesServiceInterface_Expecter{mock: &_m.Mock}
}

// CreateDraftNote provides a mock function with given fields: pid, mergeRequest, opt, options
func (_m *DraftNotesServiceInterface) CreateDraftNote(pid interface{}, mergeRequest int, opt *gitlab.CreateDraftNoteOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, mergeRequest, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateDraftNote")
	}

	var r0 *gitlab.DraftNote
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.CreateDraftNoteOptions, ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error)); ok {
		return rf(pid, mergeRequest, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.CreateDraftNoteOptions, ...gitlab.RequestOptionFunc) *gitlab.DraftNote); ok {
		r0 = rf(pid, mergeRequest, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.DraftNote)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, *gitlab.CreateDraftNoteOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, mergeRequest, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, *gitlab.CreateDraftNoteOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, mergeRequest, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DraftNotesServiceInterface_CreateDraftNote_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateDraftNote'
type DraftNotesServiceInterface_CreateDraftNote_Call struct {
	*mock.Call
}

// CreateDraftNote is a helper method to define mock.On call
//   - pid interface{}
//   - mergeRequest int
//   - opt *gitlab.CreateDraftNoteOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *DraftNotesServiceInterface_Expecter) CreateDraftNote(pid interface{}, mergeRequest interface{}, opt interface{}, options ...interface{}) *DraftNotesServiceInterface_CreateDraftNote_Call {
	return &DraftNotesServiceInterface_CreateDraftNote_Call{Call: _e.mock.On("CreateDraftNote",
		append([]interface{}{pid, mergeRequest, opt}, options...)...)}
}

func (_c *DraftNotesServiceInterface_CreateDraftNote_Call) Run(run func(pid interface{}, mergeRequest int, opt *gitlab.CreateDraftNoteOptions, options ...gitlab.RequestOptionFunc)) *DraftNotesServiceInterface_CreateDraftNote_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(*gitlab.CreateDraftNoteOptions), variadicArgs...)
	})
	return _c
}

func (_c *DraftNotesServiceInterface_CreateDraftNote_Call) Return(_a0 *gitlab.DraftNote, _a1 *gitlab.Response, _a2 error) *DraftNotesServiceInterface_CreateDraftNote_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *DraftNotesServiceInterface_CreateDraftNote_Call) RunAndReturn(run func(interface{}, int, *gitlab.CreateDraftNoteOptions, ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error)) *DraftNotesServiceInterface_CreateDraftNote_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteDraftNote provides a mock function with given fields: pid, mergeRequest, note, options
func (_m *DraftNotesServiceInterface) DeleteDraftNote(pid interface{}, mergeRequest int, note int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, mergeRequest, note)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteDraftNote")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(pid, mergeRequest, note, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(pid, mergeRequest, note, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(pid, mergeRequest, note, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DraftNotesServiceInterface_DeleteDraftNote_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteDraftNote'
type DraftNotesServiceInterface_DeleteDraftNote_Call struct {
	*mock.Call
}

// DeleteDraftNote is a helper method to define mock.On call
//   - pid interface{}
//   - mergeRequest int
//   - note int
//   - options ...gitlab.RequestOptionFunc
func (_e *DraftNotesServiceInterface_Expecter) DeleteDraftNote(pid interface{}, mergeRequest interface{}, note interface{}, options ...interface{}) *DraftNotesServiceInterface_DeleteDraftNote_Call {
	return &DraftNotesServiceInterface_DeleteDraftNote_Call{Call: _e.mock.On("DeleteDraftNote",
		append([]interface{}{pid, mergeRequest, note}, options...)...)}
}

func (_c *DraftNotesServiceInterface_DeleteDraftNote_Call) Run(run func(pid interface{}, mergeRequest int, note int, options ...gitlab.RequestOptionFunc)) *DraftNotesServiceInterface_DeleteDraftNote_Call {
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

func (_c *DraftNotesServiceInterface_DeleteDraftNote_Call) Return(_a0 *gitlab.Response, _a1 error) *DraftNotesServiceInterface_DeleteDraftNote_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DraftNotesServiceInterface_DeleteDraftNote_Call) RunAndReturn(run func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *DraftNotesServiceInterface_DeleteDraftNote_Call {
	_c.Call.Return(run)
	return _c
}

// GetDraftNote provides a mock function with given fields: pid, mergeRequest, note, options
func (_m *DraftNotesServiceInterface) GetDraftNote(pid interface{}, mergeRequest int, note int, options ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, mergeRequest, note)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetDraftNote")
	}

	var r0 *gitlab.DraftNote
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error)); ok {
		return rf(pid, mergeRequest, note, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) *gitlab.DraftNote); ok {
		r0 = rf(pid, mergeRequest, note, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.DraftNote)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, mergeRequest, note, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, mergeRequest, note, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DraftNotesServiceInterface_GetDraftNote_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDraftNote'
type DraftNotesServiceInterface_GetDraftNote_Call struct {
	*mock.Call
}

// GetDraftNote is a helper method to define mock.On call
//   - pid interface{}
//   - mergeRequest int
//   - note int
//   - options ...gitlab.RequestOptionFunc
func (_e *DraftNotesServiceInterface_Expecter) GetDraftNote(pid interface{}, mergeRequest interface{}, note interface{}, options ...interface{}) *DraftNotesServiceInterface_GetDraftNote_Call {
	return &DraftNotesServiceInterface_GetDraftNote_Call{Call: _e.mock.On("GetDraftNote",
		append([]interface{}{pid, mergeRequest, note}, options...)...)}
}

func (_c *DraftNotesServiceInterface_GetDraftNote_Call) Run(run func(pid interface{}, mergeRequest int, note int, options ...gitlab.RequestOptionFunc)) *DraftNotesServiceInterface_GetDraftNote_Call {
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

func (_c *DraftNotesServiceInterface_GetDraftNote_Call) Return(_a0 *gitlab.DraftNote, _a1 *gitlab.Response, _a2 error) *DraftNotesServiceInterface_GetDraftNote_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *DraftNotesServiceInterface_GetDraftNote_Call) RunAndReturn(run func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error)) *DraftNotesServiceInterface_GetDraftNote_Call {
	_c.Call.Return(run)
	return _c
}

// ListDraftNotes provides a mock function with given fields: pid, mergeRequest, opt, options
func (_m *DraftNotesServiceInterface) ListDraftNotes(pid interface{}, mergeRequest int, opt *gitlab.ListDraftNotesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.DraftNote, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, mergeRequest, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListDraftNotes")
	}

	var r0 []*gitlab.DraftNote
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.ListDraftNotesOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.DraftNote, *gitlab.Response, error)); ok {
		return rf(pid, mergeRequest, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, *gitlab.ListDraftNotesOptions, ...gitlab.RequestOptionFunc) []*gitlab.DraftNote); ok {
		r0 = rf(pid, mergeRequest, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.DraftNote)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, *gitlab.ListDraftNotesOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, mergeRequest, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, *gitlab.ListDraftNotesOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, mergeRequest, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DraftNotesServiceInterface_ListDraftNotes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListDraftNotes'
type DraftNotesServiceInterface_ListDraftNotes_Call struct {
	*mock.Call
}

// ListDraftNotes is a helper method to define mock.On call
//   - pid interface{}
//   - mergeRequest int
//   - opt *gitlab.ListDraftNotesOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *DraftNotesServiceInterface_Expecter) ListDraftNotes(pid interface{}, mergeRequest interface{}, opt interface{}, options ...interface{}) *DraftNotesServiceInterface_ListDraftNotes_Call {
	return &DraftNotesServiceInterface_ListDraftNotes_Call{Call: _e.mock.On("ListDraftNotes",
		append([]interface{}{pid, mergeRequest, opt}, options...)...)}
}

func (_c *DraftNotesServiceInterface_ListDraftNotes_Call) Run(run func(pid interface{}, mergeRequest int, opt *gitlab.ListDraftNotesOptions, options ...gitlab.RequestOptionFunc)) *DraftNotesServiceInterface_ListDraftNotes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(*gitlab.ListDraftNotesOptions), variadicArgs...)
	})
	return _c
}

func (_c *DraftNotesServiceInterface_ListDraftNotes_Call) Return(_a0 []*gitlab.DraftNote, _a1 *gitlab.Response, _a2 error) *DraftNotesServiceInterface_ListDraftNotes_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *DraftNotesServiceInterface_ListDraftNotes_Call) RunAndReturn(run func(interface{}, int, *gitlab.ListDraftNotesOptions, ...gitlab.RequestOptionFunc) ([]*gitlab.DraftNote, *gitlab.Response, error)) *DraftNotesServiceInterface_ListDraftNotes_Call {
	_c.Call.Return(run)
	return _c
}

// PublishAllDraftNotes provides a mock function with given fields: pid, mergeRequest, options
func (_m *DraftNotesServiceInterface) PublishAllDraftNotes(pid interface{}, mergeRequest int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, mergeRequest)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PublishAllDraftNotes")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(pid, mergeRequest, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(pid, mergeRequest, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(pid, mergeRequest, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DraftNotesServiceInterface_PublishAllDraftNotes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PublishAllDraftNotes'
type DraftNotesServiceInterface_PublishAllDraftNotes_Call struct {
	*mock.Call
}

// PublishAllDraftNotes is a helper method to define mock.On call
//   - pid interface{}
//   - mergeRequest int
//   - options ...gitlab.RequestOptionFunc
func (_e *DraftNotesServiceInterface_Expecter) PublishAllDraftNotes(pid interface{}, mergeRequest interface{}, options ...interface{}) *DraftNotesServiceInterface_PublishAllDraftNotes_Call {
	return &DraftNotesServiceInterface_PublishAllDraftNotes_Call{Call: _e.mock.On("PublishAllDraftNotes",
		append([]interface{}{pid, mergeRequest}, options...)...)}
}

func (_c *DraftNotesServiceInterface_PublishAllDraftNotes_Call) Run(run func(pid interface{}, mergeRequest int, options ...gitlab.RequestOptionFunc)) *DraftNotesServiceInterface_PublishAllDraftNotes_Call {
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

func (_c *DraftNotesServiceInterface_PublishAllDraftNotes_Call) Return(_a0 *gitlab.Response, _a1 error) *DraftNotesServiceInterface_PublishAllDraftNotes_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DraftNotesServiceInterface_PublishAllDraftNotes_Call) RunAndReturn(run func(interface{}, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *DraftNotesServiceInterface_PublishAllDraftNotes_Call {
	_c.Call.Return(run)
	return _c
}

// PublishDraftNote provides a mock function with given fields: pid, mergeRequest, note, options
func (_m *DraftNotesServiceInterface) PublishDraftNote(pid interface{}, mergeRequest int, note int, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, mergeRequest, note)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PublishDraftNote")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(pid, mergeRequest, note, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(pid, mergeRequest, note, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, int, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(pid, mergeRequest, note, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DraftNotesServiceInterface_PublishDraftNote_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PublishDraftNote'
type DraftNotesServiceInterface_PublishDraftNote_Call struct {
	*mock.Call
}

// PublishDraftNote is a helper method to define mock.On call
//   - pid interface{}
//   - mergeRequest int
//   - note int
//   - options ...gitlab.RequestOptionFunc
func (_e *DraftNotesServiceInterface_Expecter) PublishDraftNote(pid interface{}, mergeRequest interface{}, note interface{}, options ...interface{}) *DraftNotesServiceInterface_PublishDraftNote_Call {
	return &DraftNotesServiceInterface_PublishDraftNote_Call{Call: _e.mock.On("PublishDraftNote",
		append([]interface{}{pid, mergeRequest, note}, options...)...)}
}

func (_c *DraftNotesServiceInterface_PublishDraftNote_Call) Run(run func(pid interface{}, mergeRequest int, note int, options ...gitlab.RequestOptionFunc)) *DraftNotesServiceInterface_PublishDraftNote_Call {
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

func (_c *DraftNotesServiceInterface_PublishDraftNote_Call) Return(_a0 *gitlab.Response, _a1 error) *DraftNotesServiceInterface_PublishDraftNote_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DraftNotesServiceInterface_PublishDraftNote_Call) RunAndReturn(run func(interface{}, int, int, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *DraftNotesServiceInterface_PublishDraftNote_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateDraftNote provides a mock function with given fields: pid, mergeRequest, note, opt, options
func (_m *DraftNotesServiceInterface) UpdateDraftNote(pid interface{}, mergeRequest int, note int, opt *gitlab.UpdateDraftNoteOptions, options ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, pid, mergeRequest, note, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDraftNote")
	}

	var r0 *gitlab.DraftNote
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, int, int, *gitlab.UpdateDraftNoteOptions, ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error)); ok {
		return rf(pid, mergeRequest, note, opt, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, int, int, *gitlab.UpdateDraftNoteOptions, ...gitlab.RequestOptionFunc) *gitlab.DraftNote); ok {
		r0 = rf(pid, mergeRequest, note, opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.DraftNote)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, int, int, *gitlab.UpdateDraftNoteOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(pid, mergeRequest, note, opt, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, int, int, *gitlab.UpdateDraftNoteOptions, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(pid, mergeRequest, note, opt, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DraftNotesServiceInterface_UpdateDraftNote_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateDraftNote'
type DraftNotesServiceInterface_UpdateDraftNote_Call struct {
	*mock.Call
}

// UpdateDraftNote is a helper method to define mock.On call
//   - pid interface{}
//   - mergeRequest int
//   - note int
//   - opt *gitlab.UpdateDraftNoteOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *DraftNotesServiceInterface_Expecter) UpdateDraftNote(pid interface{}, mergeRequest interface{}, note interface{}, opt interface{}, options ...interface{}) *DraftNotesServiceInterface_UpdateDraftNote_Call {
	return &DraftNotesServiceInterface_UpdateDraftNote_Call{Call: _e.mock.On("UpdateDraftNote",
		append([]interface{}{pid, mergeRequest, note, opt}, options...)...)}
}

func (_c *DraftNotesServiceInterface_UpdateDraftNote_Call) Run(run func(pid interface{}, mergeRequest int, note int, opt *gitlab.UpdateDraftNoteOptions, options ...gitlab.RequestOptionFunc)) *DraftNotesServiceInterface_UpdateDraftNote_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-4)
		for i, a := range args[4:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), args[1].(int), args[2].(int), args[3].(*gitlab.UpdateDraftNoteOptions), variadicArgs...)
	})
	return _c
}

func (_c *DraftNotesServiceInterface_UpdateDraftNote_Call) Return(_a0 *gitlab.DraftNote, _a1 *gitlab.Response, _a2 error) *DraftNotesServiceInterface_UpdateDraftNote_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *DraftNotesServiceInterface_UpdateDraftNote_Call) RunAndReturn(run func(interface{}, int, int, *gitlab.UpdateDraftNoteOptions, ...gitlab.RequestOptionFunc) (*gitlab.DraftNote, *gitlab.Response, error)) *DraftNotesServiceInterface_UpdateDraftNote_Call {
	_c.Call.Return(run)
	return _c
}

// NewDraftNotesServiceInterface creates a new instance of DraftNotesServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDraftNotesServiceInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *DraftNotesServiceInterface {
	mock := &DraftNotesServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
