package models

import (
	"time"
)

// // Receipt represents a receipt linked to a specific expense, with OCR and image metadata
// type Receipt struct {
// 	ID         string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"receipt_id"`
// 	ImageURL   string         `gorm:"type:text;not null" json:"image_url"`              // URL or path to the stored image
// 	OCRData    string         `gorm:"type:text" json:"ocr_data"`                        // Text extracted from the receipt via OCR
// 	ScannedDate time.Time     `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"scanned_date"`
// 	ExpenseID  *string        `gorm:"type:uuid" json:"expense_id,omitempty"`            // Foreign key to the Expense (nullable)
// 	Date  			time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
// 	UpdatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
// 	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // Soft delete
// }

// Receipt represents the receipt model with its associated fields.
type Receipt struct {
	ReceiptID   string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"receipt_id"`
	UserID      string    `gorm:"type:uuid;not null" json:"user_id"`
	ImageURL    string    `gorm:"type:text;not null" json:"image_url"`
	Status      string    `gorm:"type:varchar(50);not null" json:"status"` // Values: pending, processing, completed, failed
	TotalAmount float64   `gorm:"type:decimal(10,2)" json:"total_amount"`
	Merchant    string    `gorm:"type:varchar(255)" json:"merchant"`
	Items       []Item    `json:"items" gorm:"foreignKey:ReceiptID"` // Associated items for this receipt
	ScannedDate time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"scanned_date"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}


// Item represents an individual item extracted from a receipt.
type Item struct {
	ID         uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	ReceiptID  string  `gorm:"type:uuid;not null" json:"receipt_id"` // Foreign key to Receipt
	Name       string  `gorm:"type:varchar(255);not null" json:"name"`
	Quantity   int     `gorm:"type:int;not null" json:"quantity"`
	UnitPrice  float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice float64 `gorm:"type:decimal(10,2);not null" json:"total_price"`
}
