package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/repository"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrOTPCooldown        = errors.New("please wait before requesting a new OTP")

	ErrOTPExpired         = errors.New("OTP has expired")
	ErrOTPMaxAttempts     = errors.New("too many failed attempts, request a new OTP")
	ErrInvalidOTP         = errors.New("invalid OTP")
	ErrPasswordNotSet     = errors.New("password not set, use OTP login")
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrTokenExpired       = errors.New("refresh token has expired")
	ErrTokenReuse         = errors.New("token reuse detected, possible theft")
)

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	SendOTP(email string) error
	VerifyOTP(email string, code string) (*AuthTokens, error)
	LoginWithPassword(email string, password string) (*AuthTokens, error)
	SetPassword(userID uuid.UUID, password string) error
	RefreshToken(rawRefreshToken string) (*AuthTokens, error)
}

type authService struct {
	userRepo         repository.UserRepository
	otpRepo          repository.OTPRepository
	refreshTokenRepo repository.RefreshTokenRepository
	tokenService     TokenService
	emailService     EmailService
	jwtConfig        shared.JWTConfig
}

func NewAuthService(
	userRepo repository.UserRepository,
	otpRepo repository.OTPRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	tokenService TokenService,
	emailService EmailService,
	jwtConfig shared.JWTConfig,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		otpRepo:          otpRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenService:     tokenService,
		emailService:     emailService,
		jwtConfig:        jwtConfig,
	}
}

// SendOTP generates a 6-digit OTP, hashes it, stores it, and sends it to the user's email.
func (s *authService) SendOTP(email string) error {
	// Check cooldown — 90 seconds between OTP requests
	existingOTP, err := s.otpRepo.FindLatestActiveByEmail(email)
	if err == nil && existingOTP != nil {
		elapsed := time.Since(existingOTP.CreatedAt)
		if elapsed < 90*time.Second {
			return ErrOTPCooldown
		}
	}

	// Generate 6-digit OTP
	code, err := generateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash the OTP code
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash OTP: %w", err)
	}

	otp := &entity.OTP{
		Email:     email,
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := s.otpRepo.Create(otp); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send OTP via email (or log in dev)
	if err := s.emailService.SendOTP(email, code); err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	return nil
}

// VerifyOTP verifies the OTP code. If valid, creates user (if new) and returns tokens.
func (s *authService) VerifyOTP(email string, code string) (*AuthTokens, error) {
	otp, err := s.otpRepo.FindLatestActiveByEmail(email)
	if err != nil {
		return nil, ErrInvalidOTP
	}

	// Check expiry
	if time.Now().After(otp.ExpiresAt) {
		return nil, ErrOTPExpired
	}

	// Check max attempts
	if otp.Attempts >= 3 {
		return nil, ErrOTPMaxAttempts
	}

	// Verify the code
	if err := bcrypt.CompareHashAndPassword([]byte(otp.CodeHash), []byte(code)); err != nil {
		_ = s.otpRepo.IncrementAttempts(otp)
		return nil, ErrInvalidOTP
	}

	// Mark OTP as used
	_ = s.otpRepo.MarkAsUsed(otp)

	// Find or create user
	user, err := s.userRepo.FindByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = &entity.User{
			Email:           email,
			IsEmailVerified: true,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	} else {
		// Existing user — ensure email is marked verified
		if !user.IsEmailVerified {
			_ = s.userRepo.MarkEmailVerified(user.ID)
		}
	}

	return s.generateTokenPair(user)
}

// LoginWithPassword authenticates a user with email and password.
func (s *authService) LoginWithPassword(email string, password string) (*AuthTokens, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		return nil, ErrPasswordNotSet
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

// SetPassword sets a password for a user who doesn't have one yet.
func (s *authService) SetPassword(userID uuid.UUID, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.userRepo.UpdatePassword(userID, string(hash))
}

// RefreshToken validates a refresh token and issues a new token pair (rotation).
func (s *authService) RefreshToken(rawRefreshToken string) (*AuthTokens, error) {
	tokenHash := s.tokenService.HashToken(rawRefreshToken)

	token, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiry
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Token reuse detection — if token is already used, someone stole it
	if token.Status == entity.TokenStatusUsed {
		// Revoke the entire family
		_ = s.refreshTokenRepo.RevokeFamily(token.FamilyID)
		return nil, ErrTokenReuse
	}

	if token.Status == entity.TokenStatusRevoked {
		return nil, ErrTokenRevoked
	}

	// Mark current token as used
	_ = s.refreshTokenRepo.MarkAsUsed(token)

	// Get user for new access token claims
	user, err := s.userRepo.FindByID(token.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Generate new token pair in the same family
	return s.generateTokenPairWithFamily(user, token.FamilyID)
}

// generateTokenPair creates a new access + refresh token pair with a new family.
func (s *authService) generateTokenPair(user *entity.User) (*AuthTokens, error) {
	familyID := uuid.New()
	return s.generateTokenPairWithFamily(user, familyID)
}

// generateTokenPairWithFamily creates tokens using an existing family ID (for rotation).
func (s *authService) generateTokenPairWithFamily(user *entity.User, familyID uuid.UUID) (*AuthTokens, error) {
	accessToken, err := s.tokenService.GenerateAccessToken(TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	rawRefresh, hashRefresh, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshToken := &entity.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashRefresh,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(time.Duration(s.jwtConfig.RefreshTokenExpiry) * time.Minute),
	}

	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

// generateOTPCode generates a cryptographically secure 6-digit OTP.
func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
