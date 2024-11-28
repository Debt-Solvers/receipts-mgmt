package common

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
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

// Function to generate SHA-256 hash of file content
func GenerateFileHash(file io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
			return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}