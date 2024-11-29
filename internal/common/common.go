package common

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
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

// GenerateFileHash generates a SHA-256 hash for the given file
func GenerateFileHash(file io.Reader) (string, error) {
	// Create a new SHA-256 hash
	hasher := sha256.New()
	// Copy the file contents into the hasher
	if _, err := io.Copy(hasher, file); err != nil {
		log.Println("Error generating file hash:", err)
		return "", err
	}
	// Compute the hash
	hash := hasher.Sum(nil)
	// Return the hash as a hexadecimal string
	return hex.EncodeToString(hash), nil
}