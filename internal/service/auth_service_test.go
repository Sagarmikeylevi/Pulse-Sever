package service_test

import (
	"testing"
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/mocks"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func newTestAuthService() (
	service.AuthService,
	*mocks.MockUserRepository,
	*mocks.MockOTPRepository,
	*mocks.MockRefreshTokenRepository,
	*mocks.MockTokenService,
	*mocks.MockEmailService,
) {
	userRepo := new(mocks.MockUserRepository)
	otpRepo := new(mocks.MockOTPRepository)
	refreshTokenRepo := new(mocks.MockRefreshTokenRepository)
	tokenSvc := new(mocks.MockTokenService)
	emailSvc := new(mocks.MockEmailService)

	jwtConfig := shared.JWTConfig{
		Secret:             "test-secret",
		AccessTokenExpiry:  15,
		RefreshTokenExpiry: 43200,
	}

	authSvc := service.NewAuthService(userRepo, otpRepo, refreshTokenRepo, tokenSvc, emailSvc, jwtConfig)

	return authSvc, userRepo, otpRepo, refreshTokenRepo, tokenSvc, emailSvc
}

// ==================== SendOTP Tests ====================

func TestSendOTP_Success(t *testing.T) {
	authSvc, _, otpRepo, _, _, emailSvc := newTestAuthService()

	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(nil, gorm.ErrRecordNotFound)
	otpRepo.On("Create", mock.AnythingOfType("*entity.OTP")).Return(nil)
	emailSvc.On("SendOTP", "sagar@test.com", mock.AnythingOfType("string")).Return(nil)

	err := authSvc.SendOTP("sagar@test.com")

	assert.NoError(t, err)
	otpRepo.AssertCalled(t, "Create", mock.AnythingOfType("*entity.OTP"))
	emailSvc.AssertCalled(t, "SendOTP", "sagar@test.com", mock.AnythingOfType("string"))
}

func TestSendOTP_Cooldown(t *testing.T) {
	authSvc, _, otpRepo, _, _, _ := newTestAuthService()

	existingOTP := &entity.OTP{
		Email:     "sagar@test.com",
		CreatedAt: time.Now().Add(-30 * time.Second),
	}
	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(existingOTP, nil)

	err := authSvc.SendOTP("sagar@test.com")

	assert.ErrorIs(t, err, service.ErrOTPCooldown)
}

func TestSendOTP_CooldownExpired(t *testing.T) {
	authSvc, _, otpRepo, _, _, emailSvc := newTestAuthService()

	existingOTP := &entity.OTP{
		Email:     "sagar@test.com",
		CreatedAt: time.Now().Add(-2 * time.Minute),
	}
	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(existingOTP, nil)
	otpRepo.On("Create", mock.AnythingOfType("*entity.OTP")).Return(nil)
	emailSvc.On("SendOTP", "sagar@test.com", mock.AnythingOfType("string")).Return(nil)

	err := authSvc.SendOTP("sagar@test.com")

	assert.NoError(t, err)
}

// ==================== VerifyOTP Tests ====================

func TestVerifyOTP_Success_NewUser(t *testing.T) {
	authSvc, userRepo, otpRepo, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	otp := &entity.OTP{
		ID:        uuid.New(),
		Email:     "sagar@test.com",
		CodeHash:  string(codeHash),
		Attempts:  0,
		Status:    entity.OTPStatusActive,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(otp, nil)
	otpRepo.On("MarkAsUsed", otp).Return(nil)
	userRepo.On("FindByEmail", "sagar@test.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil)
	tokenSvc.On("GenerateAccessToken", mock.AnythingOfType("service.TokenClaims")).Return("access-token", nil)
	tokenSvc.On("GenerateRefreshToken").Return("raw-refresh", "hashed-refresh", nil)
	refreshTokenRepo.On("Create", mock.AnythingOfType("*entity.RefreshToken")).Return(nil)

	tokens, err := authSvc.VerifyOTP("sagar@test.com", "123456")

	assert.NoError(t, err)
	assert.Equal(t, "access-token", tokens.AccessToken)
	assert.Equal(t, "raw-refresh", tokens.RefreshToken)
	userRepo.AssertCalled(t, "Create", mock.AnythingOfType("*entity.User"))
}

func TestVerifyOTP_Success_ExistingUser(t *testing.T) {
	authSvc, userRepo, otpRepo, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	otp := &entity.OTP{
		ID:        uuid.New(),
		Email:     "sagar@test.com",
		CodeHash:  string(codeHash),
		Attempts:  0,
		Status:    entity.OTPStatusActive,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	existingUser := &entity.User{
		Email:           "sagar@test.com",
		IsEmailVerified: true,
	}
	existingUser.ID = uuid.New()

	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(otp, nil)
	otpRepo.On("MarkAsUsed", otp).Return(nil)
	userRepo.On("FindByEmail", "sagar@test.com").Return(existingUser, nil)
	tokenSvc.On("GenerateAccessToken", mock.AnythingOfType("service.TokenClaims")).Return("access-token", nil)
	tokenSvc.On("GenerateRefreshToken").Return("raw-refresh", "hashed-refresh", nil)
	refreshTokenRepo.On("Create", mock.AnythingOfType("*entity.RefreshToken")).Return(nil)

	tokens, err := authSvc.VerifyOTP("sagar@test.com", "123456")

	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	userRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestVerifyOTP_ExpiredOTP(t *testing.T) {
	authSvc, _, otpRepo, _, _, _ := newTestAuthService()

	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	otp := &entity.OTP{
		ID:        uuid.New(),
		Email:     "sagar@test.com",
		CodeHash:  string(codeHash),
		Attempts:  0,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
		CreatedAt: time.Now().Add(-6 * time.Minute),
	}
	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(otp, nil)

	tokens, err := authSvc.VerifyOTP("sagar@test.com", "123456")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrOTPExpired)
}

func TestVerifyOTP_MaxAttempts(t *testing.T) {
	authSvc, _, otpRepo, _, _, _ := newTestAuthService()

	otp := &entity.OTP{
		ID:        uuid.New(),
		Email:     "sagar@test.com",
		Attempts:  3,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(otp, nil)

	tokens, err := authSvc.VerifyOTP("sagar@test.com", "123456")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrOTPMaxAttempts)
}

func TestVerifyOTP_WrongCode(t *testing.T) {
	authSvc, _, otpRepo, _, _, _ := newTestAuthService()

	codeHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	otp := &entity.OTP{
		ID:        uuid.New(),
		Email:     "sagar@test.com",
		CodeHash:  string(codeHash),
		Attempts:  0,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(otp, nil)
	otpRepo.On("IncrementAttempts", otp).Return(nil)

	tokens, err := authSvc.VerifyOTP("sagar@test.com", "000000")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrInvalidOTP)
	otpRepo.AssertCalled(t, "IncrementAttempts", otp)
}

func TestVerifyOTP_NoActiveOTP(t *testing.T) {
	authSvc, _, otpRepo, _, _, _ := newTestAuthService()

	otpRepo.On("FindLatestActiveByEmail", "sagar@test.com").Return(nil, gorm.ErrRecordNotFound)

	tokens, err := authSvc.VerifyOTP("sagar@test.com", "123456")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrInvalidOTP)
}

// ==================== LoginWithPassword Tests ====================

func TestLoginWithPassword_Success(t *testing.T) {
	authSvc, userRepo, _, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("mypassword123"), bcrypt.DefaultCost)
	hashStr := string(passwordHash)
	user := &entity.User{
		Email:           "sagar@test.com",
		PasswordHash:    &hashStr,
		IsEmailVerified: true,
	}
	user.ID = uuid.New()

	userRepo.On("FindByEmail", "sagar@test.com").Return(user, nil)
	tokenSvc.On("GenerateAccessToken", mock.AnythingOfType("service.TokenClaims")).Return("access-token", nil)
	tokenSvc.On("GenerateRefreshToken").Return("raw-refresh", "hashed-refresh", nil)
	refreshTokenRepo.On("Create", mock.AnythingOfType("*entity.RefreshToken")).Return(nil)

	tokens, err := authSvc.LoginWithPassword("sagar@test.com", "mypassword123")

	assert.NoError(t, err)
	assert.Equal(t, "access-token", tokens.AccessToken)
	assert.Equal(t, "raw-refresh", tokens.RefreshToken)
}

func TestLoginWithPassword_WrongPassword(t *testing.T) {
	authSvc, userRepo, _, _, _, _ := newTestAuthService()

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("mypassword123"), bcrypt.DefaultCost)
	hashStr := string(passwordHash)
	user := &entity.User{
		Email:        "sagar@test.com",
		PasswordHash: &hashStr,
	}
	user.ID = uuid.New()

	userRepo.On("FindByEmail", "sagar@test.com").Return(user, nil)

	tokens, err := authSvc.LoginWithPassword("sagar@test.com", "wrongpassword")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestLoginWithPassword_UserNotFound(t *testing.T) {
	authSvc, userRepo, _, _, _, _ := newTestAuthService()

	userRepo.On("FindByEmail", "nobody@test.com").Return(nil, gorm.ErrRecordNotFound)

	tokens, err := authSvc.LoginWithPassword("nobody@test.com", "password")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestLoginWithPassword_PasswordNotSet(t *testing.T) {
	authSvc, userRepo, _, _, _, _ := newTestAuthService()

	user := &entity.User{
		Email:        "sagar@test.com",
		PasswordHash: nil,
	}
	user.ID = uuid.New()

	userRepo.On("FindByEmail", "sagar@test.com").Return(user, nil)

	tokens, err := authSvc.LoginWithPassword("sagar@test.com", "password")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrPasswordNotSet)
}

// ==================== SetPassword Tests ====================

func TestSetPassword_Success(t *testing.T) {
	authSvc, userRepo, _, _, _, _ := newTestAuthService()

	userID := uuid.New()
	userRepo.On("UpdatePassword", userID, mock.AnythingOfType("string")).Return(nil)

	err := authSvc.SetPassword(userID, "newpassword123")

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "UpdatePassword", userID, mock.AnythingOfType("string"))
}

// ==================== RefreshToken Tests ====================

func TestRefreshToken_Success(t *testing.T) {
	authSvc, userRepo, _, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	userID := uuid.New()
	familyID := uuid.New()
	token := &entity.RefreshToken{
		UserID:    userID,
		TokenHash: "hashed-token",
		FamilyID:  familyID,
		Status:    entity.TokenStatusActive,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	token.ID = uuid.New()

	user := &entity.User{
		Email: "sagar@test.com",
	}
	user.ID = userID

	tokenSvc.On("HashToken", "raw-token").Return("hashed-token")
	refreshTokenRepo.On("FindByTokenHash", "hashed-token").Return(token, nil)
	refreshTokenRepo.On("MarkAsUsed", token).Return(nil)
	userRepo.On("FindByID", userID).Return(user, nil)
	tokenSvc.On("GenerateAccessToken", mock.AnythingOfType("service.TokenClaims")).Return("new-access-token", nil)
	tokenSvc.On("GenerateRefreshToken").Return("new-raw-refresh", "new-hashed-refresh", nil)
	refreshTokenRepo.On("Create", mock.AnythingOfType("*entity.RefreshToken")).Return(nil)

	tokens, err := authSvc.RefreshToken("raw-token")

	assert.NoError(t, err)
	assert.Equal(t, "new-access-token", tokens.AccessToken)
	assert.Equal(t, "new-raw-refresh", tokens.RefreshToken)
	refreshTokenRepo.AssertCalled(t, "MarkAsUsed", token)
}

func TestRefreshToken_TokenReuse_TheftDetected(t *testing.T) {
	authSvc, _, _, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	familyID := uuid.New()
	token := &entity.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: "hashed-token",
		FamilyID:  familyID,
		Status:    entity.TokenStatusUsed,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	token.ID = uuid.New()

	tokenSvc.On("HashToken", "stolen-token").Return("hashed-token")
	refreshTokenRepo.On("FindByTokenHash", "hashed-token").Return(token, nil)
	refreshTokenRepo.On("RevokeFamily", familyID).Return(nil)

	tokens, err := authSvc.RefreshToken("stolen-token")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrTokenReuse)
	refreshTokenRepo.AssertCalled(t, "RevokeFamily", familyID)
}

func TestRefreshToken_Revoked(t *testing.T) {
	authSvc, _, _, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	token := &entity.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: "hashed-token",
		FamilyID:  uuid.New(),
		Status:    entity.TokenStatusRevoked,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	token.ID = uuid.New()

	tokenSvc.On("HashToken", "revoked-token").Return("hashed-token")
	refreshTokenRepo.On("FindByTokenHash", "hashed-token").Return(token, nil)

	tokens, err := authSvc.RefreshToken("revoked-token")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrTokenRevoked)
}

func TestRefreshToken_Expired(t *testing.T) {
	authSvc, _, _, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	token := &entity.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: "hashed-token",
		FamilyID:  uuid.New(),
		Status:    entity.TokenStatusActive,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	token.ID = uuid.New()

	tokenSvc.On("HashToken", "expired-token").Return("hashed-token")
	refreshTokenRepo.On("FindByTokenHash", "hashed-token").Return(token, nil)

	tokens, err := authSvc.RefreshToken("expired-token")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrTokenExpired)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	authSvc, _, _, refreshTokenRepo, tokenSvc, _ := newTestAuthService()

	tokenSvc.On("HashToken", "garbage-token").Return("hashed-garbage")
	refreshTokenRepo.On("FindByTokenHash", "hashed-garbage").Return(nil, gorm.ErrRecordNotFound)

	tokens, err := authSvc.RefreshToken("garbage-token")

	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, service.ErrInvalidToken)
}
