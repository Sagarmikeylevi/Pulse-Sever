package repository

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(email string) (*entity.User, error)
	FindByID(id uuid.UUID) (*entity.User, error)
	Create(user *entity.User) error
	UpdatePassword(id uuid.UUID, passwordHash string) error
	MarkEmailVerified(id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UpdatePassword(id uuid.UUID, passwordHash string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

func (r *userRepository) MarkEmailVerified(id uuid.UUID) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("is_email_verified", true).Error
}
