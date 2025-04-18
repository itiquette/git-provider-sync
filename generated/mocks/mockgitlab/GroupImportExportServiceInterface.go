// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	bytes "bytes"

	mock "github.com/stretchr/testify/mock"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GroupImportExportServiceInterface is an autogenerated mock type for the GroupImportExportServiceInterface type
type GroupImportExportServiceInterface struct {
	mock.Mock
}

type GroupImportExportServiceInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *GroupImportExportServiceInterface) EXPECT() *GroupImportExportServiceInterface_Expecter {
	return &GroupImportExportServiceInterface_Expecter{mock: &_m.Mock}
}

// ExportDownload provides a mock function with given fields: gid, options
func (_m *GroupImportExportServiceInterface) ExportDownload(gid interface{}, options ...gitlab.RequestOptionFunc) (*bytes.Reader, *gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, gid)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ExportDownload")
	}

	var r0 *bytes.Reader
	var r1 *gitlab.Response
	var r2 error
	if rf, ok := ret.Get(0).(func(interface{}, ...gitlab.RequestOptionFunc) (*bytes.Reader, *gitlab.Response, error)); ok {
		return rf(gid, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, ...gitlab.RequestOptionFunc) *bytes.Reader); ok {
		r0 = rf(gid, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bytes.Reader)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r1 = rf(gid, options...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(2).(func(interface{}, ...gitlab.RequestOptionFunc) error); ok {
		r2 = rf(gid, options...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GroupImportExportServiceInterface_ExportDownload_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExportDownload'
type GroupImportExportServiceInterface_ExportDownload_Call struct {
	*mock.Call
}

// ExportDownload is a helper method to define mock.On call
//   - gid interface{}
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupImportExportServiceInterface_Expecter) ExportDownload(gid interface{}, options ...interface{}) *GroupImportExportServiceInterface_ExportDownload_Call {
	return &GroupImportExportServiceInterface_ExportDownload_Call{Call: _e.mock.On("ExportDownload",
		append([]interface{}{gid}, options...)...)}
}

func (_c *GroupImportExportServiceInterface_ExportDownload_Call) Run(run func(gid interface{}, options ...gitlab.RequestOptionFunc)) *GroupImportExportServiceInterface_ExportDownload_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *GroupImportExportServiceInterface_ExportDownload_Call) Return(_a0 *bytes.Reader, _a1 *gitlab.Response, _a2 error) *GroupImportExportServiceInterface_ExportDownload_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *GroupImportExportServiceInterface_ExportDownload_Call) RunAndReturn(run func(interface{}, ...gitlab.RequestOptionFunc) (*bytes.Reader, *gitlab.Response, error)) *GroupImportExportServiceInterface_ExportDownload_Call {
	_c.Call.Return(run)
	return _c
}

// ImportFile provides a mock function with given fields: opt, options
func (_m *GroupImportExportServiceInterface) ImportFile(opt *gitlab.GroupImportFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, opt)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ImportFile")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(*gitlab.GroupImportFileOptions, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(opt, options...)
	}
	if rf, ok := ret.Get(0).(func(*gitlab.GroupImportFileOptions, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(opt, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(*gitlab.GroupImportFileOptions, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(opt, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GroupImportExportServiceInterface_ImportFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ImportFile'
type GroupImportExportServiceInterface_ImportFile_Call struct {
	*mock.Call
}

// ImportFile is a helper method to define mock.On call
//   - opt *gitlab.GroupImportFileOptions
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupImportExportServiceInterface_Expecter) ImportFile(opt interface{}, options ...interface{}) *GroupImportExportServiceInterface_ImportFile_Call {
	return &GroupImportExportServiceInterface_ImportFile_Call{Call: _e.mock.On("ImportFile",
		append([]interface{}{opt}, options...)...)}
}

func (_c *GroupImportExportServiceInterface_ImportFile_Call) Run(run func(opt *gitlab.GroupImportFileOptions, options ...gitlab.RequestOptionFunc)) *GroupImportExportServiceInterface_ImportFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(*gitlab.GroupImportFileOptions), variadicArgs...)
	})
	return _c
}

func (_c *GroupImportExportServiceInterface_ImportFile_Call) Return(_a0 *gitlab.Response, _a1 error) *GroupImportExportServiceInterface_ImportFile_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GroupImportExportServiceInterface_ImportFile_Call) RunAndReturn(run func(*gitlab.GroupImportFileOptions, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *GroupImportExportServiceInterface_ImportFile_Call {
	_c.Call.Return(run)
	return _c
}

// ScheduleExport provides a mock function with given fields: gid, options
func (_m *GroupImportExportServiceInterface) ScheduleExport(gid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, gid)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ScheduleExport")
	}

	var r0 *gitlab.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(interface{}, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)); ok {
		return rf(gid, options...)
	}
	if rf, ok := ret.Get(0).(func(interface{}, ...gitlab.RequestOptionFunc) *gitlab.Response); ok {
		r0 = rf(gid, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(interface{}, ...gitlab.RequestOptionFunc) error); ok {
		r1 = rf(gid, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GroupImportExportServiceInterface_ScheduleExport_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ScheduleExport'
type GroupImportExportServiceInterface_ScheduleExport_Call struct {
	*mock.Call
}

// ScheduleExport is a helper method to define mock.On call
//   - gid interface{}
//   - options ...gitlab.RequestOptionFunc
func (_e *GroupImportExportServiceInterface_Expecter) ScheduleExport(gid interface{}, options ...interface{}) *GroupImportExportServiceInterface_ScheduleExport_Call {
	return &GroupImportExportServiceInterface_ScheduleExport_Call{Call: _e.mock.On("ScheduleExport",
		append([]interface{}{gid}, options...)...)}
}

func (_c *GroupImportExportServiceInterface_ScheduleExport_Call) Run(run func(gid interface{}, options ...gitlab.RequestOptionFunc)) *GroupImportExportServiceInterface_ScheduleExport_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]gitlab.RequestOptionFunc, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(gitlab.RequestOptionFunc)
			}
		}
		run(args[0].(interface{}), variadicArgs...)
	})
	return _c
}

func (_c *GroupImportExportServiceInterface_ScheduleExport_Call) Return(_a0 *gitlab.Response, _a1 error) *GroupImportExportServiceInterface_ScheduleExport_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *GroupImportExportServiceInterface_ScheduleExport_Call) RunAndReturn(run func(interface{}, ...gitlab.RequestOptionFunc) (*gitlab.Response, error)) *GroupImportExportServiceInterface_ScheduleExport_Call {
	_c.Call.Return(run)
	return _c
}

// NewGroupImportExportServiceInterface creates a new instance of GroupImportExportServiceInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGroupImportExportServiceInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *GroupImportExportServiceInterface {
	mock := &GroupImportExportServiceInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
