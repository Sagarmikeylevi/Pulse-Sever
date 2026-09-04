package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "active"
	TokenStatusUsed    TokenStatus = "used"
	TokenStatusRevoked TokenStatus = "revoked"
)

type RefreshToken struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null;index"`
	TokenHash string      `gorm:"type:varchar(255);not null;index"`
	FamilyID  uuid.UUID   `gorm:"type:uuid;not null;index"`
	Status    TokenStatus `gorm:"type:varchar(20);not null;default:'active'"`
	ExpiresAt time.Time   `gorm:"not null"`
	CreatedAt time.Time   `gorm:"not null"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}
