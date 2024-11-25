package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"receipt-mgmt/db"
	"receipt-mgmt/internal/models"
	"receipt-mgmt/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// services/ocr/ocr.go
type OCRService struct {
	client *vision.ImageAnnotatorClient
}

func NewOCRService(ctx context.Context) (*OCRService, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
			return nil, fmt.Errorf("failed to create client: %v", err)
	}
	return &OCRService{client: client}, nil
}

// handlers/receipt_handler.go
func UploadReceipt(c *gin.Context) {
  // Get the uploaded file
  file, err := c.FormFile("receipt")
  if err != nil {
      utils.SendResponse(c, http.StatusBadRequest, "No file uploaded", nil, err.Error())
      return
  }

  // Get user ID from context
  userID, _ := c.Get("userId")

  // Generate unique filename
  filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
  filepath := fmt.Sprintf("uploads/receipts/%s", filename)

  // Save the file
  if err := c.SaveUploadedFile(file, filepath); err != nil {
      utils.SendResponse(c, http.StatusInternalServerError, "Failed to save file", nil, err.Error())
      return
  }

  // Create receipt record
  receipt := &models.Receipt{
    UserID:    userID.(uint),
    ImagePath: filepath,
    Status:    "pending",
  }

  if err := db.GetDBInstance().Create(receipt).Error; err != nil {
    utils.SendResponse(c, http.StatusInternalServerError, "Failed to create receipt record", nil, err.Error())
    return
  }

  // Trigger OCR processing
  go ProcessReceipt(receipt.ID)

  utils.SendResponse(c, http.StatusOK, "Receipt uploaded successfully", receipt, nil)
}

// services/ocr/process.go
func (s *OCRService) ProcessReceipt(ctx context.Context, imagePath string) (*models.Receipt, error) {
  // Read the image file
  file, err := os.Open(imagePath)
  if err != nil {
      return nil, fmt.Errorf("failed to open image: %v", err)
  }
  defer file.Close()

  image, err := vision.NewImageFromReader(file)
  if err != nil {
      return nil, fmt.Errorf("failed to create image: %v", err)
  }

  // Detect text in the image
  texts, err := s.client.DetectDocumentText(ctx, image, nil)
  if err != nil {
      return nil, fmt.Errorf("failed to detect text: %v", err)
  }

  // Parse the extracted text
  receiptData := parseReceiptText(texts.Text)
  
  return receiptData, nil
}

// services/ocr/process.go
func (s *OCRService) ProcessReceipt(ctx context.Context, imagePath string) (*models.Receipt, error) {
  // Read the image file
  file, err := os.Open(imagePath)
  if err != nil {
      return nil, fmt.Errorf("failed to open image: %v", err)
  }
  defer file.Close()

  image, err := vision.NewImageFromReader(file)
  if err != nil {
      return nil, fmt.Errorf("failed to create image: %v", err)
  }

  // Detect text in the image
  texts, err := s.client.DetectDocumentText(ctx, image, nil)
  if err != nil {
      return nil, fmt.Errorf("failed to detect text: %v", err)
  }

  // Parse the extracted text
  receiptData := parseReceiptText(texts.Text)
  
  return receiptData, nil
}

func parseReceiptText(text string) *models.Receipt {
  receipt := &models.Receipt{}
  
  // Use regular expressions to extract information
  // This is a basic example - you'll need to customize based on your receipt formats
  
  // Extract total amount
  totalRegex := regexp.MustCompile(`Total:?\s*\$?(\d+\.\d{2})`)
  if matches := totalRegex.FindStringSubmatch(text); len(matches) > 1 {
      receipt.TotalAmount, _ = strconv.ParseFloat(matches[1], 64)
  }

  // Extract date
  dateRegex := regexp.MustCompile(`(\d{2}/\d{2}/\d{4})`)
  if matches := dateRegex.FindStringSubmatch(text); len(matches) > 1 {
      receipt.Date, _ = time.Parse("01/02/2006", matches[1])
  }

  // Extract merchant name (this is simplified)
  lines := strings.Split(text, "\n")
  if len(lines) > 0 {
      receipt.Merchant = lines[0]
  }

  return receipt
}
// workers/receipt_processor.go
func ProcessReceipt(receiptID uint) {
  db := db.GetDBInstance()
  
  // Get the receipt record
  var receipt models.Receipt
  if err := db.First(&receipt, receiptID).Error; err != nil {
      log.Printf("Failed to find receipt: %v", err)
      return
  }

  // Update status to processing
  db.Model(&receipt).Update("status", "processing")

  // Initialize OCR service
  ctx := context.Background()
  ocrService, err := ocr.NewOCRService(ctx)
  if err != nil {
      log.Printf("Failed to create OCR service: %v", err)
      db.Model(&receipt).Update("status", "failed")
      return
  }

  // Process the receipt
  processedReceipt, err := ocrService.ProcessReceipt(ctx, receipt.ImagePath)
  if err != nil {
      log.Printf("Failed to process receipt: %v", err)
      db.Model(&receipt).Update("status", "failed")
      return
  }

  // Update receipt with extracted data
  updates := map[string]interface{}{
      "total_amount": processedReceipt.TotalAmount,
      "date":        processedReceipt.Date,
      "merchant":    processedReceipt.Merchant,
      "status":      "completed",
  }
  
  if err := db.Model(&receipt).Updates(updates).Error; err != nil {
      log.Printf("Failed to update receipt: %v", err)
      return
  }

  // Create expense record
  expense := models.Expense{
      UserID:     receipt.UserID,
      Amount:     processedReceipt.TotalAmount,
      Date:       processedReceipt.Date,
      Merchant:   processedReceipt.Merchant,
      ReceiptID:  receipt.ID,
  }

  if err := db.Create(&expense).Error; err != nil {
      log.Printf("Failed to create expense: %v", err)
      return
  }
}