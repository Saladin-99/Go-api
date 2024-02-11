package utils

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dgrijalva/jwt-go"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the JWT token from the request header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}
		tokenString := strings.Split(authHeader, " ")[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// Return the secret key used to sign the token
			return []byte("your-secret-key"), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization token"})
			c.Abort()
			return
		}

		// Check if the token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization token"})
			c.Abort()
			return
		}

		// Extract claims from the token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization token"})
			c.Abort()
			return
		}

		// Check expiration
		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		if exp.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token has expired"})
			c.Abort()
			return
		}

		// Check if the token is blacklisted (if needed)

		// If the token is valid, proceed to the next handler
		c.Set("user_id", claims["user_id"]) // Set user_id in context for further use
		c.Next()
	}
}
