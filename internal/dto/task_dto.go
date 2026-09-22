package dto

import "time"

// Requests

type CreateTaskRequest struct {
	Title          string `json:"title" binding:"required,min=1,max=255"`
	Type           string `json:"type" binding:"required,oneof=recurring check_in"`
	RecurrenceDays []int  `json:"recurrence_days,omitempty"`
	IsDaily        bool   `json:"is_daily,omitempty"`
}

type UpdateTaskRequest struct {
	Title          *string `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
	RecurrenceDays []int   `json:"recurrence_days,omitempty"`
	IsDaily        *bool   `json:"is_daily,omitempty"`
}

type LogTaskRequest struct {
	Rating string  `json:"rating" binding:"required,oneof=rest rough okay good crushed_it done"`
	Notes  *string `json:"notes,omitempty"`
}

// Responses

type TaskResponse struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Type           string     `json:"type"`
	RecurrenceDays []int      `json:"recurrence_days,omitempty"`
	IsDaily        bool       `json:"is_daily"`
	TargetDate     *string    `json:"target_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type TaskLogResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	LogDate   string    `json:"log_date"`
	Rating    string    `json:"rating"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskWithLogResponse struct {
	TaskResponse
	Log *TaskLogResponse `json:"log"`
}

type TodayResponse struct {
	Date           string                `json:"date"`
	RecurringTasks []TaskWithLogResponse `json:"recurring_tasks"`
	CheckIns       []TaskWithLogResponse `json:"check_ins"`
}
