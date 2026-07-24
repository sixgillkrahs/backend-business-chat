package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/application"
	"github.com/sixgillkrahs/backend-business-chat/internal/application/dto"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

type AuthHandler struct {
	AuthService *application.AuthService
}

func NewAuthHandler(authService *application.AuthService) AuthHandler {
	return AuthHandler{AuthService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.AuthService.Login(c.Request.Context(), req)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrInvalidCredentials: http.StatusUnauthorized,
			domain.ErrUserNotFound:       http.StatusUnauthorized,
		})
		return
	}
	utils.Success(c, resp)
}

func (h *AuthHandler) ListActions(c *gin.Context) {
	actions, err := h.AuthService.GetAllActions(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.Success(c, actions)
}

func (h *AuthHandler) ListResources(c *gin.Context) {
	resources, err := h.AuthService.GetAllResources(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.Success(c, resources)
}

func (h *AuthHandler) ListPolicies(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	resp, err := h.AuthService.GetPoliciesPaging(c, page, limit)

	utils.Success(c, resp)
}

func (h *AuthHandler) CreatePolicy(c *gin.Context) {
	utils.CreateSuccess(c, "Tạo thành công")
}
