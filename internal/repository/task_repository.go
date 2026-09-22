package repository

import (
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *entity.Task) error
	FindByID(id, userID uuid.UUID) (*entity.Task, error)
	FindByUserID(userID uuid.UUID, taskType *string) ([]entity.Task, error)
	FindRecurringByUserID(userID uuid.UUID) ([]entity.Task, error)
	FindCheckInsByDate(userID uuid.UUID, date time.Time) ([]entity.Task, error)
	Update(task *entity.Task) error
	Delete(id, userID uuid.UUID) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(task *entity.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) FindByID(id, userID uuid.UUID) (*entity.Task, error) {
	var task entity.Task
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) FindByUserID(userID uuid.UUID, taskType *string) ([]entity.Task, error) {
	var tasks []entity.Task
	query := r.db.Where("user_id = ?", userID)
	if taskType != nil {
		query = query.Where("type = ?", *taskType)
	}
	err := query.Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) FindRecurringByUserID(userID uuid.UUID) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.Where("user_id = ? AND type = ?", userID, entity.TaskTypeRecurring).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) FindCheckInsByDate(userID uuid.UUID, date time.Time) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.Where("user_id = ? AND type = ? AND target_date = ?", userID, entity.TaskTypeCheckIn, date).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) Update(task *entity.Task) error {
	return r.db.Save(task).Error
}

func (r *taskRepository) Delete(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.Task{}).Error
}
