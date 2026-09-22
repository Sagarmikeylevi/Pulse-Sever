package mocks

import (
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockTaskLogRepository struct {
	mock.Mock
}

func (m *MockTaskLogRepository) Upsert(log *entity.TaskLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockTaskLogRepository) FindByUserAndDate(userID uuid.UUID, date time.Time) ([]entity.TaskLog, error) {
	args := m.Called(userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.TaskLog), args.Error(1)
}

func (m *MockTaskLogRepository) FindByTaskID(taskID uuid.UUID) ([]entity.TaskLog, error) {
	args := m.Called(taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.TaskLog), args.Error(1)
}
