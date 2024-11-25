package controller

import (
	"context"
	"mime/multipart"
	"net/http"
	"receipt-mgmt/internal/models"
	"receipt-mgmt/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReceiptController struct {
	DB                   *gorm.DB
	CustomVisionClient   *customvision.PredictionClient
	ComputerVisionClient *computervision.BaseClient
}

func (rc *ReceiptController) UploadReceipt(c *gin.Context) {
	// Extract user ID from authenticated context
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendResponse(c, http.StatusUnauthorized, "User not authenticated", nil, nil)
		return
	}

	// Get file from request
	file, err := c.FormFile("receipt")
	if err != nil {
		utils.SendResponse(c, http.StatusBadRequest, "No receipt file uploaded", nil, nil)
		return
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Could not open file", nil, nil)
		return
	}
	defer src.Close()

	// Step 1: Custom Vision Classification
	classificationResult, err := rc.classifyReceipt(src)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Receipt classification failed", nil, err.Error())
		return
	}
	if !classificationResult.IsReceipt {
		utils.SendResponse(c, http.StatusBadRequest, "Invalid receipt image", nil, gin.H{
			"confidence": classificationResult.Confidence,
		})
		return
	}

	// Reset file reader for OCR
	_, err = src.Seek(0, 0)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Failed to reset file reader", nil, err.Error())
		return
	}

	// Step 2: OCR with Computer Vision
	ocrResult, err := rc.performOCR(src)
	if err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "OCR processing failed", nil, err.Error())
		return
	}

	// Step 3: Create Receipt Record
	receipt := models.Receipt{
		UserID:      userID.(string),
		ImageURL:    "uploadedImageURL", // Placeholder; implement image upload to storage
		Status:      "processing",
		TotalAmount: ocrResult.TotalAmount,
		Merchant:    ocrResult.Merchant,
		ScannedDate: time.Now(),
		Items:       ocrResult.Items,
	}

	// Save to database
	if err := rc.DB.Create(&receipt).Error; err != nil {
		utils.SendResponse(c, http.StatusInternalServerError, "Failed to save receipt", nil, err.Error())
		return
	}

	utils.SendResponse(c, http.StatusOK, "Receipt uploaded successfully", gin.H{
		"receipt_id": receipt.ReceiptID,
	}, nil)
}


func (rc *ReceiptController) classifyReceipt(file multipart.File) (*ClassificationResult, error) {
	// Implement Custom Vision prediction
	// Use your Custom Vision endpoint and prediction key
	predictionResults, err := rc.CustomVisionClient.PredictImage(
		context.Background(), 
		// Your project ID
		// Your iteration name
		file,
		nil,
	)

	if err != nil {
		return nil, err
	}

	// Process prediction results
	for _, prediction := range *predictionResults.Predictions {
		if *prediction.TagName == "receipt" && *prediction.Probability > 0.7 {
			return &ClassificationResult{
				IsReceipt:   true,
				Confidence: *prediction.Probability,
			}, nil
		}
	}

	return &ClassificationResult{IsReceipt: false}, nil
}



func (rc *ReceiptController) performOCR(file multipart.File) (*OCRResult, error) {
	// Perform OCR using Azure Computer Vision
	result, err := rc.ComputerVisionClient.RecognizeReceiptInStream(
		context.Background(), 
		file,
	)

	if err != nil {
		return nil, err
	}

	// Extract and parse receipt details
	return parseReceiptDetails(result), nil
}

// Structs for intermediate processing
type ClassificationResult struct {
	IsReceipt   bool
	Confidence float64
}

type OCRResult struct {
	TotalAmount float64
	Merchant    string
	Items       []models.Item
}

// Utility function to parse Computer Vision receipt result
func parseReceiptDetails(result computervision.ReadOperationResult) *OCRResult {
	// Implement complex parsing logic to extract:
	// - Total amount
	// - Merchant name
	// - Individual line items
	// This will require regex, text parsing, and possibly ML techniques
}