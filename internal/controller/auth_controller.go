package controller

import (
	"errors"
	"net/http"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/dto"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	authService service.AuthService
	userService service.UserService
}

func NewAuthController(authService service.AuthService, userService service.UserService) *AuthController {
	return &AuthController{authService: authService, userService: userService}
}

func (c *AuthController) SendOTP(ctx *gin.Context) {
	var req dto.SendOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	err := c.authService.SendOTP(req.Email)
	if err != nil {
		if errors.Is(err, service.ErrOTPCooldown) {
			ctx.JSON(http.StatusTooManyRequests, dto.ErrorResponse{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to send OTP"})
		return
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{Message: "OTP sent successfully"})
}

func (c *AuthController) VerifyOTP(ctx *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	// Step 1: Verify OTP
	if err := c.authService.VerifyOTP(req.Email, req.Code); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOTP):
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrOTPExpired):
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrOTPMaxAttempts):
			ctx.JSON(http.StatusTooManyRequests, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "verification failed"})
		}
		return
	}

	// Step 2: Find or create user
	user, err := c.userService.FindOrCreateByEmail(req.Email, req.Timezone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTimezone):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to process user"})
		}
		return
	}

	// Step 3: Generate tokens
	tokens, err := c.authService.GenerateTokens(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to generate tokens"})
		return
	}

	ctx.JSON(http.StatusOK, dto.AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	tokens, err := c.authService.LoginWithPassword(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrPasswordNotSet):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "login failed"})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (c *AuthController) SetPassword(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.SetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	if err := c.authService.SetPassword(userID.(uuid.UUID), req.Password); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to set password"})
		return
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{Message: "password set successfully"})
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	tokens, err := c.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenReuse):
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token reuse detected, please login again"})
		case errors.Is(err, service.ErrInvalidToken):
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrTokenRevoked):
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrTokenExpired):
			ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "token refresh failed"})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
