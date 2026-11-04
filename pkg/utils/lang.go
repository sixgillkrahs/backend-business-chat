package utils

import (
	"context"
)

type ContextKey string

const LanguageKey ContextKey = "lang"

func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(LanguageKey).(string); ok && lang != "" {
		return lang
	}
	return "vi"
}
