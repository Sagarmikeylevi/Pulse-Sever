package service_test

import (
	"testing"
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/dto"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/mocks"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newTestTaskService() (service.TaskService, *mocks.MockTaskRepository, *mocks.MockTaskLogRepository, *mocks.MockUserRepository) {
	taskRepo := new(mocks.MockTaskRepository)
	taskLogRepo := new(mocks.MockTaskLogRepository)
	userRepo := new(mocks.MockUserRepository)
	taskSvc := service.NewTaskService(taskRepo, taskLogRepo, userRepo)
	return taskSvc, taskRepo, taskLogRepo, userRepo
}

var testUserID = uuid.New()
var testUser = &entity.User{
	Email:    "sagar@test.com",
	Timezone: "Asia/Kolkata",
}

func init() {
	testUser.ID = testUserID
}

// ==================== CreateTask Tests ====================

func TestCreateTask_RecurringWithDays(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskRepo.On("Create", mock.AnythingOfType("*entity.Task")).Return(nil)

	req := dto.CreateTaskRequest{
		Title:          "Go to gym",
		Type:           "recurring",
		RecurrenceDays: []int{1, 3, 5},
	}

	task, err := taskSvc.CreateTask(testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, "Go to gym", task.Title)
	assert.Equal(t, entity.TaskTypeRecurring, task.Type)
	assert.Equal(t, entity.RecurrenceDays{1, 3, 5}, task.RecurrenceDays)
	assert.False(t, task.IsDaily)
	assert.Nil(t, task.TargetDate)
}

func TestCreateTask_RecurringDaily(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskRepo.On("Create", mock.AnythingOfType("*entity.Task")).Return(nil)

	req := dto.CreateTaskRequest{
		Title:   "Meditate",
		Type:    "recurring",
		IsDaily: true,
	}

	task, err := taskSvc.CreateTask(testUserID, req)

	assert.NoError(t, err)
	assert.True(t, task.IsDaily)
	assert.Equal(t, entity.RecurrenceDays{0, 1, 2, 3, 4, 5, 6}, task.RecurrenceDays)
}

func TestCreateTask_RecurringAllSevenDays_SetsIsDaily(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskRepo.On("Create", mock.AnythingOfType("*entity.Task")).Return(nil)

	req := dto.CreateTaskRequest{
		Title:          "Read",
		Type:           "recurring",
		RecurrenceDays: []int{0, 1, 2, 3, 4, 5, 6},
	}

	task, err := taskSvc.CreateTask(testUserID, req)

	assert.NoError(t, err)
	assert.True(t, task.IsDaily)
}

func TestCreateTask_RecurringNoRecurrence_Error(t *testing.T) {
	taskSvc, _, _, _ := newTestTaskService()

	req := dto.CreateTaskRequest{
		Title: "Read",
		Type:  "recurring",
	}

	task, err := taskSvc.CreateTask(testUserID, req)

	assert.Nil(t, task)
	assert.ErrorIs(t, err, service.ErrRecurrenceRequired)
}

func TestCreateTask_RecurringInvalidDay_Error(t *testing.T) {
	taskSvc, _, _, _ := newTestTaskService()

	req := dto.CreateTaskRequest{
		Title:          "Read",
		Type:           "recurring",
		RecurrenceDays: []int{1, 7},
	}

	task, err := taskSvc.CreateTask(testUserID, req)

	assert.Nil(t, task)
	assert.ErrorIs(t, err, service.ErrInvalidRecurrence)
}

func TestCreateTask_RecurringDuplicateDays_Error(t *testing.T) {
	taskSvc, _, _, _ := newTestTaskService()

	req := dto.CreateTaskRequest{
		Title:          "Read",
		Type:           "recurring",
		RecurrenceDays: []int{1, 3, 3},
	}

	task, err := taskSvc.CreateTask(testUserID, req)

	assert.Nil(t, task)
	assert.ErrorIs(t, err, service.ErrInvalidRecurrence)
}

func TestCreateTask_CheckIn(t *testing.T) {
	taskSvc, taskRepo, _, userRepo := newTestTaskService()

	userRepo.On("FindByID", testUserID).Return(testUser, nil)
	taskRepo.On("Create", mock.AnythingOfType("*entity.Task")).Return(nil)

	req := dto.CreateTaskRequest{
		Title: "Buy groceries",
		Type:  "check_in",
	}

	task, err := taskSvc.CreateTask(testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.TaskTypeCheckIn, task.Type)
	assert.NotNil(t, task.TargetDate)
	assert.Nil(t, task.RecurrenceDays)
}

// ==================== GetTask Tests ====================

func TestGetTask_Success(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	existingTask := &entity.Task{
		Title: "Gym",
		Type:  entity.TaskTypeRecurring,
	}
	existingTask.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(existingTask, nil)

	task, err := taskSvc.GetTask(taskID, testUserID)

	assert.NoError(t, err)
	assert.Equal(t, "Gym", task.Title)
}

func TestGetTask_NotFound(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	taskRepo.On("FindByID", taskID, testUserID).Return(nil, gorm.ErrRecordNotFound)

	task, err := taskSvc.GetTask(taskID, testUserID)

	assert.Nil(t, task)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

// ==================== UpdateTask Tests ====================

func TestUpdateTask_Title(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	existingTask := &entity.Task{
		Title: "Gym",
		Type:  entity.TaskTypeRecurring,
	}
	existingTask.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(existingTask, nil)
	taskRepo.On("Update", existingTask).Return(nil)

	newTitle := "Go to gym"
	req := dto.UpdateTaskRequest{Title: &newTitle}

	task, err := taskSvc.UpdateTask(taskID, testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, "Go to gym", task.Title)
}

func TestUpdateTask_RecurrenceDays(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	existingTask := &entity.Task{
		Title:          "Gym",
		Type:           entity.TaskTypeRecurring,
		RecurrenceDays: entity.RecurrenceDays{1, 3, 5},
	}
	existingTask.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(existingTask, nil)
	taskRepo.On("Update", existingTask).Return(nil)

	req := dto.UpdateTaskRequest{RecurrenceDays: []int{1, 2, 3, 4, 5}}

	task, err := taskSvc.UpdateTask(taskID, testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.RecurrenceDays{1, 2, 3, 4, 5}, task.RecurrenceDays)
	assert.False(t, task.IsDaily)
}

func TestUpdateTask_NotFound(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	taskRepo.On("FindByID", taskID, testUserID).Return(nil, gorm.ErrRecordNotFound)

	req := dto.UpdateTaskRequest{}
	task, err := taskSvc.UpdateTask(taskID, testUserID, req)

	assert.Nil(t, task)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

// ==================== DeleteTask Tests ====================

func TestDeleteTask_Success(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	existingTask := &entity.Task{Title: "Gym"}
	existingTask.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(existingTask, nil)
	taskRepo.On("Delete", taskID, testUserID).Return(nil)

	err := taskSvc.DeleteTask(taskID, testUserID)

	assert.NoError(t, err)
	taskRepo.AssertCalled(t, "Delete", taskID, testUserID)
}

func TestDeleteTask_NotFound(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	taskRepo.On("FindByID", taskID, testUserID).Return(nil, gorm.ErrRecordNotFound)

	err := taskSvc.DeleteTask(taskID, testUserID)

	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

// ==================== LogTask Tests ====================

func TestLogTask_RecurringGood(t *testing.T) {
	taskSvc, taskRepo, taskLogRepo, userRepo := newTestTaskService()

	taskID := uuid.New()
	task := &entity.Task{Type: entity.TaskTypeRecurring}
	task.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(task, nil)
	userRepo.On("FindByID", testUserID).Return(testUser, nil)
	taskLogRepo.On("Upsert", mock.AnythingOfType("*entity.TaskLog")).Return(nil)

	req := dto.LogTaskRequest{Rating: "good"}

	log, err := taskSvc.LogTask(taskID, testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.RatingGood, log.Rating)
	assert.NotNil(t, log.Score)
	assert.Equal(t, 63, *log.Score)
}

func TestLogTask_RecurringRest_NilScore(t *testing.T) {
	taskSvc, taskRepo, taskLogRepo, userRepo := newTestTaskService()

	taskID := uuid.New()
	task := &entity.Task{Type: entity.TaskTypeRecurring}
	task.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(task, nil)
	userRepo.On("FindByID", testUserID).Return(testUser, nil)
	taskLogRepo.On("Upsert", mock.AnythingOfType("*entity.TaskLog")).Return(nil)

	req := dto.LogTaskRequest{Rating: "rest"}

	log, err := taskSvc.LogTask(taskID, testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.RatingRest, log.Rating)
	assert.Nil(t, log.Score)
}

func TestLogTask_CheckInDone(t *testing.T) {
	taskSvc, taskRepo, taskLogRepo, userRepo := newTestTaskService()

	taskID := uuid.New()
	task := &entity.Task{Type: entity.TaskTypeCheckIn}
	task.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(task, nil)
	userRepo.On("FindByID", testUserID).Return(testUser, nil)
	taskLogRepo.On("Upsert", mock.AnythingOfType("*entity.TaskLog")).Return(nil)

	req := dto.LogTaskRequest{Rating: "done"}

	log, err := taskSvc.LogTask(taskID, testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.RatingDone, log.Rating)
	assert.Equal(t, 100, *log.Score)
}

func TestLogTask_CheckInWithRecurringRating_Error(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	task := &entity.Task{Type: entity.TaskTypeCheckIn}
	task.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(task, nil)

	req := dto.LogTaskRequest{Rating: "good"}

	log, err := taskSvc.LogTask(taskID, testUserID, req)

	assert.Nil(t, log)
	assert.ErrorIs(t, err, service.ErrInvalidRatingForType)
}

func TestLogTask_RecurringWithDoneRating_Error(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	task := &entity.Task{Type: entity.TaskTypeRecurring}
	task.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(task, nil)

	req := dto.LogTaskRequest{Rating: "done"}

	log, err := taskSvc.LogTask(taskID, testUserID, req)

	assert.Nil(t, log)
	assert.ErrorIs(t, err, service.ErrInvalidRatingForType)
}

func TestLogTask_TaskNotFound_Error(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	taskRepo.On("FindByID", taskID, testUserID).Return(nil, gorm.ErrRecordNotFound)

	req := dto.LogTaskRequest{Rating: "good"}

	log, err := taskSvc.LogTask(taskID, testUserID, req)

	assert.Nil(t, log)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

func TestLogTask_WithNotes(t *testing.T) {
	taskSvc, taskRepo, taskLogRepo, userRepo := newTestTaskService()

	taskID := uuid.New()
	task := &entity.Task{Type: entity.TaskTypeRecurring}
	task.ID = taskID

	taskRepo.On("FindByID", taskID, testUserID).Return(task, nil)
	userRepo.On("FindByID", testUserID).Return(testUser, nil)
	taskLogRepo.On("Upsert", mock.AnythingOfType("*entity.TaskLog")).Return(nil)

	notes := "Hit all my sets today"
	req := dto.LogTaskRequest{Rating: "crushed_it", Notes: &notes}

	log, err := taskSvc.LogTask(taskID, testUserID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.RatingCrushedIt, log.Rating)
	assert.Equal(t, 88, *log.Score)
	assert.Equal(t, "Hit all my sets today", *log.Notes)
}

// ==================== GetTaskLogs Tests ====================

func TestGetTaskLogs_Success(t *testing.T) {
	taskSvc, taskRepo, taskLogRepo, _ := newTestTaskService()

	taskID := uuid.New()
	task := &entity.Task{Title: "Gym"}
	task.ID = taskID

	score := 65
	logs := []entity.TaskLog{
		{Rating: entity.RatingGood, Score: &score, LogDate: time.Now()},
	}

	taskRepo.On("FindByID", taskID, testUserID).Return(task, nil)
	taskLogRepo.On("FindByTaskID", taskID).Return(logs, nil)

	result, err := taskSvc.GetTaskLogs(taskID, testUserID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetTaskLogs_TaskNotFound(t *testing.T) {
	taskSvc, taskRepo, _, _ := newTestTaskService()

	taskID := uuid.New()
	taskRepo.On("FindByID", taskID, testUserID).Return(nil, gorm.ErrRecordNotFound)

	result, err := taskSvc.GetTaskLogs(taskID, testUserID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, service.ErrTaskNotFound)
}

// ==================== GetTodayTasks Tests ====================

func TestGetTodayTasks_Success(t *testing.T) {
	taskSvc, taskRepo, taskLogRepo, userRepo := newTestTaskService()

	userRepo.On("FindByID", testUserID).Return(testUser, nil)

	recurringTask := entity.Task{
		Title:          "Gym",
		Type:           entity.TaskTypeRecurring,
		IsDaily:        true,
		RecurrenceDays: entity.RecurrenceDays{0, 1, 2, 3, 4, 5, 6},
	}
	recurringTask.ID = uuid.New()

	taskRepo.On("FindRecurringByUserID", testUserID).Return([]entity.Task{recurringTask}, nil)
	taskRepo.On("FindCheckInsByDate", testUserID, mock.AnythingOfType("time.Time")).Return([]entity.Task{}, nil)
	taskLogRepo.On("FindByUserAndDate", testUserID, mock.AnythingOfType("time.Time")).Return([]entity.TaskLog{}, nil)

	response, err := taskSvc.GetTodayTasks(testUserID)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Date)
	assert.Len(t, response.RecurringTasks, 1)
	assert.Len(t, response.CheckIns, 0)
	assert.Nil(t, response.RecurringTasks[0].Log)
}

func TestGetTodayTasks_WithLog(t *testing.T) {
	taskSvc, taskRepo, taskLogRepo, userRepo := newTestTaskService()

	userRepo.On("FindByID", testUserID).Return(testUser, nil)

	taskID := uuid.New()
	recurringTask := entity.Task{
		Title:          "Gym",
		Type:           entity.TaskTypeRecurring,
		IsDaily:        true,
		RecurrenceDays: entity.RecurrenceDays{0, 1, 2, 3, 4, 5, 6},
	}
	recurringTask.ID = taskID

	score := 90
	logEntry := entity.TaskLog{
		TaskID: taskID,
		UserID: testUserID,
		Rating: entity.RatingCrushedIt,
		Score:  &score,
	}
	logEntry.ID = uuid.New()

	taskRepo.On("FindRecurringByUserID", testUserID).Return([]entity.Task{recurringTask}, nil)
	taskRepo.On("FindCheckInsByDate", testUserID, mock.AnythingOfType("time.Time")).Return([]entity.Task{}, nil)
	taskLogRepo.On("FindByUserAndDate", testUserID, mock.AnythingOfType("time.Time")).Return([]entity.TaskLog{logEntry}, nil)

	response, err := taskSvc.GetTodayTasks(testUserID)

	assert.NoError(t, err)
	assert.Len(t, response.RecurringTasks, 1)
	assert.NotNil(t, response.RecurringTasks[0].Log)
	assert.Equal(t, "crushed_it", response.RecurringTasks[0].Log.Rating)
}

func TestGetTodayTasks_UserNotFound(t *testing.T) {
	taskSvc, _, _, userRepo := newTestTaskService()

	userRepo.On("FindByID", testUserID).Return(nil, gorm.ErrRecordNotFound)

	response, err := taskSvc.GetTodayTasks(testUserID)

	assert.Nil(t, response)
	assert.ErrorIs(t, err, service.ErrUserNotFound)
}

// ==================== ScoreFromRating Tests ====================

func TestScoreFromRating(t *testing.T) {
	tests := []struct {
		rating   entity.Rating
		expected *int
	}{
		{entity.RatingRest, nil},
		{entity.RatingRough, intPtr(13)},
		{entity.RatingOkay, intPtr(38)},
		{entity.RatingGood, intPtr(63)},
		{entity.RatingCrushedIt, intPtr(88)},
		{entity.RatingDone, intPtr(100)},
	}

	for _, tt := range tests {
		t.Run(string(tt.rating), func(t *testing.T) {
			score := entity.ScoreFromRating(tt.rating)
			if tt.expected == nil {
				assert.Nil(t, score)
			} else {
				assert.Equal(t, *tt.expected, *score)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
