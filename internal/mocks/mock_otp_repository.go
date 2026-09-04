package mocks

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockOTPRepository struct {
	mock.Mock
}

func (m *MockOTPRepository) Create(otp *entity.OTP) error {
	args := m.Called(otp)
	return args.Error(0)
}

func (m *MockOTPRepository) FindLatestActiveByEmail(email string) (*entity.OTP, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OTP), args.Error(1)
}

func (m *MockOTPRepository) IncrementAttempts(otp *entity.OTP) error {
	args := m.Called(otp)
	return args.Error(0)
}

func (m *MockOTPRepository) MarkAsUsed(otp *entity.OTP) error {
	args := m.Called(otp)
	return args.Error(0)
}
