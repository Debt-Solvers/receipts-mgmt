package routes

import (
	"receipt-mgmt/internal/middleware"
	"receipt-mgmt/internal/tests"

	"github.com/gin-gonic/gin"
)

// Add the health check route to your main router
func AddHealthCheckRoute(router *gin.Engine) {
	router.GET("/health", tests.HealthCheck)
}


func CategoryRoutes(router *gin.Engine) {
	receiptGroup := router.Group("/api/v1/receipts")
	receiptGroup.Use(middleware.AuthMiddleware())
	{
		// receiptGroup.GET("/defaults", controller.ParseReceipt)


	}
}
