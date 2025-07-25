// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	client "github.com/zelenin/go-tdlib/client"

	mock "github.com/stretchr/testify/mock"
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

// GetChatHistory provides a mock function with given fields: _a0
func (_m *TelegramRepo) GetChatHistory(_a0 *client.GetChatHistoryRequest) (*client.Messages, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetChatHistory")
	}

	var r0 *client.Messages
	var r1 error
	if rf, ok := ret.Get(0).(func(*client.GetChatHistoryRequest) (*client.Messages, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*client.GetChatHistoryRequest) *client.Messages); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Messages)
		}
	}

	if rf, ok := ret.Get(1).(func(*client.GetChatHistoryRequest) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TelegramRepo_GetChatHistory_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetChatHistory'
type TelegramRepo_GetChatHistory_Call struct {
	*mock.Call
}

// GetChatHistory is a helper method to define mock.On call
//   - _a0 *client.GetChatHistoryRequest
func (_e *TelegramRepo_Expecter) GetChatHistory(_a0 interface{}) *TelegramRepo_GetChatHistory_Call {
	return &TelegramRepo_GetChatHistory_Call{Call: _e.mock.On("GetChatHistory", _a0)}
}

func (_c *TelegramRepo_GetChatHistory_Call) Run(run func(_a0 *client.GetChatHistoryRequest)) *TelegramRepo_GetChatHistory_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*client.GetChatHistoryRequest))
	})
	return _c
}

func (_c *TelegramRepo_GetChatHistory_Call) Return(_a0 *client.Messages, _a1 error) *TelegramRepo_GetChatHistory_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TelegramRepo_GetChatHistory_Call) RunAndReturn(run func(*client.GetChatHistoryRequest) (*client.Messages, error)) *TelegramRepo_GetChatHistory_Call {
	_c.Call.Return(run)
	return _c
}

// LoadChats provides a mock function with given fields: _a0
func (_m *TelegramRepo) LoadChats(_a0 *client.LoadChatsRequest) (*client.Ok, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for LoadChats")
	}

	var r0 *client.Ok
	var r1 error
	if rf, ok := ret.Get(0).(func(*client.LoadChatsRequest) (*client.Ok, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*client.LoadChatsRequest) *client.Ok); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Ok)
		}
	}

	if rf, ok := ret.Get(1).(func(*client.LoadChatsRequest) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TelegramRepo_LoadChats_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoadChats'
type TelegramRepo_LoadChats_Call struct {
	*mock.Call
}

// LoadChats is a helper method to define mock.On call
//   - _a0 *client.LoadChatsRequest
func (_e *TelegramRepo_Expecter) LoadChats(_a0 interface{}) *TelegramRepo_LoadChats_Call {
	return &TelegramRepo_LoadChats_Call{Call: _e.mock.On("LoadChats", _a0)}
}

func (_c *TelegramRepo_LoadChats_Call) Run(run func(_a0 *client.LoadChatsRequest)) *TelegramRepo_LoadChats_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*client.LoadChatsRequest))
	})
	return _c
}

func (_c *TelegramRepo_LoadChats_Call) Return(_a0 *client.Ok, _a1 error) *TelegramRepo_LoadChats_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TelegramRepo_LoadChats_Call) RunAndReturn(run func(*client.LoadChatsRequest) (*client.Ok, error)) *TelegramRepo_LoadChats_Call {
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
