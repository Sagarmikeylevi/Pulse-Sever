package mocks

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/stretchr/testify/mock"
)

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(claims service.TokenClaims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ValidateAccessToken(tokenString string) (*service.TokenClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenClaims), args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken() (string, string, error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockTokenService) HashToken(raw string) string {
	args := m.Called(raw)
	return args.String(0)
}
