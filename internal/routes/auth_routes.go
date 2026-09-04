package routes

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/controller"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/middleware"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup, authController *controller.AuthController, tokenService service.TokenService) {
	auth := router.Group("/auth")

	// Public routes — no JWT required
	auth.POST("/otp/send", authController.SendOTP)
	auth.POST("/otp/verify", authController.VerifyOTP)
	auth.POST("/login", authController.Login)
	auth.POST("/token/refresh", authController.RefreshToken)

	// Protected routes — JWT required
	protected := auth.Group("")
	protected.Use(middleware.AuthRequired(tokenService))
	protected.POST("/password/set", authController.SetPassword)
}
