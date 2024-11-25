package db

import (
	"fmt"
	"log"
	"receipt-mgmt/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB - Global database connection
var DB *gorm.DB

// ConnectDatabase initializes a connection to the PostgreSQL database
func ConnectDatabase() (*gorm.DB, error) {
	config := configs.LoadConfig() // Load the configuration

	// Build the connection string
	dbURI := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode)

	// Open a connection to the database
	var err error
	DB, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	log.Println("Database connected successfully")
	return DB,nil // Return nil error if successful
}

// GetDBInstance returns the DB instance
func GetDBInstance() *gorm.DB {
	return DB
}