package service_test

import (
	"testing"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/mocks"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newTestUserService() (service.UserService, *mocks.MockUserRepository) {
	userRepo := new(mocks.MockUserRepository)
	userSvc := service.NewUserService(userRepo)
	return userSvc, userRepo
}

// ==================== FindOrCreateByEmail Tests ====================

func TestFindOrCreateByEmail_NewUser(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userRepo.On("FindByEmail", "sagar@test.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil)

	user, err := userSvc.FindOrCreateByEmail("sagar@test.com", "Asia/Kolkata")

	assert.NoError(t, err)
	assert.Equal(t, "sagar@test.com", user.Email)
	assert.True(t, user.IsEmailVerified)
	assert.Equal(t, "Asia/Kolkata", user.Timezone)
	userRepo.AssertCalled(t, "Create", mock.AnythingOfType("*entity.User"))
}

func TestFindOrCreateByEmail_ExistingUser(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	existingUser := &entity.User{
		Email:           "sagar@test.com",
		IsEmailVerified: true,
		Timezone:        "Asia/Kolkata",
	}
	existingUser.ID = uuid.New()

	userRepo.On("FindByEmail", "sagar@test.com").Return(existingUser, nil)

	user, err := userSvc.FindOrCreateByEmail("sagar@test.com", "America/New_York")

	assert.NoError(t, err)
	assert.Equal(t, existingUser.ID, user.ID)
	userRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestFindOrCreateByEmail_ExistingUser_MarksEmailVerified(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	unverifiedUser := &entity.User{
		Email:           "sagar@test.com",
		IsEmailVerified: false,
		Timezone:        "Asia/Kolkata",
	}
	unverifiedUser.ID = uuid.New()

	userRepo.On("FindByEmail", "sagar@test.com").Return(unverifiedUser, nil)
	userRepo.On("MarkEmailVerified", unverifiedUser.ID).Return(nil)

	user, err := userSvc.FindOrCreateByEmail("sagar@test.com", "UTC")

	assert.NoError(t, err)
	assert.Equal(t, unverifiedUser.ID, user.ID)
	userRepo.AssertCalled(t, "MarkEmailVerified", unverifiedUser.ID)
}

func TestFindOrCreateByEmail_NormalizesLegacyTimezone(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userRepo.On("FindByEmail", "sagar@test.com").Return(nil, gorm.ErrRecordNotFound)
	userRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil)

	user, err := userSvc.FindOrCreateByEmail("sagar@test.com", "Asia/Calcutta")

	assert.NoError(t, err)
	assert.Equal(t, "Asia/Kolkata", user.Timezone)
}

func TestFindOrCreateByEmail_InvalidTimezone(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	wrongTimezone := "Garbage/Zone"

	user, err := userSvc.FindOrCreateByEmail("sagar@test.com", wrongTimezone)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, service.ErrInvalidTimezone)

	userRepo.AssertNotCalled(t, "FindByEmail", mock.Anything)
	userRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// ==================== CheckTimezone Tests ====================

func TestCheckTimezone_Match(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userID := uuid.New()
	user := &entity.User{
		Email:    "sagar@test.com",
		Timezone: "Asia/Kolkata",
	}
	user.ID = userID

	userRepo.On("FindByID", userID).Return(user, nil)

	result, err := userSvc.CheckTimezone(userID, "Asia/Kolkata")

	assert.NoError(t, err)
	assert.True(t, result.Match)
	assert.Equal(t, "Asia/Kolkata", result.Current)
	assert.Equal(t, "Asia/Kolkata", result.Detected)
}

func TestCheckTimezone_Mismatch(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userID := uuid.New()
	user := &entity.User{
		Email:    "sagar@test.com",
		Timezone: "Asia/Kolkata",
	}
	user.ID = userID

	userRepo.On("FindByID", userID).Return(user, nil)

	result, err := userSvc.CheckTimezone(userID, "America/New_York")

	assert.NoError(t, err)
	assert.False(t, result.Match)
	assert.Equal(t, "Asia/Kolkata", result.Current)
	assert.Equal(t, "America/New_York", result.Detected)
}

func TestCheckTimezone_AliasMatch(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userID := uuid.New()
	user := &entity.User{
		Email:    "sagar@test.com",
		Timezone: "Asia/Kolkata",
	}
	user.ID = userID

	userRepo.On("FindByID", userID).Return(user, nil)

	result, err := userSvc.CheckTimezone(userID, "Asia/Calcutta")

	assert.NoError(t, err)
	assert.True(t, result.Match)
	assert.Equal(t, "Asia/Kolkata", result.Current)
	assert.Equal(t, "Asia/Kolkata", result.Detected)
}

func TestCheckTimezone_InvalidTimezone(t *testing.T) {
	userSvc, _ := newTestUserService()

	userID := uuid.New()

	result, err := userSvc.CheckTimezone(userID, "Garbage/Zone")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, service.ErrInvalidTimezone)
}

func TestCheckTimezone_UserNotFound(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userID := uuid.New()
	userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := userSvc.CheckTimezone(userID, "Asia/Kolkata")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, service.ErrUserNotFound)
}

// ==================== UpdateTimezone Tests ====================

func TestUpdateTimezone_Success(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userID := uuid.New()
	user := &entity.User{
		Email:    "sagar@test.com",
		Timezone: "Asia/Kolkata",
	}
	user.ID = userID

	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("UpdateTimezone", userID, "America/New_York").Return(nil)

	err := userSvc.UpdateTimezone(userID, "America/New_York")

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "UpdateTimezone", userID, "America/New_York")
}

func TestUpdateTimezone_NormalizesLegacyTimezone(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userID := uuid.New()
	user := &entity.User{
		Email:    "sagar@test.com",
		Timezone: "Asia/Kolkata",
	}
	user.ID = userID

	userRepo.On("FindByID", userID).Return(user, nil)
	userRepo.On("UpdateTimezone", userID, "America/New_York").Return(nil)

	err := userSvc.UpdateTimezone(userID, "US/Eastern")

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "UpdateTimezone", userID, "America/New_York")
}

func TestUpdateTimezone_InvalidTimezone(t *testing.T) {
	userSvc, _ := newTestUserService()

	userID := uuid.New()

	err := userSvc.UpdateTimezone(userID, "Garbage/Zone")

	assert.ErrorIs(t, err, service.ErrInvalidTimezone)
}

func TestUpdateTimezone_UserNotFound(t *testing.T) {
	userSvc, userRepo := newTestUserService()

	userID := uuid.New()
	userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	err := userSvc.UpdateTimezone(userID, "America/New_York")

	assert.ErrorIs(t, err, service.ErrUserNotFound)
}
