package mocks

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) FindOrCreateByEmail(email string, timezone string) (*entity.User, error) {
	args := m.Called(email, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) CheckTimezone(userID uuid.UUID, detectedTimezone string) (*service.TimezoneCheckResult, error) {
	args := m.Called(userID, detectedTimezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TimezoneCheckResult), args.Error(1)
}

func (m *MockUserService) UpdateTimezone(userID uuid.UUID, timezone string) error {
	args := m.Called(userID, timezone)
	return args.Error(0)
}
