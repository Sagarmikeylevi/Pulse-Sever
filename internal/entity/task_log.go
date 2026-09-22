package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Rating string

const (
	RatingRest      Rating = "rest"
	RatingRough     Rating = "rough"
	RatingOkay      Rating = "okay"
	RatingGood      Rating = "good"
	RatingCrushedIt Rating = "crushed_it"
	RatingDone      Rating = "done"
)

type TaskLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	LogDate   time.Time `gorm:"type:date;not null"`
	Rating    Rating    `gorm:"type:varchar(20);not null"`
	Score     *int      `gorm:"type:smallint"`
	Notes     *string   `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (t *TaskLog) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// ScoreFromRating maps a rating to its internal score value.
func ScoreFromRating(rating Rating) *int {
	scores := map[Rating]int{
		RatingRough:     13,
		RatingOkay:      38,
		RatingGood:      63,
		RatingCrushedIt: 88,
		RatingDone:      100,
	}
	if score, ok := scores[rating]; ok {
		return &score
	}
	return nil // rest -> no score
}
