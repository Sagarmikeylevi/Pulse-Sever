package registry

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/controller"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/repository"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/routes"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *shared.Config) *gin.Engine {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Services
	tokenService := service.NewTokenService(cfg.JWT)
	emailService := service.NewEmailService(cfg.App.Env)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(
		userRepo,
		otpRepo,
		refreshTokenRepo,
		tokenService,
		emailService,
		cfg.JWT,
	)

	// Controllers
	authController := controller.NewAuthController(authService, userService)
	userController := controller.NewUserController(userService)

	// Router
	router := routes.SetupRouter(authController, userController, tokenService)

	return router
}
