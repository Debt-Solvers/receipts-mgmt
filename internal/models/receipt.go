package models

import (
	"encoding/json"
	"log"
	"receipt-mgmt/db"
	"time"

	"github.com/google/uuid"
)

// Receipt represents the receipt model with its associated fields.
type Receipt struct {
	ReceiptID       uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"receipt_id"` // UUID for receipt ID
	UserID          uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"` 
	Image          []byte    `gorm:"type:bytea;not null" json:"image"` // Changed to hold the actual image as bytea
	Status         string    `gorm:"type:varchar(50);not null" json:"status"` // Values: pending, processing, completed, failed
	TotalAmount    float64   `gorm:"type:decimal(10,2)" json:"total_amount"`
	Merchant       string    `gorm:"type:varchar(255)" json:"merchant"`
	Items          json.RawMessage    `gorm:"type:jsonb" json:"items"` // Store items as JSON
	ScannedDate    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"scanned_date"`
	TransactionDate string    `gorm:"type:varchar(50);not null" json:"transaction_date"`
	TransactionTime string   `gorm:"type:varchar(50);not null" json:"transaction_time"`
  FileHash        string   `gorm:"type:varchar(64);unique;not null" json:"file_hash"`
	Tax            float64   `gorm:"type:decimal(10,2)" json:"tax"`
	Discounts      float64   `gorm:"type:decimal(10,2)" json:"discounts"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// CheckFileHashExists checks if a receipt with the given file hash already exists in the database
func CheckFileHashExists(fileHash string) (bool, error) {
  // Use GetDBInstance to get the DB instance
	DB:= db.GetDBInstance()
	var count int64
	// Query the database for receipts with the same file_hash
	if err := DB.Model(&Receipt{}).Where("file_hash = ?", fileHash).Count(&count).Error; err != nil {
		log.Println("Error checking file hash in the database:", err)
		return false, err
	}
	// If the count is greater than 0, a duplicate exists
	return count > 0, nil
}


// // Existing Receipt and Item structs remain the same
// func CreateReceipt(receipt *Receipt) error {
// 	// Start a database transaction
// 	DB := db.GetDBInstance()

// 	return db.Transaction(func(tx *gorm.DB) error {
// 		// Create the receipt
// 		if err := tx.Create(receipt).Error; err != nil {
// 			return err
// 		}

// 		// Create associated items
// 		for i := range receipt.Items {
// 			// Set the ReceiptID for each item
// 			receipt.Items[i].ReceiptID = receipt.ReceiptID
// 		}

// 		// Batch create items
// 		if len(receipt.Items) > 0 {
// 			if err := tx.Create(&receipt.Items).Error; err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})
// }