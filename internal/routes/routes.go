package routes

import (
	"receipt-mgmt/internal/controller"
	"receipt-mgmt/internal/middleware"
	"receipt-mgmt/internal/tests"

	"github.com/gin-gonic/gin"
)

// Add the health check route to your main router
func AddHealthCheckRoute(router *gin.Engine) {
	router.GET("/health", tests.HealthCheck)
}

func ReceiptRoutes(router *gin.Engine) {
	receiptsGroup := router.Group("/api/v1/receipts")
	receiptsGroup.Use(middleware.AuthMiddleware())
	{
		receiptsGroup.POST("/upload", controller.UploadReceipt)
		// receiptsGroup.GET("/", controller.ListReceipts)
		// receiptsGroup.GET("/:id", controller.GetReceipt)
	}
}