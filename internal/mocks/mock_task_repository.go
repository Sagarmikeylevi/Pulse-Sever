package mocks

import (
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(task *entity.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) FindByID(id, userID uuid.UUID) (*entity.Task, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskRepository) FindByUserID(userID uuid.UUID, taskType *string) ([]entity.Task, error) {
	args := m.Called(userID, taskType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Task), args.Error(1)
}

func (m *MockTaskRepository) FindRecurringByUserID(userID uuid.UUID) ([]entity.Task, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Task), args.Error(1)
}

func (m *MockTaskRepository) FindCheckInsByDate(userID uuid.UUID, date time.Time) ([]entity.Task, error) {
	args := m.Called(userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(task *entity.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(id, userID uuid.UUID) error {
	args := m.Called(id, userID)
	return args.Error(0)
}
