package routes

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/controller"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/middleware"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterTaskRoutes(router *gin.RouterGroup, taskController *controller.TaskController, tokenService service.TokenService) {
	tasks := router.Group("/tasks")
	tasks.Use(middleware.AuthRequired(tokenService))

	tasks.POST("", taskController.CreateTask)
	tasks.GET("", taskController.ListTasks)

	tasks.GET("/today", taskController.GetTodayTasks)

	tasks.GET("/:id", taskController.GetTask)
	tasks.PUT("/:id", taskController.UpdateTask)
	tasks.DELETE("/:id", taskController.DeleteTask)

	tasks.POST("/:id/logs", taskController.LogTask)
	tasks.GET("/:id/logs", taskController.GetTaskLogs)
}
