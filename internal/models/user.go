package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"user_id"`
	FirstName         string    `gorm:"not null" json:"first_name"`
	LastName          string    `gorm:"not null" json:"last_name"`
	Email             string    `gorm:"unique;not null" json:"email"`
	PasswordHash      string    `gorm:"not null" json:"password"` // Change json tag to "password"
	Salt              string    `gorm:"not null" json:"-"`
	IsEmailVerified   bool      `gorm:"default:false" json:"is_email_verified"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	ResetPasswordToken string    `gorm:"size:255" json:"-"`
	ResetPasswordExpires time.Time `json:"reset_password_expires"`
	Currency          string    `gorm:"type:char(3);default:CAD;check:currency in ('CAD', 'USD')" json:"currency"`
}
