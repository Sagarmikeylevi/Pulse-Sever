package routes

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/controller"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/middleware"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.RouterGroup, userController *controller.UserController, tokenService service.TokenService) {
	user := router.Group("/user")
	user.Use(middleware.AuthRequired(tokenService))

	user.POST("/timezone/check", userController.CheckTimezone)
	user.PUT("/timezone", userController.UpdateTimezone)
}
