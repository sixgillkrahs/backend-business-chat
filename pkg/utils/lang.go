package utils

import (
	"context"

	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/middleware"
)

func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(middleware.LanguageKey).(string); ok && lang != "" {
		return lang
	}
	return "vi"
}
