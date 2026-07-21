package handlers

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/application"
	"github.com/sixgillkrahs/backend-business-chat/internal/domain"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
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

	var (
		policies []domain.Policy
		count    int64
		err1     error
		err2     error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		policies, err1 = h.AuthService.GetPolicies(c.Request.Context(), page, limit)
	}()

	go func() {
		defer wg.Done()
		count, err2 = h.AuthService.CountPolicies(c.Request.Context())
	}()

	wg.Wait()

	if err1 != nil {
		utils.Error(c, http.StatusInternalServerError, err1.Error())
		return
	}
	if err2 != nil {
		utils.Error(c, http.StatusInternalServerError, err2.Error())
		return
	}

	utils.Success(c, gin.H{
		"items": policies,
		"total": count,
		"page":  page,
		"limit": limit,
	})
}

func (h *AuthHandler) CreatePolicy(c *gin.Context) {
	utils.CreateSuccess(c, "Tạo thành công")
}
