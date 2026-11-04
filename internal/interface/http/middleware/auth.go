package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

// AuthMiddleware validates the JWT access token and injects userID, username, and role into the Gin context.
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.Error(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			utils.Error(c, http.StatusUnauthorized, "Token is required")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenStr, secretKey)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		if claims.TokenType != "access" && claims.TokenType != "" {
			utils.Error(c, http.StatusUnauthorized, "Access token required")
			c.Abort()
			return
		}

		c.Set("userID", int(claims.UserID))
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
