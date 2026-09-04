package routes

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/controller"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRouter(authController *controller.AuthController, tokenService service.TokenService) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")

	RegisterAuthRoutes(api, authController, tokenService)

	return router
}
