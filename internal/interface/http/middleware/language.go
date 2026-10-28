package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

type contextKey string

const LanguageKey contextKey = "lang"

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang, err := c.Cookie("lang")
		if err != nil || lang == "" {
			lang = c.GetHeader("Accept-Language")
			if lang == "" {
				lang = "vi"
			}
		}

		ctx := context.WithValue(c.Request.Context(), LanguageKey, lang)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
