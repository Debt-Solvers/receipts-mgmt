package models

import (
	"encoding/json"
	"fmt"
	"log"
	"receipt-mgmt/db"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Receipt represents the receipt model with its associated fields.
type Receipt struct {
	ReceiptID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"receipt_id"`
	UserID           uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	CategoryID       uuid.UUID       `gorm:"type:uuid;not null" json:"category_id"`
	Image            []byte          `gorm:"type:bytea;not null" json:"image"`
	Status           string          `gorm:"type:varchar(50);not null" json:"status"`
	TotalAmount      float64         `gorm:"type:decimal(10,2)" json:"total_amount"`
	Merchant         string          `gorm:"type:varchar(255)" json:"merchant"`
	Items            json.RawMessage `gorm:"type:jsonb" json:"items"` // JSONB column
	ScannedDate      time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP" json:"scanned_date"`
	TransactionDate  string          `gorm:"type:varchar(50);not null" json:"transaction_date"`
	TransactionTime  string          `gorm:"type:varchar(50);not null" json:"transaction_time"`
	FileHash         string          `gorm:"type:varchar(64);unique;not null" json:"file_hash"`
	Tax              float64         `gorm:"type:decimal(10,2)" json:"tax"`
	Discounts        float64         `gorm:"type:decimal(10,2)" json:"discounts"`
	CreatedAt        time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}


// Category represents a user-defined or default spending category
type Category struct {
	ID 					uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"category_id"`
	UserID      *uuid.UUID     `gorm:"type:uuid" json:"user_id,omitempty"`           // Nullable for default categories
	Name        string         `gorm:"size:50;not null" json:"name"`                 // e.g., "Food", "Utilities"
	Description string         `gorm:"type:text" json:"description"`                 // Optional description
	ColorCode   string         `gorm:"size:7" json:"color_code"`                     // Optional color code (e.g., #FFFFFF)
	IsDefault   bool           `gorm:"default:false" json:"is_default"`              // True if the category is default
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`            // Soft delete
}


// Expense represents an individual expense entry associated with a user and category
type Expense struct {
	ExpenseID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"expense_id"`
	UserID             uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	CategoryID         uuid.UUID     `gorm:"type:uuid;not null" json:"category_id"`
	Amount             float64       `gorm:"type:decimal(10,2);not null" json:"amount"`
	Date               time.Time     `gorm:"type:timestamp;not null" json:"date"`
	Description        string        `gorm:"type:text" json:"description"`
	ReceiptID          *uuid.UUID    `gorm:"type:uuid" json:"receipt_id"`
	CreatedAt          time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
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


// IsCategoryIDValid checks if a category ID exists in the database
func IsCategoryIDValid(categoryID string) (bool, error) {
	DB := db.GetDBInstance()

	var count int64
	err := DB.Model(&Category{}).Where("id = ?", categoryID).Count(&count).Error
	if err != nil {
		// Return false and propagate the error
		return false, err
	}

	// Return true if the count is greater than zero
	return count > 0, nil
}

func CreateExpense(expense *Expense) error {
	DB := db.GetDBInstance()
	return DB.Create(expense).Error
}


func CreateReceipt(receipt *Receipt) error {
	// Start a database transaction
	DB := db.GetDBInstance()

	return DB.Transaction(func(tx *gorm.DB) error {
		// Create the receipt (including the JSON Items field)
		if err := tx.Create(receipt).Error; err != nil {
			log.Printf("Error creating receipt: %v", err)
			return fmt.Errorf("error creating receipt: %w", err)
		}
		return nil
	})
}

// Get All receipts for a user
func GetReceiptsByUserID(userID uuid.UUID) ([]Receipt, error) {
	// Start a database transaction
	DB := db.GetDBInstance()

	var receipts []Receipt
	err := DB.Where("user_id = ?", userID).Find(&receipts).Error
	return receipts, err
}

// Get Single Receipt By ID
func GetReceiptByID(receiptID, userID uuid.UUID) (Receipt, error) {
	// Start a database transaction
	DB := db.GetDBInstance()

	var receipt Receipt
	err := DB.Where("receipt_id = ? AND user_id = ?", receiptID, userID).First(&receipt).Error
	return receipt, err
}


