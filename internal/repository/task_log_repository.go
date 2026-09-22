package repository

import (
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TaskLogRepository interface {
	Upsert(log *entity.TaskLog) error
	FindByUserAndDate(userID uuid.UUID, date time.Time) ([]entity.TaskLog, error)
	FindByTaskID(taskID uuid.UUID) ([]entity.TaskLog, error)
}

type taskLogRepository struct {
	db *gorm.DB
}

func NewTaskLogRepository(db *gorm.DB) TaskLogRepository {
	return &taskLogRepository{db: db}
}

func (r *taskLogRepository) Upsert(log *entity.TaskLog) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "task_id"}, {Name: "log_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"rating", "score", "notes", "updated_at"}),
	}).Create(log).Error
}

func (r *taskLogRepository) FindByUserAndDate(userID uuid.UUID, date time.Time) ([]entity.TaskLog, error) {
	var logs []entity.TaskLog
	err := r.db.Where("user_id = ? AND log_date = ?", userID, date).Find(&logs).Error
	return logs, err
}

func (r *taskLogRepository) FindByTaskID(taskID uuid.UUID) ([]entity.TaskLog, error) {
	var logs []entity.TaskLog
	err := r.db.Where("task_id = ?", taskID).Order("log_date DESC").Find(&logs).Error
	return logs, err
}
