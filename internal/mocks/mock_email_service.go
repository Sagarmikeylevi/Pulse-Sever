package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendOTP(email string, code string) error {
	args := m.Called(email, code)
	return args.Error(0)
}
