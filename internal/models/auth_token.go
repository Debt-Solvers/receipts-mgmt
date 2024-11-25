package models

import (
	"time"

	"github.com/google/uuid"
)

type AuthToken struct {
	TokenID   uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"` // Primary key
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Token     string    `gorm:"not null;unique"` // Token as a string
	CreatedAt time.Time
	ExpiresAt time.Time
}

