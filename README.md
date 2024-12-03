# Debt Solver - Receipt Management Microservice

This repository contains the **Receipt Management** microservice for the **Debt Solver** project, a mobile application that enables users to track and manage their finances. This service handles key functionalities such as Image Recognition, parsing and storing receipts.

## Key Features

- **Receipt Upload**: Enable users to upload receipts as images for secure storage and processing
- **OCR Integration**: Use Optical Character Recognition (OCR) to extract transaction details, including dates, amounts, and merchants, directly from receipts.
- **Automatic Expense Linking**: Automatically associate extracted receipt data with existing expenses or create new entries for seamless management.
- **Receipt Insights**: Provide users with a detailed view of receipt data, including breakdowns of items, taxes, and discounts, in a user-friendly format.

## Technologies Used

- **Golang & Gin**: For building the service.
- **PostgreSQL**: For data storage (expenses, budgets, categories, and receipts).
- **GORM**: For ORM database interactions.
- **JWT**: For user authorization on protected endpoints.
- **Viper**: For configuration management.

## Directory Structure

```plaintext
expense-service/
│
├── cmd/
│   └── receipt-service/
│       └── main.go                  # Entry point for the application
│
├── configs/
│   └── config.yaml                  # Configuration file for the service
│
├── db/
│   └── migrate.go                   # Database migrations for creating tables
│
├── internal/
│   ├── common/
│   │   └── common.go                # Common utility functions
│   ├── controller/
│   │   |__receipt_controller.go     # Controller for receipt management
│   │
│   │
│   │
│   ├── middleware/
│   │   └── auth_middleware.go       # Middleware for JWT-based route protection
│   ├── model/
interactions
|   |   └── receipt.go               # Receipt model and database
        |__ user.go                  # User model
        |__auth_token.go
interactions
│   └── routes/
│       └── routes.go                # Define routes for all expense-related endpoints
│
├── utils/
│   ├── response.go                  # Utility functions for handling responses
│   └── ocr_utils.go                 # Utility functions for OCR processing
│
├── Dockerfile                       # Dockerfile for building the container
├── go.mod                           # Go module file
└── README.md                        # Project documentation
```

## Setup and Installation

git clone https://github.com/debt-solver/receipt-service.git
cd expense-service

## Setup and PostgreSQL

docker run --name debt-solver-expense-db -e POSTGRES_PASSWORD=yourpassword -d -p 5432:5432 postgres

## Install Dependencies

go mod tidy

## Run Database Migrateions

go run db/migrate.go

## Run the Application

go run cmd/expense-service/main.go

## Build and Run with Docker

docker build -t expense-service .
docker run -p 8082:8082 expense-service

## Environment Varibles

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=expense_service_db

JWT_SECRET=DebtSolverSecret
JWT_EXPIRATION_HOURS=24

# API Endpoints

## Receipts

### Get Single Receipt

- **Endpoint**: `/api/v1/receipts/{receipt_id}`
- **Method**: `/GET`
- **Description**: This endpoint allows users to retrieve the details of a single expense by its unique `receiptId`.

### Query Parameters

None.

### Path Parameters

| Parameter   | Type   | Description                              | Options/Format |
| ----------- | ------ | ---------------------------------------- | -------------- |
| `receiptId` | string | The unique ID of the expense to retrieve | UUID format    |

- **Response**:

  #### Success

  ```json
  {
  	"status": 200,
  	"message": "Receipt retrieved successfully",
  	"data": {
  		"receipt_id": "b0b87e74-b3aa-481d-a91e-d240cac56e0a",
  		"user_id": "f3486758-899e-462c-98b7-ba8f691c8718",
  		"category_id": "8c135496-ea27-446b-919e-b312394c5f36",
  		"image": "<base64_image_data>",
  		"status": "processed",
  		"total_amount": 123.45,
  		"merchant": "Walmart",
  		"items": "[{\"item\":\"item1\",\"price\":20.0},{\"item\":\"item2\",\"price\":10.0}]",
  		"scanned_date": "2024-12-01T18:00:22.735473-05:00",
  		"transaction_date": "2024-11-30",
  		"transaction_time": "14:30",
  		"file_hash": "b6d7d8e452398c4f",
  		"tax": 5.45,
  		"discounts": 2.5,
  		"created_at": "2024-12-01T18:00:22.735473-05:00",
  		"updated_at": "2024-12-01T18:00:22.735473-05:00"
  	}
  }
  ```

  #### Error

  ```json
  {
  	"status": 404,
  	"message": "Receipt not found"
  }

  {
    "status": 401,
    "message": "Unauthorized"
  }
  ```

### Get All Receipts

**Endpoint**: `GET /api/v1/receipts`  
**Description**: Retrieve all receipts associated with the authenticated user.

---

#### Query Parameters

None.

---

#### Response

##### Success

- **Status Code**: `200 OK`
- **Response Body**:

```json
{
	"status": 200,
	"message": "Receipts retrieved successfully",
	"data": [
		{
			"receipt_id": "b0b87e74-b3aa-481d-a91e-d240cac56e0a",
			"user_id": "f3486758-899e-462c-98b7-ba8f691c8718",
			"category_id": "8c135496-ea27-446b-919e-b312394c5f36",
			"image": "<base64_image_data>",
			"status": "processed",
			"total_amount": 123.45,
			"merchant": "Walmart",
			"items": "[{\"item\":\"item1\",\"price\":20.0},{\"item\":\"item2\",\"price\":10.0}]",
			"scanned_date": "2024-12-01T18:00:22.735473-05:00",
			"transaction_date": "2024-11-30",
			"transaction_time": "14:30",
			"file_hash": "b6d7d8e452398c4f",
			"tax": 5.45,
			"discounts": 2.5,
			"created_at": "2024-12-01T18:00:22.735473-05:00",
			"updated_at": "2024-12-01T18:00:22.735473-05:00"
		},
		{
			"receipt_id": "e0c68b2e-4907-44d1-a971-675ba9e3eaae",
			"user_id": "f3486758-899e-462c-98b7-ba8f691c8718",
			"category_id": "7d11f852-f0da-4c9d-b2b5-315bb76496d8",
			"image": "<base64_image_data>",
			"status": "processed",
			"total_amount": 54.0,
			"merchant": "Costco",
			"items": "[{\"item\":\"item1\",\"price\":30.0},{\"item\":\"item2\",\"price\":24.0}]",
			"scanned_date": "2024-12-02T18:00:22.735473-05:00",
			"transaction_date": "2024-11-29",
			"transaction_time": "10:15",
			"file_hash": "a3d7e9f5a12398c8",
			"tax": 3.5,
			"discounts": 1.0,
			"created_at": "2024-12-02T18:00:22.735473-05:00",
			"updated_at": "2024-12-02T18:00:22.735473-05:00"
		}
	]
}
```

### Delete Receipt

**Endpoint**: `DELETE /api/v1/receipts/{receiptId}`  
**Description**: Permanently deletes a receipt from the database using its unique ID.

---

#### Path Parameters

| Parameter   | Type   | Description                                 | Options/Format |
| ----------- | ------ | ------------------------------------------- | -------------- |
| `receiptId` | string | The unique ID of the receipt to be deleted. | UUID format    |

---

#### Response

##### Success

- **Status Code**: `200 OK`
- **Response Body**:

```json
{
	"status": 200,
	"message": "Receipt deleted successfully"
}
```

- **Status Code**: 400 Bad Reques

##### Error

```json
{
	"status": 400,
	"message": "Invalid receipt ID format",
	"error": "<error_details>"
}
```

### Upload Receipt

**Endpoint**: `POST /api/v1/receipts/upload`  
**Description**: Uploads a receipt image, processes it using OCR to extract transaction details, stores the receipt, and creates an associated expense entry.

---

#### Request Parameters

**Content-Type**: `multipart/form-data`

| Parameter     | Type   | Description                                             | Required | Format                          |
| ------------- | ------ | ------------------------------------------------------- | -------- | ------------------------------- |
| `receipt`     | File   | The receipt image file to upload.                       | Yes      | Image formats (e.g., JPEG, PNG) |
| `category_id` | String | The UUID of the category to associate with the receipt. | Yes      | UUID format                     |

---

#### Response

##### Success

- **Status Code**: `200 OK`
- **Response Body**:

```json
{
	"status": 200,
	"message": "Receipt processed successfully",
	"data": {
		"receipt": {
			"receipt_id": "b0b87e74-b3aa-481d-a91e-d240cac56e0a",
			"user_id": "f3486758-899e-462c-98b7-ba8f691c8718",
			"category_id": "8c135496-ea27-446b-919e-b312394c5f36",
			"image": "<base64_image_data>",
			"status": "completed",
			"total_amount": 123.45,
			"merchant": "Walmart",
			"items": "[{\"item\":\"item1\",\"price\":20.0},{\"item\":\"item2\",\"price\":10.0}]",
			"scanned_date": "2024-12-01T18:00:22.735473-05:00",
			"transaction_date": "2024-11-30",
			"transaction_time": "14:30",
			"tax": 5.45,
			"discounts": 2.5,
			"file_hash": "b6d7d8e452398c4f",
			"created_at": "2024-12-01T18:00:22.735473-05:00",
			"updated_at": "2024-12-01T18:00:22.735473-05:00"
		}
	}
}
```

## Error Responses

```json
{
	"status": 500,
	"message": "Failed to save receipt",
	"error": "Detailed server error message"
}
```

### Status Code: 400 Bad Request

The following errors can cause the server to respond with a `400 Bad Request` status code:

- **Missing required fields**: Ensure the request includes the `receipt` (file) and `category_id` (form data).
- **Invalid `category_id` format**: The `category_id` must be a valid UUID string.
- **Invalid receipt image file**: The uploaded file must meet the required format and validation criteria.
- **Duplicate receipt**: A file with the same hash has already been uploaded.
<!-- - **Invalid transaction date or time formats**: Ensure `TransactionDate` and `TransactionTime` values follow these formats:
  - `TransactionDate`: `YYYY-MM-DD` (e.g., `2023-11-30`)
  - `TransactionTime`: `HH:mm` (24-hour clock, e.g., `14:30`) -->
