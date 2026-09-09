package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidTimezone = errors.New("invalid IANA timezone")
	ErrUserNotFound    = errors.New("user not found")
)

type TimezoneCheckResult struct {
	Match    bool
	Current  string
	Detected string
}

type UserService interface {
	FindOrCreateByEmail(email string, timezone string) (*entity.User, error)
	CheckTimezone(userID uuid.UUID, detectedTimezone string) (*TimezoneCheckResult, error)
	UpdateTimezone(userID uuid.UUID, timezone string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) FindOrCreateByEmail(email string, timezone string) (*entity.User, error) {
	if err := validateTimezone(timezone); err != nil {
		return nil, err
	}
	user, err := s.userRepo.FindByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = &entity.User{
			Email:           email,
			IsEmailVerified: true,
			Timezone:        timezone,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		return user, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Existing user — ensure email is marked verified
	if !user.IsEmailVerified {
		_ = s.userRepo.MarkEmailVerified(user.ID)
	}
	return user, nil
}

func (s *userService) CheckTimezone(userID uuid.UUID, detectedTimezone string) (*TimezoneCheckResult, error) {
	if err := validateTimezone(detectedTimezone); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &TimezoneCheckResult{
		Match:    user.Timezone == detectedTimezone,
		Current:  user.Timezone,
		Detected: detectedTimezone,
	}, nil
}

func (s *userService) UpdateTimezone(userID uuid.UUID, timezone string) error {
	if err := validateTimezone(timezone); err != nil {
		return err
	}

	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	return s.userRepo.UpdateTimezone(userID, timezone)
}

// Helper function to validate timezone
func validateTimezone(tz string) error {
	_, err := time.LoadLocation(tz)
	if err != nil {
		return ErrInvalidTimezone
	}
	return nil
}
