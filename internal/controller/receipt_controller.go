package controller

import (
	"fmt"
	"io"
	"net/http"
	"receipt-mgmt/internal/common"
	"receipt-mgmt/internal/models"
	"receipt-mgmt/internal/services"
	"receipt-mgmt/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	
  // Prepare Receipt Model
	 receipt := models.Receipt{
		ReceiptID:       uuid.New(),
		UserID:          userID.(uuid.UUID),
		Image:           fileBytes, // Store the actual image as byte array
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
	// utils.SendResponse(c, http.StatusOK, "Receipt processed successfully", parsedReceiptDetails, nil)
	utils.SendResponse(c, http.StatusOK, "Receipt processed successfully", receipt, nil)


	// // Respond with success
	// utils.SendResponse(c, http.StatusOK, "Receipt processed successfully", receipt, nil)
}


// Additional methods for receipt management can be added here
// func (rc *ReceiptController) ListReceipts(c *gin.Context) {
// 	// Implement listing receipts
// }

// func (rc *ReceiptController) GetReceipt(c *gin.Context) {
// 	// Implement get single receipt
// }