package repository

import (
	"github.com/Sagarmikeylevi/Pulse-Sever/internal/entity"
	"gorm.io/gorm"
)

type OTPRepository interface {
	Create(otp *entity.OTP) error
	FindLatestActiveByEmail(email string) (*entity.OTP, error)
	IncrementAttempts(otp *entity.OTP) error
	MarkAsUsed(otp *entity.OTP) error
}

type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) Create(otp *entity.OTP) error {
	return r.db.Create(otp).Error
}

func (r *otpRepository) FindLatestActiveByEmail(email string) (*entity.OTP, error) {
	var otp entity.OTP
	err := r.db.Where("email = ? AND status = ?", email, entity.OTPStatusActive).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) IncrementAttempts(otp *entity.OTP) error {
	return r.db.Model(otp).Update("attempts", gorm.Expr("attempts + 1")).Error
}

func (r *otpRepository) MarkAsUsed(otp *entity.OTP) error {
	return r.db.Model(otp).Update("status", entity.OTPStatusUsed).Error
}
