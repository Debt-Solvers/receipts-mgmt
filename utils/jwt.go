package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// VerifyToken verifies a JWT token and returns the user ID if the token is valid
func VerifyToken(tokenString string) (uuid.UUID, error) {
	// Parse the token
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key
		return []byte(viper.GetString("JWT_SECRET")), nil
	})

	// Check if parsing the token failed
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not parse token: %v", err)
	}

	// Verify if the token is valid
	if !parsedToken.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	// Extract the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("could not parse token claims")
	}

	// Check if the "exp" claim is still valid
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return uuid.Nil, fmt.Errorf("token has expired")
		}
	} else {
		return uuid.Nil, fmt.Errorf("invalid expiration time")
	}

	// Extract the user ID
	userId, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID")
	}

	// Convert user ID to UUID
	return uuid.Parse(userId)
}
