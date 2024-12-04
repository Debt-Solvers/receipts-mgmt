package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"receipt-mgmt/db"
	"receipt-mgmt/internal/common"
	"receipt-mgmt/internal/models"
	"receipt-mgmt/internal/services"
	"receipt-mgmt/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func UploadReceipt(c *gin.Context) {
	// Extract user ID from context (assumes AuthMiddleware sets this)
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil, nil)
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, "Invalid file upload", nil, nil)
		return
	}

	// Get the file from the request
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, "No file uploaded", nil, nil)
		return
	}
	defer file.Close()

	// Get the category_id from the form
	categoryID := c.PostForm("category_id")
	if categoryID == "" {
		utils.SendResponse(c, http.StatusBadRequest, "category_id is required", nil, nil)
		return
	}

	// Validate category_id
	isValid, err := models.IsCategoryIDValid(categoryID)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, fmt.Sprintf("Error checking category: %v", err), nil, nil)
		return
	}
	if !isValid {
		utils.SendResponse(c, http.StatusBadRequest, "Invalid category_id", nil, nil)
		return
	}


	// Generate file hash
	file.Seek(0, io.SeekStart) // Ensure pointer starts at beginning
	fileHash, err := common.GenerateFileHash(file)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Failed to generate file hash", nil, nil)
		return
	}
	file.Seek(0, io.SeekStart) // Reset pointer for further use

	// Check if file hash already exists in the database
	if exists, err := models.CheckFileHashExists(fileHash); err == nil && exists {
		utils.SendResponse(c, http.StatusBadRequest, "Receipt already uploaded", nil, nil)
		return
	}

	// Reset file reader for further operations (e.g., saving the file)
	file.Seek(0, io.SeekStart)

	// Read file contents
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Failed to read file", nil, nil)
		return
	}
	// Log the filename and size
	fmt.Printf("Received file: %s (%d bytes)\n", header.Filename, len(fileBytes))
	
  //Step 1: Validate the receipt image using Custom Vision
  isValidReceipt, err := services.NewCustomVisionService().ValidateReceiptImage(fileBytes)
	
	if err != nil {
    // Log the error
    errorMsg := fmt.Sprintf("Error in Custom Vision API: %v\n", err)
    // Send the error response
    utils.SendResponse(c, http.StatusBadRequest, errorMsg, nil, nil)
    return
	}

	if !isValidReceipt { 
    // Create an error message string
    errorMsg := "Receipt is invalid according to Custom Vision (no positive tag found or too low probability)."
    
    // Log the error
    fmt.Println(errorMsg)
    
    // Send the error response with the error message
    utils.SendResponse(c, http.StatusBadRequest, errorMsg, nil, nil)
    return 
  }

  // Step 2: Extract Receipt Details
  receiptDetails, err := services.AnalyzeReceipt(fileBytes)
  if err != nil {
    utils.SendResponse(c, http.StatusInternalServerError,fmt.Sprintf("Analyze receipt error: %v", err), nil, nil)
    return
  }

	// Step 3: Return extracted details Parse the receipt information
  parsedReceiptDetails, err := services.ParseReceiptInformation(receiptDetails)
  if err != nil {
    utils.SendResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to parse receipt details: %v", err), nil, nil)
    return
  }

	// Convert to uuid.UUID
	parsedCategoryID, err := uuid.Parse(categoryID)
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, "Invalid category ID", nil, nil)
		return
	}
	
  // Prepare Receipt Model
	 receipt := models.Receipt{
		ReceiptID:       uuid.New(),
		UserID:          userID.(uuid.UUID),
		CategoryID: 		 parsedCategoryID,	
		Image:           fileBytes, 	// Store the actual image as byte array
		Status:          "completed",
		TotalAmount:     parsedReceiptDetails.TotalAmount,
		Merchant:        parsedReceiptDetails.Merchant,
		ScannedDate:     time.Now(),
		TransactionDate: parsedReceiptDetails.TransactionDate, // Extracted from receipt
		TransactionTime: parsedReceiptDetails.TransactionTime, // Extracted from receipt
		Tax:             parsedReceiptDetails.Tax,
		Discounts:       parsedReceiptDetails.Discounts,
		Items:           parsedReceiptDetails.Items, // Assuming items are in JSON format
		FileHash: 			 fileHash,			
	}

	// Save Receipt and Associated Items
	if err := models.CreateReceipt(&receipt); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Failed to save receipt", nil, map[string]interface{}{
			"error": err.Error(), // Include detailed error message
		})
		return
	}

	// Parse TransactionDate from string to time.Time
	var combinedDateTime time.Time

	// Handle missing TransactionDate or TransactionTime
	if parsedReceiptDetails.TransactionDate == "" && parsedReceiptDetails.TransactionTime == "" {
		// Fallback to current date and time
		combinedDateTime = time.Now()
	} else {
		// Parse TransactionDate if available
		transactionDate := time.Now()
		if parsedReceiptDetails.TransactionDate != "" {
			transactionDate, err = time.Parse("2006-01-02", parsedReceiptDetails.TransactionDate)
			if err != nil {
				utils.SendResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid transaction date: %v", err), nil, nil)
				return
			}
		}

		// Parse TransactionTime if available
		if parsedReceiptDetails.TransactionTime != "" {
			transactionTime, err := time.Parse("15:04:05", parsedReceiptDetails.TransactionTime)
			if err != nil {
					utils.SendResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid transaction time: %v", err), nil, nil)
					return
			}
	
			// Combine date and time
			combinedDateTime = time.Date(
					transactionDate.Year(), transactionDate.Month(), transactionDate.Day(),
					transactionTime.Hour(), transactionTime.Minute(), transactionTime.Second(),
					0, time.UTC,
			)
		} else {
				// Use the transaction date with default time (midnight)
				combinedDateTime = time.Date(
						transactionDate.Year(), transactionDate.Month(), transactionDate.Day(),
						0, 0, 0, 0, time.UTC,
				)
		}
	
	}

	// Create expense with fallback or parsed date
	expense := models.Expense{
		ExpenseID:   uuid.New(),
		UserID:      receipt.UserID,
		CategoryID:  receipt.CategoryID,
		Amount:      receipt.TotalAmount,
		Date:        combinedDateTime, // Always valid
		Description: fmt.Sprintf("Expense from receipt: %s", receipt.Merchant),
		ReceiptID:   &receipt.ReceiptID, // Link to the receipt
	}

	if err := models.CreateExpense(&expense); err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Failed to save expense", nil, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Respond with success
	utils.SendResponse(c, http.StatusOK, "Receipt processed successfully", receipt, nil)
}

// Get all receipts
func GetAllReceipts(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
			utils.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil, nil)
			return
	}

	receipts, err := models.GetReceiptsByUserID(userID.(uuid.UUID))
	if err != nil {
			utils.SendResponse(c, http.StatusInternalServerError, "Failed to fetch receipts", nil, map[string]interface{}{
					"error": err.Error(),
			})
			return
	}

	utils.SendResponse(c, http.StatusOK, "Receipts retrieved successfully", receipts, nil)
}

// Get single Receipts
func GetReceiptByID(c *gin.Context) {
	// Get the user ID from context
	userID, exists := c.Get("userId")
	if !exists {
			utils.SendResponse(c, http.StatusUnauthorized, "Unauthorized", nil, nil)
			return
	}

	// Parse the receipt ID from the URL parameter
	receiptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
			utils.SendResponse(c, http.StatusBadRequest, "Invalid receipt ID", nil, nil)
			return
	}

	// Debugging - log the userID and receiptID
	fmt.Printf("UserID: %v, ReceiptID: %v\n", userID, receiptID)

	// Check if the receipt exists for the user in the database
	DB := db.GetDBInstance()
	var receipt models.Receipt
	err = DB.Where("receipt_id = ? AND user_id = ?", receiptID, userID).First(&receipt).Error
	if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
					utils.SendResponse(c, http.StatusNotFound, "Receipt not found", nil, nil)
			} else {
					utils.SendResponse(c, http.StatusInternalServerError, "Failed to fetch receipt", nil, map[string]interface{}{
							"error": err.Error(),
					})
			}
			return
	}

	// Send the response with the receipt details
	utils.SendResponse(c, http.StatusOK, "Receipt retrieved successfully", receipt, nil)
}


// DeleteReceipt deletes a receipt by its ID permanently
func DeleteReceipt(c *gin.Context) {
	// Get the receipt ID from the URL parameter
	receiptIDStr := c.Param("id")
	receiptID, err := uuid.Parse(receiptIDStr)
	if err != nil {
		// Return bad request if ID format is invalid
		utils.SendResponse(c, http.StatusBadRequest, "Invalid receipt ID format", nil, err)
		return
	}

	// Get DB instance
	DB := db.GetDBInstance()

	// Check if the receipt exists for the user and is not deleted (no soft delete)
	var receipt models.Receipt
	if err := DB.Where("receipt_id = ?", receiptID).First(&receipt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return not found if receipt does not exist
			utils.SendResponse(c, http.StatusNotFound, "Receipt not found", nil, nil)
		} else {
			// Return internal server error if fetching receipt fails
			utils.SendResponse(c, http.StatusInternalServerError, "Failed to check receipt", nil, map[string]interface{}{"error": err.Error()})
		}
		return
	}

	// Perform the actual permanent deletion (remove the receipt from the database)
	if err := DB.Unscoped().Delete(&receipt).Error; err != nil {
		// Return internal server error if deletion fails
		utils.SendResponse(c, http.StatusInternalServerError, "Failed to delete receipt", nil, err)
		return
	}

	// Send the success response after deletion
	utils.SendResponse(c, http.StatusOK, "Receipt deleted successfully", nil, nil)
}
