package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/application"
)

type AuthHandler struct {
	AuthService *application.AuthService
}

func NewAuthHandler(authService *application.AuthService) AuthHandler {
	return AuthHandler{AuthService: authService}
}

func (h *AuthHandler) ListActions(c *gin.Context) {
	actions, err := h.AuthService.GetAllActions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actions)
}
