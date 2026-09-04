package service

import (
	"fmt"
	"log"
)

type EmailService interface {
	SendOTP(email string, code string) error
}

type emailService struct {
	appEnv string
}

func NewEmailService(appEnv string) EmailService {
	return &emailService{appEnv: appEnv}
}

func (s *emailService) SendOTP(email string, code string) error {
	if s.appEnv == "production" {
		// TODO: Implement AWS SES integration
		return fmt.Errorf("SES not implemented yet")
	}

	log.Printf("[DEV] OTP for %s: %s", email, code)
	return nil
}
