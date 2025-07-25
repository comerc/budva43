// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	client "github.com/zelenin/go-tdlib/client"
)

// MessageService is an autogenerated mock type for the messageService type
type MessageService struct {
	mock.Mock
}

type MessageService_Expecter struct {
	mock *mock.Mock
}

func (_m *MessageService) EXPECT() *MessageService_Expecter {
	return &MessageService_Expecter{mock: &_m.Mock}
}

// GetFormattedText provides a mock function with given fields: message
func (_m *MessageService) GetFormattedText(message *client.Message) *client.FormattedText {
	ret := _m.Called(message)

	if len(ret) == 0 {
		panic("no return value specified for GetFormattedText")
	}

	var r0 *client.FormattedText
	if rf, ok := ret.Get(0).(func(*client.Message) *client.FormattedText); ok {
		r0 = rf(message)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.FormattedText)
		}
	}

	return r0
}

// MessageService_GetFormattedText_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFormattedText'
type MessageService_GetFormattedText_Call struct {
	*mock.Call
}

// GetFormattedText is a helper method to define mock.On call
//   - message *client.Message
func (_e *MessageService_Expecter) GetFormattedText(message interface{}) *MessageService_GetFormattedText_Call {
	return &MessageService_GetFormattedText_Call{Call: _e.mock.On("GetFormattedText", message)}
}

func (_c *MessageService_GetFormattedText_Call) Run(run func(message *client.Message)) *MessageService_GetFormattedText_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*client.Message))
	})
	return _c
}

func (_c *MessageService_GetFormattedText_Call) Return(_a0 *client.FormattedText) *MessageService_GetFormattedText_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MessageService_GetFormattedText_Call) RunAndReturn(run func(*client.Message) *client.FormattedText) *MessageService_GetFormattedText_Call {
	_c.Call.Return(run)
	return _c
}

// GetInputMessageContent provides a mock function with given fields: message, formattedText
func (_m *MessageService) GetInputMessageContent(message *client.Message, formattedText *client.FormattedText) client.InputMessageContent {
	ret := _m.Called(message, formattedText)

	if len(ret) == 0 {
		panic("no return value specified for GetInputMessageContent")
	}

	var r0 client.InputMessageContent
	if rf, ok := ret.Get(0).(func(*client.Message, *client.FormattedText) client.InputMessageContent); ok {
		r0 = rf(message, formattedText)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.InputMessageContent)
		}
	}

	return r0
}

// MessageService_GetInputMessageContent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetInputMessageContent'
type MessageService_GetInputMessageContent_Call struct {
	*mock.Call
}

// GetInputMessageContent is a helper method to define mock.On call
//   - message *client.Message
//   - formattedText *client.FormattedText
func (_e *MessageService_Expecter) GetInputMessageContent(message interface{}, formattedText interface{}) *MessageService_GetInputMessageContent_Call {
	return &MessageService_GetInputMessageContent_Call{Call: _e.mock.On("GetInputMessageContent", message, formattedText)}
}

func (_c *MessageService_GetInputMessageContent_Call) Run(run func(message *client.Message, formattedText *client.FormattedText)) *MessageService_GetInputMessageContent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*client.Message), args[1].(*client.FormattedText))
	})
	return _c
}

func (_c *MessageService_GetInputMessageContent_Call) Return(_a0 client.InputMessageContent) *MessageService_GetInputMessageContent_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MessageService_GetInputMessageContent_Call) RunAndReturn(run func(*client.Message, *client.FormattedText) client.InputMessageContent) *MessageService_GetInputMessageContent_Call {
	_c.Call.Return(run)
	return _c
}

// GetReplyMarkupData provides a mock function with given fields: message
func (_m *MessageService) GetReplyMarkupData(message *client.Message) []byte {
	ret := _m.Called(message)

	if len(ret) == 0 {
		panic("no return value specified for GetReplyMarkupData")
	}

	var r0 []byte
	if rf, ok := ret.Get(0).(func(*client.Message) []byte); ok {
		r0 = rf(message)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// MessageService_GetReplyMarkupData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetReplyMarkupData'
type MessageService_GetReplyMarkupData_Call struct {
	*mock.Call
}

// GetReplyMarkupData is a helper method to define mock.On call
//   - message *client.Message
func (_e *MessageService_Expecter) GetReplyMarkupData(message interface{}) *MessageService_GetReplyMarkupData_Call {
	return &MessageService_GetReplyMarkupData_Call{Call: _e.mock.On("GetReplyMarkupData", message)}
}

func (_c *MessageService_GetReplyMarkupData_Call) Run(run func(message *client.Message)) *MessageService_GetReplyMarkupData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*client.Message))
	})
	return _c
}

func (_c *MessageService_GetReplyMarkupData_Call) Return(_a0 []byte) *MessageService_GetReplyMarkupData_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MessageService_GetReplyMarkupData_Call) RunAndReturn(run func(*client.Message) []byte) *MessageService_GetReplyMarkupData_Call {
	_c.Call.Return(run)
	return _c
}

// NewMessageService creates a new instance of MessageService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMessageService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MessageService {
	mock := &MessageService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
