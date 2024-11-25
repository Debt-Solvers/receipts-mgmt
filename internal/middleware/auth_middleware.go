package middleware

import (
	"net/http"
	"receipt-mgmt/internal/common"
	"receipt-mgmt/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is middleware to validate the JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the token from the "Authorization" header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			utils.SendResponse(c, http.StatusUnauthorized, "Missing token", nil, nil)
			c.Abort()
			return
		}

		// Stripping "Bearer " prefix
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Check if token exists in the database
		if !common.IsTokenActive(tokenString) { // Create `IsTokenActive` function
			utils.SendResponse(c, http.StatusUnauthorized, "Token is invalid or expired", nil, nil)
			c.Abort()
			return
		}

		// Call VerifyToken to validate the token and extract the user ID
		userId, err := utils.VerifyToken(tokenString)
		if err != nil {
			utils.SendResponse(c, http.StatusUnauthorized, err.Error(), nil, nil)
			c.Abort()
			return
		}

		// Store the userId and tokenString in the context for further use
		c.Set("userId", userId)
		c.Set("tokenString", tokenString)
		c.Next()
	}
}

