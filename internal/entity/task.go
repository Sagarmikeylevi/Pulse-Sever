package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskTypeRecurring TaskType = "recurring"
	TaskTypeCheckIn   TaskType = "check_in"
)

// RecurrenceDays is a custom type that maps JSONB <-> []int for GORM.
type RecurrenceDays []int

func (r RecurrenceDays) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

func (r *RecurrenceDays) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan RecurrenceDays: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, r)
}

type Task struct {
	BaseModel
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index"`
	Title          string         `gorm:"type:varchar(255);not null"`
	Type           TaskType       `gorm:"type:varchar(20);not null"`
	RecurrenceDays RecurrenceDays `gorm:"type:jsonb"`
	IsDaily        bool           `gorm:"not null;default:false"`
	TargetDate     *time.Time     `gorm:"type:date"`
	Logs           []TaskLog      `gorm:"foreignKey:TaskID"`
}
