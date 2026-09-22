package mocks

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/dto"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(userID uuid.UUID, req dto.CreateTaskRequest) (*entity.Task, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskService) GetTask(id, userID uuid.UUID) (*entity.Task, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskService) ListTasks(userID uuid.UUID, taskType *string) ([]entity.Task, error) {
	args := m.Called(userID, taskType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(id, userID uuid.UUID, req dto.UpdateTaskRequest) (*entity.Task, error) {
	args := m.Called(id, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(id, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockTaskService) LogTask(taskID, userID uuid.UUID, req dto.LogTaskRequest) (*entity.TaskLog, error) {
	args := m.Called(taskID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TaskLog), args.Error(1)
}

func (m *MockTaskService) GetTaskLogs(taskID, userID uuid.UUID) ([]entity.TaskLog, error) {
	args := m.Called(taskID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.TaskLog), args.Error(1)
}

func (m *MockTaskService) GetTodayTasks(userID uuid.UUID) (*dto.TodayResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TodayResponse), args.Error(1)
}
