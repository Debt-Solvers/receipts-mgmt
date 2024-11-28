package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type CustomVisionService struct {
	computerVisionKey      string
	computerVisionEndpoint string
	customVisionKey        string
	customVisionEndpoint   string
	httpClient            *http.Client
}

type ReceiptParseResult struct {
  Merchant         string          `json:"merchant"`
  TotalAmount      float64         `json:"totalAmount"`
  ReceiptDate      string          `json:"receiptDate"`
  TransactionDate  string          `json:"transactionDate"`
  TransactionTime  string          `json:"transactionTime"`
  Items            json.RawMessage `json:"items"`  // Storing as raw JSON
  Tax              float64         `json:"tax,omitempty"`
  Discounts        float64         `json:"discounts,omitempty"`
}

// struct for the custom vision client
func NewCustomVisionService() *CustomVisionService {
	return &CustomVisionService{
		computerVisionKey:      viper.GetString("azure.computer_vision.key"),
		computerVisionEndpoint: viper.GetString("azure.computer_vision.endpoint"),
		customVisionKey:        viper.GetString("azure.custom_vision.key"),
		customVisionEndpoint:   viper.GetString("azure.custom_vision.endpoint"),
		httpClient:            &http.Client{Timeout: 30 * time.Second},
	}
}

type FormRecognizerService struct {
  Endpoint string
  Key string
  Client *http.Client
}

// Validate The Image Using Trained custom vision service (Azure AI)
func (s *CustomVisionService) ValidateReceiptImage(imageBytes []byte) (bool, error) {
  // Load the URL dynamically from the environment config
  url := viper.GetString("azure.custom_vision.url")  
  
  if url == "" {
    return false, fmt.Errorf("custom vision URL is not configured in the environment")
  }
  
  // Create a new request with the image bytes
  req, err := http.NewRequest("POST", url, bytes.NewReader(imageBytes))
  if err != nil {
    return false, fmt.Errorf("create request failed: %v", err)
  }
  
  // Set the Prediction-Key header from the environment
  req.Header.Set("Prediction-Key", viper.GetString("azure.custom_vision.key"))
  req.Header.Set("Content-Type", "application/octet-stream")
  
  // Send the request using the http client
  resp, err := s.httpClient.Do(req)
  if err != nil {
    return false, fmt.Errorf("custom vision request failed: %v", err)
  }
  defer resp.Body.Close()
  
  // Check if the response is OK (status code 200)
  if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return false, fmt.Errorf("custom vision API returned non-OK status: %d, response: %s", resp.StatusCode, string(body))
  }
  
  // Parse the response
  var result struct {
      Predictions []struct {
          TagName      string  `json:"tagName"`
          Probability float64 `json:"probability"`
      } `json:"predictions"`
  }
  
  // Decode the response body into the result struct
  if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
      return false, fmt.Errorf("failed to decode response: %v", err)
  }
  
  // Check for a "Positive" tag with high probability
  for _, prediction := range result.Predictions {
      if strings.EqualFold(prediction.TagName, "Positive") && prediction.Probability > 0.7 {
          return true, nil
      }
  }
  
  // If no valid tag is found, return false
  return false, nil
}


func AnalyzeReceipt(imageBytes []byte) (map[string]interface{}, error) {
	// Load endpoint and key using Viper
	endpoint := viper.GetString("azure.document_intelligence.endpoint")
	key := viper.GetString("azure.document_intelligence.key")

	// Log the endpoint and key (do not log the actual key in production)
	if endpoint == "" || key == "" {
		return nil, fmt.Errorf("azure form recognizer endpoint or key not configured")
	}

	// Construct the Analyze Receipt endpoint URL
	url := fmt.Sprintf("%s/formrecognizer/v2.1/prebuilt/receipt/analyze", endpoint)

	// Create the POST request
	req, err := http.NewRequest("POST", url, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set required headers
	req.Header.Set("Ocp-Apim-Subscription-key", key)
	req.Header.Set("Content-Type", "application/octet-stream")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("analyze receipt failed: %v", err)
	}
	defer resp.Body.Close()

	// If the status code is 202 (Accepted), we need to poll for results
	if resp.StatusCode == http.StatusAccepted {
		operationLocation := resp.Header.Get("Operation-Location")
		if operationLocation == "" {
			return nil, fmt.Errorf("missing operation-location header in response")
		}
		// Poll for the result and return the response as a map
		return getPoll(operationLocation, key)
	}

	// Handle cases where the response isn't accepted, if needed
	return nil, fmt.Errorf("unexpected response status: %v", resp.StatusCode)
}



func getPoll(operationLocation, key string) (map[string]interface{}, error) {
	client := &http.Client{}

	for {
		// Create GET request to the operation location
		req, err := http.NewRequest("GET", operationLocation, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create poll request: %v", err)
		}
		req.Header.Set("Ocp-Apim-Subscription-Key", key)

		// Execute the request
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to poll operation location: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read poll response body: %v", err)
		}

		// Log the raw response body for debugging
		// fmt.Printf("Raw poll response: %s\n", string(body))

		// Parse the response into a map to return the entire response
		var pollResponse map[string]interface{}
		if err := json.Unmarshal(body, &pollResponse); err != nil {
			return nil, fmt.Errorf("failed to parse poll response: %v", err)
		}

		// Log parsed response for debugging
		fmt.Printf("Parsed poll response: %+v\n", pollResponse)

		// Check status field and handle accordingly
		status, ok := pollResponse["status"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected poll response format: missing status field")
		}

		switch status {
		case "succeeded":
			// If successful, return the entire parsed response
			return pollResponse, nil

		case "failed":
			// If failed, return an error with the entire response
			return nil, fmt.Errorf("receipt analysis failed: %s", string(body))

		default:
			// Receipt analysis still in progress; wait and retry
			time.Sleep(2 * time.Second)
		}
	}
}


// ParseReceiptInformation parses the receipt information and returns a ReceiptParseResult
func ParseReceiptInformation(response map[string]interface{}) (*ReceiptParseResult, error) {
	// Initialize the result struct
	receiptResult := &ReceiptParseResult{}

	// Access 'analyzeResult' -> 'documentResults' -> first item
	analyzeResult, ok := response["analyzeResult"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to find analyzeResult in the response")
	}
	documentResults, ok := analyzeResult["documentResults"].([]interface{})
	if !ok || len(documentResults) == 0 {
		return nil, fmt.Errorf("failed to find documentResults in the response")
	}

	// Access the first document result
	documentResult := documentResults[0].(map[string]interface{})
	fields, ok := documentResult["fields"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to find fields in documentResults")
	}

	// Extract and assign the merchant name (if available)
  if merchant, ok := fields["MerchantName"].(map[string]interface{}); ok {
  var merchantText string
  
  // Try valueString first
  if val, ok := merchant["valueString"].(string); ok {
      merchantText = val
  }
  
  // If valueString is empty, try text field
  if merchantText == "" {
      if val, ok := merchant["text"].(string); ok {
          merchantText = val
      }
  }
  
  // Trim any trailing or leading whitespace
  merchantText = strings.TrimSpace(merchantText)
  
  // Set merchant, default to "Unknown" if no text found
  if merchantText == "" {
      receiptResult.Merchant = "Unknown"
  } else {
      receiptResult.Merchant = merchantText
  }
  } else {
    // If MerchantName field is not found at all, set to "Unknown"
    receiptResult.Merchant = "Unknown"
  }
  // Extract and assign the merchant name (if available)
	if merchant, ok := fields["MerchantName"].(map[string]interface{}); ok {
		if merchantText, ok := merchant["valueString"].(string); ok && merchantText != "" {
			receiptResult.Merchant = merchantText
		} else {
			receiptResult.Merchant = "Unknown" // Default value
		}
	}

	// Extract and assign total amount (if available)
	if total, ok := fields["Total"].(map[string]interface{}); ok {
		if totalAmount, ok := total["valueNumber"].(float64); ok && totalAmount > 0 {
			receiptResult.TotalAmount = totalAmount
		} else {
			receiptResult.TotalAmount = 0.0 // Default value
		}
	}

  // Input receipt Date
  if receiptResult.ReceiptDate == "" {
    receiptResult.ReceiptDate = time.Now().Format("2006-01-02 15:04:05")
  }

	// Extract and assign receipt date (if available)
	if receiptDate, ok := fields["ReceiptDate"].(map[string]interface{}); ok {
		if date, ok := receiptDate["valueString"].(string); ok && date != "" {
			receiptResult.ReceiptDate = date
		}
	}

	// Extract and assign transaction date (if available)
  if transactionDate, ok := fields["TransactionDate"].(map[string]interface{}); ok {
    fmt.Println("TransactionDate field found:", transactionDate)
    // Try valueDate first, then fallback to valueString
    if date, ok := transactionDate["valueDate"].(string); ok && date != "" {
        receiptResult.TransactionDate = date
    } else if date, ok := transactionDate["valueString"].(string); ok && date != "" {
        receiptResult.TransactionDate = date
    }
  } else {
    fmt.Println("TransactionDate field not found or empty")
  }


  // Extract and assign transaction time (if available)
  if transactionTime, ok := fields["TransactionTime"].(map[string]interface{}); ok {
    fmt.Println("TransactionTime field found:", transactionTime)
    // Try valueTime first, then fallback to valueString
    if time, ok := transactionTime["valueTime"].(string); ok && time != "" {
        receiptResult.TransactionTime = time
    } else if time, ok := transactionTime["valueString"].(string); ok && time != "" {
        receiptResult.TransactionTime = time
    }
  } else {
    fmt.Println("TransactionTime field not found or empty")
  }

  // Extract and assign tax (if available)
  if tax, ok := fields["Tax"].(map[string]interface{}); ok {
    // Try valueNumber first, then valueString converted to float
    if taxAmount, ok := tax["valueNumber"].(float64); ok {
        receiptResult.Tax = taxAmount
    } else if taxValueStr, ok := tax["valueString"].(string); ok {
        if taxAmount, err := strconv.ParseFloat(taxValueStr, 64); err == nil {
            receiptResult.Tax = taxAmount
        }
    }
  }

  // Extract and assign discounts (if available)
  // Note: Many receipts don't have a direct "Discounts" field, so we might need to adjust this
  if discounts, ok := fields["Discounts"].(map[string]interface{}); ok {
    // Try valueNumber first, then valueString converted to float
    if discountAmount, ok := discounts["valueNumber"].(float64); ok {
        receiptResult.Discounts = discountAmount
    } else if discountValueStr, ok := discounts["valueString"].(string); ok {
        if discountAmount, err := strconv.ParseFloat(discountValueStr, 64); err == nil {
            receiptResult.Discounts = discountAmount
        }
    }
  }

	// Prepare items and store them as raw JSON
	// Modify items extraction
  if items, ok := fields["Items"].(map[string]interface{}); ok {
    cleanedItems, err := cleanItems(items)
    if err != nil {
        return nil, fmt.Errorf("failed to clean items: %v", err)
    }
    
    // Convert cleaned items to JSON
    itemsJSON, err := json.Marshal(cleanedItems)
    if err != nil {
        return nil, fmt.Errorf("failed to serialize cleaned items to JSON: %v", err)
    }
    receiptResult.Items = json.RawMessage(itemsJSON)
  }
  fmt.Println("Fields in OCR Response:", fields)  // Log the entire fields structure for inspection

	return receiptResult, nil
}


func cleanItems(items map[string]interface{}) ([]map[string]interface{}, error) {
  cleanedItems := []map[string]interface{}{}

  // Check if items have the expected structure
  valueArray, ok := items["valueArray"].([]interface{})
  if !ok {
      return nil, fmt.Errorf("invalid items structure")
  }

  for _, item := range valueArray {
      itemMap, ok := item.(map[string]interface{})
      if !ok {
          continue
      }

      valueObject, ok := itemMap["valueObject"].(map[string]interface{})
      if !ok {
          continue
      }

      cleanedItem := map[string]interface{}{}

      // Extract Name (using text or valueString)
      if name, ok := valueObject["Name"].(map[string]interface{}); ok {
          if nameText, ok := name["valueString"].(string); ok {
              cleanedItem["name"] = nameText
          } else if nameText, ok := name["text"].(string); ok {
              cleanedItem["name"] = nameText
          }
      }

      // Extract TotalPrice
      if price, ok := valueObject["TotalPrice"].(map[string]interface{}); ok {
          if priceValue, ok := price["valueNumber"].(float64); ok {
              cleanedItem["totalPrice"] = priceValue
          } else if priceText, ok := price["text"].(string); ok {
              if parsedPrice, err := strconv.ParseFloat(priceText, 64); err == nil {
                  cleanedItem["totalPrice"] = parsedPrice
              }
          }
      }

      // Only add the item if it has both name and price
      if cleanedItem["name"] != nil && cleanedItem["totalPrice"] != nil {
          cleanedItems = append(cleanedItems, cleanedItem)
      }
  }

  return cleanedItems, nil
}
