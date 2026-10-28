package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

func RegisterTranslations(lang string, dict map[string]string) {
	translationMu.Lock()
	defer translationMu.Unlock()

	if translations[lang] == nil {
		translations[lang] = make(map[string]string)
	}
	for k, v := range dict {
		translations[lang][k] = v
	}
}

func TranslateText(lang string, message string) string {
	if lang == "" {
		lang = "vi"
	}
	translationMu.RLock()
	defer translationMu.RUnlock()

	if dict, ok := translations[lang]; ok {
		if translated, found := dict[message]; found {
			return translated
		}
	}
	return message
}

func TranslateTextFromContext(ctx context.Context, message string) string {
	lang := GetLanguageFromContext(ctx)
	return TranslateText(lang, message)
}

func Translate(c *gin.Context, message string) string {
	if c == nil {
		return message
	}
	var lang string
	if val, ok := c.Get("lang"); ok {
		if l, isStr := val.(string); isStr && l != "" {
			lang = l
		}
	}
	if lang == "" && c.Request != nil {
		lang = GetLanguageFromContext(c.Request.Context())
	}
	return TranslateText(lang, message)
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: Translate(c, KeySuccess),
		Data:    data,
	})
}

func CreateSuccess(c *gin.Context, message string) {
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: Translate(c, message),
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: Translate(c, message),
		Data:    data,
	})
}

func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: Translate(c, message),
	})
}

func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: Translate(c, KeyValidationFailed),
		Errors:  errors,
	})
}

func HandleDomainError(c *gin.Context, err error, domainErrMap map[error]int) {
	for domainErr, statusCode := range domainErrMap {
		if errors.Is(err, domainErr) {
			Error(c, statusCode, err.Error())
			return
		}
	}
	Error(c, http.StatusInternalServerError, KeyInternalServerError)
}
