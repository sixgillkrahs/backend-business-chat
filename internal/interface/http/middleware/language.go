package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

type contextKey = utils.ContextKey

const LanguageKey = utils.LanguageKey

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
