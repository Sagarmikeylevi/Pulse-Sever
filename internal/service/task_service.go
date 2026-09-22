package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/dto"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTaskNotFound         = errors.New("task not found")
	ErrInvalidRecurrence    = errors.New("recurrence_days must contain values between 0 (Sunday) and 6 (Saturday)")
	ErrRecurrenceRequired   = errors.New("recurring tasks must have recurrence_days or is_daily set")
	ErrInvalidRatingForType = errors.New("invalid rating for this task type")
)

type TaskService interface {
	CreateTask(userID uuid.UUID, req dto.CreateTaskRequest) (*entity.Task, error)
	GetTask(id, userID uuid.UUID) (*entity.Task, error)
	ListTasks(userID uuid.UUID, taskType *string) ([]entity.Task, error)
	UpdateTask(id, userID uuid.UUID, req dto.UpdateTaskRequest) (*entity.Task, error)
	DeleteTask(id, userID uuid.UUID) error

	LogTask(taskID, userID uuid.UUID, req dto.LogTaskRequest) (*entity.TaskLog, error)
	GetTaskLogs(taskID, userID uuid.UUID) ([]entity.TaskLog, error)

	GetTodayTasks(userID uuid.UUID) (*dto.TodayResponse, error)
}

type taskService struct {
	taskRepo    repository.TaskRepository
	taskLogRepo repository.TaskLogRepository
	userRepo    repository.UserRepository
}

func NewTaskService(
	taskRepo repository.TaskRepository,
	taskLogRepo repository.TaskLogRepository,
	userRepo repository.UserRepository,
) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		taskLogRepo: taskLogRepo,
		userRepo:    userRepo,
	}
}

func (s *taskService) CreateTask(userID uuid.UUID, req dto.CreateTaskRequest) (*entity.Task, error) {
	task := &entity.Task{
		UserID: userID,
		Title:  req.Title,
		Type:   entity.TaskType(req.Type),
	}

	if task.Type == entity.TaskTypeRecurring {
		if req.IsDaily {
			task.IsDaily = true
			task.RecurrenceDays = entity.RecurrenceDays{0, 1, 2, 3, 4, 5, 6}
		} else {
			if len(req.RecurrenceDays) == 0 {
				return nil, ErrRecurrenceRequired
			}
			if err := validateRecurrenceDays(req.RecurrenceDays); err != nil {
				return nil, err
			}
			task.RecurrenceDays = req.RecurrenceDays
			if len(req.RecurrenceDays) == 7 {
				task.IsDaily = true
			}
		}
	}

	if task.Type == entity.TaskTypeCheckIn {
		today, err := s.getUserToday(userID)
		if err != nil {
			return nil, err
		}
		task.TargetDate = &today
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

func (s *taskService) GetTask(id, userID uuid.UUID) (*entity.Task, error) {
	task, err := s.taskRepo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return task, nil
}

func (s *taskService) ListTasks(userID uuid.UUID, taskType *string) ([]entity.Task, error) {
	return s.taskRepo.FindByUserID(userID, taskType)
}

func (s *taskService) UpdateTask(id, userID uuid.UUID, req dto.UpdateTaskRequest) (*entity.Task, error) {
	task, err := s.taskRepo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		task.Title = *req.Title
	}

	// Only recurring tasks can update recurrence
	if task.Type == entity.TaskTypeRecurring {
		if req.IsDaily != nil && *req.IsDaily {
			task.IsDaily = true
			task.RecurrenceDays = entity.RecurrenceDays{0, 1, 2, 3, 4, 5, 6}
		} else if len(req.RecurrenceDays) > 0 {
			if err := validateRecurrenceDays(req.RecurrenceDays); err != nil {
				return nil, err
			}
			task.RecurrenceDays = req.RecurrenceDays
			task.IsDaily = len(req.RecurrenceDays) == 7
		}
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return task, nil
}

func (s *taskService) DeleteTask(id, userID uuid.UUID) error {
	_, err := s.taskRepo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}
	return s.taskRepo.Delete(id, userID)
}

// --- Task Logging ---

// LogTask creates or updates a log for today using a single DB upsert.
func (s *taskService) LogTask(taskID, userID uuid.UUID, req dto.LogTaskRequest) (*entity.TaskLog, error) {
	task, err := s.taskRepo.FindByID(taskID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	rating := entity.Rating(req.Rating)
	if err := validateRatingForTaskType(rating, task.Type); err != nil {
		return nil, err
	}

	today, err := s.getUserToday(userID)
	if err != nil {
		return nil, err
	}

	log := &entity.TaskLog{
		TaskID:  taskID,
		UserID:  userID,
		LogDate: today,
		Rating:  rating,
		Score:   entity.ScoreFromRating(rating),
		Notes:   req.Notes,
	}

	if err := s.taskLogRepo.Upsert(log); err != nil {
		return nil, fmt.Errorf("failed to log task: %w", err)
	}
	return log, nil
}

func (s *taskService) GetTaskLogs(taskID, userID uuid.UUID) ([]entity.TaskLog, error) {
	// Verify user owns this task
	_, err := s.taskRepo.FindByID(taskID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return s.taskLogRepo.FindByTaskID(taskID)
}

// --- Today View ---

func (s *taskService) GetTodayTasks(userID uuid.UUID) (*dto.TodayResponse, error) {
	today, err := s.getUserToday(userID)
	if err != nil {
		return nil, err
	}

	dayOfWeek := int(today.Weekday())

	// Fetch recurring tasks and filter by today's day
	recurringTasks, err := s.taskRepo.FindRecurringByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recurring tasks: %w", err)
	}

	var todayRecurring []entity.Task
	for _, t := range recurringTasks {
		if t.IsDaily || containsDay(t.RecurrenceDays, dayOfWeek) {
			todayRecurring = append(todayRecurring, t)
		}
	}

	// Fetch today's check-ins
	checkIns, err := s.taskRepo.FindCheckInsByDate(userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch check-ins: %w", err)
	}

	// Fetch all logs for today
	todayLogs, err := s.taskLogRepo.FindByUserAndDate(userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch today's logs: %w", err)
	}

	// Index logs by task ID for quick lookup
	logsByTaskID := make(map[uuid.UUID]*entity.TaskLog)
	for i := range todayLogs {
		logsByTaskID[todayLogs[i].TaskID] = &todayLogs[i]
	}

	// Build response
	dateStr := today.Format("2006-01-02")
	response := &dto.TodayResponse{
		Date:           dateStr,
		RecurringTasks: buildTaskWithLogResponses(todayRecurring, logsByTaskID),
		CheckIns:       buildTaskWithLogResponses(checkIns, logsByTaskID),
	}

	return response, nil
}

// --- Helpers ---

func validateRecurrenceDays(days []int) error {
	seen := make(map[int]bool)
	for _, d := range days {
		if d < 0 || d > 6 {
			return ErrInvalidRecurrence
		}
		if seen[d] {
			return ErrInvalidRecurrence
		}
		seen[d] = true
	}
	return nil
}

func validateRatingForTaskType(rating entity.Rating, taskType entity.TaskType) error {
	if taskType == entity.TaskTypeCheckIn {
		if rating != entity.RatingDone {
			return ErrInvalidRatingForType
		}
		return nil
	}
	// Recurring tasks accept all ratings except "done"
	if rating == entity.RatingDone {
		return ErrInvalidRatingForType
	}
	return nil
}

func containsDay(days entity.RecurrenceDays, day int) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

func buildTaskWithLogResponses(tasks []entity.Task, logsByTaskID map[uuid.UUID]*entity.TaskLog) []dto.TaskWithLogResponse {
	responses := make([]dto.TaskWithLogResponse, 0, len(tasks))
	for _, t := range tasks {
		taskResp := dto.TaskResponse{
			ID:             t.ID.String(),
			Title:          t.Title,
			Type:           string(t.Type),
			RecurrenceDays: []int(t.RecurrenceDays),
			IsDaily:        t.IsDaily,
			CreatedAt:      t.CreatedAt,
		}
		if t.TargetDate != nil {
			formatted := t.TargetDate.Format("2006-01-02")
			taskResp.TargetDate = &formatted
		}

		result := dto.TaskWithLogResponse{TaskResponse: taskResp}
		if log, ok := logsByTaskID[t.ID]; ok {
			result.Log = &dto.TaskLogResponse{
				ID:        log.ID.String(),
				TaskID:    log.TaskID.String(),
				LogDate:   log.LogDate.Format("2006-01-02"),
				Rating:    string(log.Rating),
				Notes:     log.Notes,
				CreatedAt: log.CreatedAt,
			}
		}
		responses = append(responses, result)
	}
	return responses
}

// getUserToday returns today's date in the user's timezone.
func (s *taskService) getUserToday(userID uuid.UUID) (time.Time, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return time.Time{}, ErrUserNotFound
	}
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid stored timezone: %w", err)
	}
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return today, nil
}
