package common

import (
	"receipt-mgmt/db"
	"receipt-mgmt/internal/models"
)

// Checks if token is valid and in database
func IsTokenActive(token string) bool {
	// Get the DB instance
	db := db.GetDBInstance()

	err := db.Where("token = ?", token).First(&models.AuthToken{}).Error
	return err == nil
}
