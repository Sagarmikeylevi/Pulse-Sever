package entity

type User struct {
	BaseModel
	Email           string  `gorm:"uniqueIndex;type:varchar(255);not null"`
	PasswordHash    *string `gorm:"type:varchar(255)"`
	IsEmailVerified bool    `gorm:"not null;default:false"`
}
