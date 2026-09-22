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

type TaskController struct {
	taskService service.TaskService
}

func NewTaskController(taskService service.TaskService) *TaskController {
	return &TaskController{taskService: taskService}
}

// @Summary      Create task
// @Description  Creates a new recurring task or daily check-in
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateTaskRequest true "Task details"
// @Success      201 {object} dto.TaskResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks [post]
func (c *TaskController) CreateTask(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	task, err := c.taskService.CreateTask(userID.(uuid.UUID), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecurrenceRequired),
			errors.Is(err, service.ErrInvalidRecurrence):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create task"})
		}
		return
	}

	resp := dto.TaskResponse{
		ID:             task.ID.String(),
		Title:          task.Title,
		Type:           string(task.Type),
		RecurrenceDays: []int(task.RecurrenceDays),
		IsDaily:        task.IsDaily,
		CreatedAt:      task.CreatedAt,
	}
	if task.TargetDate != nil {
		formatted := task.TargetDate.Format("2006-01-02")
		resp.TargetDate = &formatted
	}

	ctx.JSON(http.StatusCreated, resp)
}

// @Summary      List tasks
// @Description  Lists all tasks for the authenticated user, optionally filtered by type
// @Tags         Tasks
// @Produce      json
// @Param        type query string false "Filter by task type" Enums(recurring, check_in)
// @Success      200 {array} dto.TaskResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks [get]
func (c *TaskController) ListTasks(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var taskType *string
	if t := ctx.Query("type"); t != "" {
		taskType = &t
	}

	tasks, err := c.taskService.ListTasks(userID.(uuid.UUID), taskType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list tasks"})
		return
	}

	responses := make([]dto.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		resp := dto.TaskResponse{
			ID:             task.ID.String(),
			Title:          task.Title,
			Type:           string(task.Type),
			RecurrenceDays: []int(task.RecurrenceDays),
			IsDaily:        task.IsDaily,
			CreatedAt:      task.CreatedAt,
		}
		if task.TargetDate != nil {
			formatted := task.TargetDate.Format("2006-01-02")
			resp.TargetDate = &formatted
		}
		responses = append(responses, resp)
	}

	ctx.JSON(http.StatusOK, responses)
}

// @Summary      Get today's tasks
// @Description  Returns recurring tasks scheduled for today and today's check-ins with log status
// @Tags         Tasks
// @Produce      json
// @Success      200 {object} dto.TodayResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks/today [get]
func (c *TaskController) GetTodayTasks(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	response, err := c.taskService.GetTodayTasks(userID.(uuid.UUID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get today's tasks"})
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Get task
// @Description  Returns a single task by ID
// @Tags         Tasks
// @Produce      json
// @Param        id path string true "Task ID"
// @Success      200 {object} dto.TaskResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks/{id} [get]
func (c *TaskController) GetTask(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	taskID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid task ID"})
		return
	}

	task, err := c.taskService.GetTask(taskID, userID.(uuid.UUID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get task"})
		}
		return
	}

	resp := dto.TaskResponse{
		ID:             task.ID.String(),
		Title:          task.Title,
		Type:           string(task.Type),
		RecurrenceDays: []int(task.RecurrenceDays),
		IsDaily:        task.IsDaily,
		CreatedAt:      task.CreatedAt,
	}
	if task.TargetDate != nil {
		formatted := task.TargetDate.Format("2006-01-02")
		resp.TargetDate = &formatted
	}

	ctx.JSON(http.StatusOK, resp)
}

// @Summary      Update task
// @Description  Updates a task's title or recurrence schedule
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        id path string true "Task ID"
// @Param        request body dto.UpdateTaskRequest true "Fields to update"
// @Success      200 {object} dto.TaskResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks/{id} [put]
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	taskID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid task ID"})
		return
	}

	var req dto.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	task, err := c.taskService.UpdateTask(taskID, userID.(uuid.UUID), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrInvalidRecurrence):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update task"})
		}
		return
	}

	resp := dto.TaskResponse{
		ID:             task.ID.String(),
		Title:          task.Title,
		Type:           string(task.Type),
		RecurrenceDays: []int(task.RecurrenceDays),
		IsDaily:        task.IsDaily,
		CreatedAt:      task.CreatedAt,
	}
	if task.TargetDate != nil {
		formatted := task.TargetDate.Format("2006-01-02")
		resp.TargetDate = &formatted
	}

	ctx.JSON(http.StatusOK, resp)
}

// @Summary      Delete task
// @Description  Soft deletes a task
// @Tags         Tasks
// @Produce      json
// @Param        id path string true "Task ID"
// @Success      200 {object} dto.MessageResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks/{id} [delete]
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	taskID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid task ID"})
		return
	}

	err = c.taskService.DeleteTask(taskID, userID.(uuid.UUID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to delete task"})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.MessageResponse{Message: "task deleted successfully"})
}

// @Summary      Log task
// @Description  Creates or updates a rating for a task for today (upsert)
// @Tags         Task Logs
// @Accept       json
// @Produce      json
// @Param        id path string true "Task ID"
// @Param        request body dto.LogTaskRequest true "Log details"
// @Success      200 {object} dto.TaskLogResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks/{id}/logs [post]
func (c *TaskController) LogTask(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	taskID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid task ID"})
		return
	}

	var req dto.LogTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{Errors: shared.FormatValidationErrors(err)})
		return
	}

	log, err := c.taskService.LogTask(taskID, userID.(uuid.UUID), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, service.ErrInvalidRatingForType):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to log task"})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.TaskLogResponse{
		ID:        log.ID.String(),
		TaskID:    log.TaskID.String(),
		LogDate:   log.LogDate.Format("2006-01-02"),
		Rating:    string(log.Rating),
		Notes:     log.Notes,
		CreatedAt: log.CreatedAt,
	})
}

// @Summary      Get task logs
// @Description  Returns the full log history for a task
// @Tags         Task Logs
// @Produce      json
// @Param        id path string true "Task ID"
// @Success      200 {array} dto.TaskLogResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Security     BearerAuth
// @Router       /tasks/{id}/logs [get]
func (c *TaskController) GetTaskLogs(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	taskID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid task ID"})
		return
	}

	logs, err := c.taskService.GetTaskLogs(taskID, userID.(uuid.UUID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get task logs"})
		}
		return
	}

	responses := make([]dto.TaskLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, dto.TaskLogResponse{
			ID:        log.ID.String(),
			TaskID:    log.TaskID.String(),
			LogDate:   log.LogDate.Format("2006-01-02"),
			Rating:    string(log.Rating),
			Notes:     log.Notes,
			CreatedAt: log.CreatedAt,
		})
	}

	ctx.JSON(http.StatusOK, responses)
}
