package mocks

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(token *entity.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByTokenHash(tokenHash string) (*entity.RefreshToken, error) {
	args := m.Called(tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) MarkAsUsed(token *entity.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeFamily(familyID uuid.UUID) error {
	args := m.Called(familyID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}
