package controller

import (
	"receipt-mgmt/internal/models"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v3.0/computervision"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v3.1/customvision/prediction"
)

type OCRService struct {
	CustomVisionClient   *prediction.BaseClient
	ComputerVisionClient *computervision.BaseClient
}

type ReceiptParseResult struct {
	TotalAmount float64
	Merchant    string
	Items       []models.Item
}
