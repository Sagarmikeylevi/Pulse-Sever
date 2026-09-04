package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTPStatus string

const (
	OTPStatusActive  OTPStatus = "active"
	OTPStatusUsed    OTPStatus = "used"
	OTPStatusExpired OTPStatus = "expired"
)

type OTP struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"type:varchar(255);not null;index"`
	CodeHash  string    `gorm:"type:varchar(255);not null"`
	Attempts  int       `gorm:"not null;default:0"`
	Status    OTPStatus `gorm:"type:varchar(20);not null;default:'active'"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (o *OTP) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
