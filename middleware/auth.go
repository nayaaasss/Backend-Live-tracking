package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("supersecretkey")

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is empty"})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Ambil user_id (string atau float64)
		var userID string
		switch v := claims["user_id"].(type) {
		case string:
			userID = v
		case float64:
			userID = fmt.Sprintf("%.0f", v)
		case int:
			userID = fmt.Sprintf("%d", v)
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid user_id type: %T", v)})
			c.Abort()
			return
		}

		// Ambil role
		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims: role"})
			c.Abort()
			return
		}

		// Ambil email
		email, ok := claims["email"].(string)
		if !ok {
			email = ""
		}

		c.Set("user_id", userID)
		c.Set("role", role)
		c.Set("email", email)

		c.Next()
	}
}
