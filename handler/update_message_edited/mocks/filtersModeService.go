// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	domain "github.com/comerc/budva43/app/domain"
	client "github.com/zelenin/go-tdlib/client"

	mock "github.com/stretchr/testify/mock"
)

// FiltersModeService is an autogenerated mock type for the filtersModeService type
type FiltersModeService struct {
	mock.Mock
}

type FiltersModeService_Expecter struct {
	mock *mock.Mock
}

func (_m *FiltersModeService) EXPECT() *FiltersModeService_Expecter {
	return &FiltersModeService_Expecter{mock: &_m.Mock}
}

// Map provides a mock function with given fields: formattedText, forwardRule
func (_m *FiltersModeService) Map(formattedText *client.FormattedText, forwardRule *domain.ForwardRule) string {
	ret := _m.Called(formattedText, forwardRule)

	if len(ret) == 0 {
		panic("no return value specified for Map")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(*client.FormattedText, *domain.ForwardRule) string); ok {
		r0 = rf(formattedText, forwardRule)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// FiltersModeService_Map_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Map'
type FiltersModeService_Map_Call struct {
	*mock.Call
}

// Map is a helper method to define mock.On call
//   - formattedText *client.FormattedText
//   - forwardRule *domain.ForwardRule
func (_e *FiltersModeService_Expecter) Map(formattedText interface{}, forwardRule interface{}) *FiltersModeService_Map_Call {
	return &FiltersModeService_Map_Call{Call: _e.mock.On("Map", formattedText, forwardRule)}
}

func (_c *FiltersModeService_Map_Call) Run(run func(formattedText *client.FormattedText, forwardRule *domain.ForwardRule)) *FiltersModeService_Map_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*client.FormattedText), args[1].(*domain.ForwardRule))
	})
	return _c
}

func (_c *FiltersModeService_Map_Call) Return(_a0 string) *FiltersModeService_Map_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FiltersModeService_Map_Call) RunAndReturn(run func(*client.FormattedText, *domain.ForwardRule) string) *FiltersModeService_Map_Call {
	_c.Call.Return(run)
	return _c
}

// NewFiltersModeService creates a new instance of FiltersModeService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFiltersModeService(t interface {
	mock.TestingT
	Cleanup(func())
}) *FiltersModeService {
	mock := &FiltersModeService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
