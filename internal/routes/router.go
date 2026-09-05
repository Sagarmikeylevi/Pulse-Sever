package routes

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/controller"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(authController *controller.AuthController, tokenService service.TokenService) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := router.Group("/api/v1")

	RegisterAuthRoutes(api, authController, tokenService)

	return router
}
