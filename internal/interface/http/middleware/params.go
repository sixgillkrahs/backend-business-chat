package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func ValidateParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, err := strconv.Atoi(c.Request.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}
		q := c.Request.URL.Query()
		q.Set("limit", strconv.Itoa(limit))
		c.Request.URL.RawQuery = q.Encode()
		c.Next()
	}
}
