// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	client "github.com/zelenin/go-tdlib/client"
)

// TelegramRepo is an autogenerated mock type for the telegramRepo type
type TelegramRepo struct {
	mock.Mock
}

type TelegramRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *TelegramRepo) EXPECT() *TelegramRepo_Expecter {
	return &TelegramRepo_Expecter{mock: &_m.Mock}
}

// DeleteMessages provides a mock function with given fields: _a0
func (_m *TelegramRepo) DeleteMessages(_a0 *client.DeleteMessagesRequest) (*client.Ok, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for DeleteMessages")
	}

	var r0 *client.Ok
	var r1 error
	if rf, ok := ret.Get(0).(func(*client.DeleteMessagesRequest) (*client.Ok, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*client.DeleteMessagesRequest) *client.Ok); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Ok)
		}
	}

	if rf, ok := ret.Get(1).(func(*client.DeleteMessagesRequest) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TelegramRepo_DeleteMessages_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteMessages'
type TelegramRepo_DeleteMessages_Call struct {
	*mock.Call
}

// DeleteMessages is a helper method to define mock.On call
//   - _a0 *client.DeleteMessagesRequest
func (_e *TelegramRepo_Expecter) DeleteMessages(_a0 interface{}) *TelegramRepo_DeleteMessages_Call {
	return &TelegramRepo_DeleteMessages_Call{Call: _e.mock.On("DeleteMessages", _a0)}
}

func (_c *TelegramRepo_DeleteMessages_Call) Run(run func(_a0 *client.DeleteMessagesRequest)) *TelegramRepo_DeleteMessages_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*client.DeleteMessagesRequest))
	})
	return _c
}

func (_c *TelegramRepo_DeleteMessages_Call) Return(_a0 *client.Ok, _a1 error) *TelegramRepo_DeleteMessages_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TelegramRepo_DeleteMessages_Call) RunAndReturn(run func(*client.DeleteMessagesRequest) (*client.Ok, error)) *TelegramRepo_DeleteMessages_Call {
	_c.Call.Return(run)
	return _c
}

// NewTelegramRepo creates a new instance of TelegramRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTelegramRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *TelegramRepo {
	mock := &TelegramRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
