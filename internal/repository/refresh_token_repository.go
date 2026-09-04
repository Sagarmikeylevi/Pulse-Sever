package repository

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *entity.RefreshToken) error
	FindByTokenHash(tokenHash string) (*entity.RefreshToken, error)
	MarkAsUsed(token *entity.RefreshToken) error
	RevokeFamily(familyID uuid.UUID) error
	RevokeAllForUser(userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *entity.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) MarkAsUsed(token *entity.RefreshToken) error {
	return r.db.Model(token).Update("status", entity.TokenStatusUsed).Error
}

func (r *refreshTokenRepository) RevokeFamily(familyID uuid.UUID) error {
	return r.db.Model(&entity.RefreshToken{}).
		Where("family_id = ?", familyID).
		Update("status", entity.TokenStatusRevoked).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	return r.db.Model(&entity.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("status", entity.TokenStatusRevoked).Error
}
