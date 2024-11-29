package main

import (
	"log"
	"os"
	"receipt-mgmt/db"
	"receipt-mgmt/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	if _, err := db.ConnectDatabase(); err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	
	// Initialize Gin engine
	server := gin.Default()

	// Register routes
	routes.ReceiptRoutes(server)
	routes.AddHealthCheckRoute(server)
	// Check for environment variable port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Start the server
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}