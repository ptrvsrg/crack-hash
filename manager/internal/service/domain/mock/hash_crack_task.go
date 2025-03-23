// Code generated by mockery v2.53.2. DO NOT EDIT.

package mock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/ptrvsrg/crack-hash/manager/pkg/model"
)

// HashCrackTaskMock is an autogenerated mock type for the HashCrackTask type
type HashCrackTaskMock struct {
	mock.Mock
}

type HashCrackTaskMock_Expecter struct {
	mock *mock.Mock
}

func (_m *HashCrackTaskMock) EXPECT() *HashCrackTaskMock_Expecter {
	return &HashCrackTaskMock_Expecter{mock: &_m.Mock}
}

// CreateTask provides a mock function with given fields: ctx, input
func (_m *HashCrackTaskMock) CreateTask(ctx context.Context, input *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error) {
	ret := _m.Called(ctx, input)

	if len(ret) == 0 {
		panic("no return value specified for CreateTask")
	}

	var r0 *model.HashCrackTaskIDOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error)); ok {
		return rf(ctx, input)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.HashCrackTaskInput) *model.HashCrackTaskIDOutput); ok {
		r0 = rf(ctx, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.HashCrackTaskIDOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.HashCrackTaskInput) error); ok {
		r1 = rf(ctx, input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HashCrackTaskMock_CreateTask_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateTask'
type HashCrackTaskMock_CreateTask_Call struct {
	*mock.Call
}

// CreateTask is a helper method to define mock.On call
//   - ctx context.Context
//   - input *model.HashCrackTaskInput
func (_e *HashCrackTaskMock_Expecter) CreateTask(ctx interface{}, input interface{}) *HashCrackTaskMock_CreateTask_Call {
	return &HashCrackTaskMock_CreateTask_Call{Call: _e.mock.On("CreateTask", ctx, input)}
}

func (_c *HashCrackTaskMock_CreateTask_Call) Run(run func(ctx context.Context, input *model.HashCrackTaskInput)) *HashCrackTaskMock_CreateTask_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*model.HashCrackTaskInput))
	})
	return _c
}

func (_c *HashCrackTaskMock_CreateTask_Call) Return(_a0 *model.HashCrackTaskIDOutput, _a1 error) *HashCrackTaskMock_CreateTask_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *HashCrackTaskMock_CreateTask_Call) RunAndReturn(run func(context.Context, *model.HashCrackTaskInput) (*model.HashCrackTaskIDOutput, error)) *HashCrackTaskMock_CreateTask_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteExpiredTasks provides a mock function with given fields: ctx
func (_m *HashCrackTaskMock) DeleteExpiredTasks(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for DeleteExpiredTasks")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// HashCrackTaskMock_DeleteExpiredTasks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteExpiredTasks'
type HashCrackTaskMock_DeleteExpiredTasks_Call struct {
	*mock.Call
}

// DeleteExpiredTasks is a helper method to define mock.On call
//   - ctx context.Context
func (_e *HashCrackTaskMock_Expecter) DeleteExpiredTasks(ctx interface{}) *HashCrackTaskMock_DeleteExpiredTasks_Call {
	return &HashCrackTaskMock_DeleteExpiredTasks_Call{Call: _e.mock.On("DeleteExpiredTasks", ctx)}
}

func (_c *HashCrackTaskMock_DeleteExpiredTasks_Call) Run(run func(ctx context.Context)) *HashCrackTaskMock_DeleteExpiredTasks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *HashCrackTaskMock_DeleteExpiredTasks_Call) Return(_a0 error) *HashCrackTaskMock_DeleteExpiredTasks_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *HashCrackTaskMock_DeleteExpiredTasks_Call) RunAndReturn(run func(context.Context) error) *HashCrackTaskMock_DeleteExpiredTasks_Call {
	_c.Call.Return(run)
	return _c
}

// FinishTimeoutTasks provides a mock function with given fields: ctx
func (_m *HashCrackTaskMock) FinishTimeoutTasks(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for FinishTimeoutTasks")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// HashCrackTaskMock_FinishTimeoutTasks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FinishTimeoutTasks'
type HashCrackTaskMock_FinishTimeoutTasks_Call struct {
	*mock.Call
}

// FinishTimeoutTasks is a helper method to define mock.On call
//   - ctx context.Context
func (_e *HashCrackTaskMock_Expecter) FinishTimeoutTasks(ctx interface{}) *HashCrackTaskMock_FinishTimeoutTasks_Call {
	return &HashCrackTaskMock_FinishTimeoutTasks_Call{Call: _e.mock.On("FinishTimeoutTasks", ctx)}
}

func (_c *HashCrackTaskMock_FinishTimeoutTasks_Call) Run(run func(ctx context.Context)) *HashCrackTaskMock_FinishTimeoutTasks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *HashCrackTaskMock_FinishTimeoutTasks_Call) Return(_a0 error) *HashCrackTaskMock_FinishTimeoutTasks_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *HashCrackTaskMock_FinishTimeoutTasks_Call) RunAndReturn(run func(context.Context) error) *HashCrackTaskMock_FinishTimeoutTasks_Call {
	_c.Call.Return(run)
	return _c
}

// GetTaskStatus provides a mock function with given fields: ctx, id
func (_m *HashCrackTaskMock) GetTaskStatus(ctx context.Context, id string) (*model.HashCrackTaskStatusOutput, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetTaskStatus")
	}

	var r0 *model.HashCrackTaskStatusOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.HashCrackTaskStatusOutput, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.HashCrackTaskStatusOutput); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.HashCrackTaskStatusOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HashCrackTaskMock_GetTaskStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTaskStatus'
type HashCrackTaskMock_GetTaskStatus_Call struct {
	*mock.Call
}

// GetTaskStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *HashCrackTaskMock_Expecter) GetTaskStatus(ctx interface{}, id interface{}) *HashCrackTaskMock_GetTaskStatus_Call {
	return &HashCrackTaskMock_GetTaskStatus_Call{Call: _e.mock.On("GetTaskStatus", ctx, id)}
}

func (_c *HashCrackTaskMock_GetTaskStatus_Call) Run(run func(ctx context.Context, id string)) *HashCrackTaskMock_GetTaskStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *HashCrackTaskMock_GetTaskStatus_Call) Return(_a0 *model.HashCrackTaskStatusOutput, _a1 error) *HashCrackTaskMock_GetTaskStatus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *HashCrackTaskMock_GetTaskStatus_Call) RunAndReturn(run func(context.Context, string) (*model.HashCrackTaskStatusOutput, error)) *HashCrackTaskMock_GetTaskStatus_Call {
	_c.Call.Return(run)
	return _c
}

// SaveResultSubtask provides a mock function with given fields: ctx, input
func (_m *HashCrackTaskMock) SaveResultSubtask(ctx context.Context, input *model.HashCrackTaskWebhookInput) error {
	ret := _m.Called(ctx, input)

	if len(ret) == 0 {
		panic("no return value specified for SaveResultSubtask")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.HashCrackTaskWebhookInput) error); ok {
		r0 = rf(ctx, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// HashCrackTaskMock_SaveResultSubtask_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SaveResultSubtask'
type HashCrackTaskMock_SaveResultSubtask_Call struct {
	*mock.Call
}

// SaveResultSubtask is a helper method to define mock.On call
//   - ctx context.Context
//   - input *model.HashCrackTaskWebhookInput
func (_e *HashCrackTaskMock_Expecter) SaveResultSubtask(ctx interface{}, input interface{}) *HashCrackTaskMock_SaveResultSubtask_Call {
	return &HashCrackTaskMock_SaveResultSubtask_Call{Call: _e.mock.On("SaveResultSubtask", ctx, input)}
}

func (_c *HashCrackTaskMock_SaveResultSubtask_Call) Run(run func(ctx context.Context, input *model.HashCrackTaskWebhookInput)) *HashCrackTaskMock_SaveResultSubtask_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*model.HashCrackTaskWebhookInput))
	})
	return _c
}

func (_c *HashCrackTaskMock_SaveResultSubtask_Call) Return(_a0 error) *HashCrackTaskMock_SaveResultSubtask_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *HashCrackTaskMock_SaveResultSubtask_Call) RunAndReturn(run func(context.Context, *model.HashCrackTaskWebhookInput) error) *HashCrackTaskMock_SaveResultSubtask_Call {
	_c.Call.Return(run)
	return _c
}

// NewHashCrackTaskMock creates a new instance of HashCrackTaskMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHashCrackTaskMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *HashCrackTaskMock {
	mock := &HashCrackTaskMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
