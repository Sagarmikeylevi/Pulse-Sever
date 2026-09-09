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

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) CheckTimezone(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.CheckTimezoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	result, err := c.userService.CheckTimezone(userID.(uuid.UUID), req.Timezone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTimezone):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to check timezone"})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.CheckTimezoneResponse{
		Match:    result.Match,
		Current:  result.Current,
		Detected: result.Detected,
	})
}

func (c *UserController) UpdateTimezone(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.UpdateTimezoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	err := c.userService.UpdateTimezone(userID.(uuid.UUID), req.Timezone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTimezone):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update timezone"})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{Message: "timezone updated successfully"})
}
