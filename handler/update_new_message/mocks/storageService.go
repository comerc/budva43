// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// StorageService is an autogenerated mock type for the storageService type
type StorageService struct {
	mock.Mock
}

type StorageService_Expecter struct {
	mock *mock.Mock
}

func (_m *StorageService) EXPECT() *StorageService_Expecter {
	return &StorageService_Expecter{mock: &_m.Mock}
}

// IncrementForwardedMessages provides a mock function with given fields: toChatId, date
func (_m *StorageService) IncrementForwardedMessages(toChatId int64, date string) {
	_m.Called(toChatId, date)
}

// StorageService_IncrementForwardedMessages_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IncrementForwardedMessages'
type StorageService_IncrementForwardedMessages_Call struct {
	*mock.Call
}

// IncrementForwardedMessages is a helper method to define mock.On call
//   - toChatId int64
//   - date string
func (_e *StorageService_Expecter) IncrementForwardedMessages(toChatId interface{}, date interface{}) *StorageService_IncrementForwardedMessages_Call {
	return &StorageService_IncrementForwardedMessages_Call{Call: _e.mock.On("IncrementForwardedMessages", toChatId, date)}
}

func (_c *StorageService_IncrementForwardedMessages_Call) Run(run func(toChatId int64, date string)) *StorageService_IncrementForwardedMessages_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64), args[1].(string))
	})
	return _c
}

func (_c *StorageService_IncrementForwardedMessages_Call) Return() *StorageService_IncrementForwardedMessages_Call {
	_c.Call.Return()
	return _c
}

func (_c *StorageService_IncrementForwardedMessages_Call) RunAndReturn(run func(int64, string)) *StorageService_IncrementForwardedMessages_Call {
	_c.Run(run)
	return _c
}

// IncrementViewedMessages provides a mock function with given fields: toChatId, date
func (_m *StorageService) IncrementViewedMessages(toChatId int64, date string) {
	_m.Called(toChatId, date)
}

// StorageService_IncrementViewedMessages_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IncrementViewedMessages'
type StorageService_IncrementViewedMessages_Call struct {
	*mock.Call
}

// IncrementViewedMessages is a helper method to define mock.On call
//   - toChatId int64
//   - date string
func (_e *StorageService_Expecter) IncrementViewedMessages(toChatId interface{}, date interface{}) *StorageService_IncrementViewedMessages_Call {
	return &StorageService_IncrementViewedMessages_Call{Call: _e.mock.On("IncrementViewedMessages", toChatId, date)}
}

func (_c *StorageService_IncrementViewedMessages_Call) Run(run func(toChatId int64, date string)) *StorageService_IncrementViewedMessages_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64), args[1].(string))
	})
	return _c
}

func (_c *StorageService_IncrementViewedMessages_Call) Return() *StorageService_IncrementViewedMessages_Call {
	_c.Call.Return()
	return _c
}

func (_c *StorageService_IncrementViewedMessages_Call) RunAndReturn(run func(int64, string)) *StorageService_IncrementViewedMessages_Call {
	_c.Run(run)
	return _c
}

// NewStorageService creates a new instance of StorageService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorageService(t interface {
	mock.TestingT
	Cleanup(func())
}) *StorageService {
	mock := &StorageService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
