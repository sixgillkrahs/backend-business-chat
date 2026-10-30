package handlers

import (
	"net/http"
	"strconv"
	"strings"

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

// Login godoc
// @Summary      Đăng nhập
// @Description  Xác thực người dùng và trả về JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Thông tin đăng nhập"
// @Success      200   {object}  map[string]string "access_token"
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	ip := c.ClientIP()
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	token, err := h.AuthService.Login(c.Request.Context(), req, lang, ip)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrInvalidCredentials: http.StatusUnauthorized,
			domain.ErrUserNotFound:       http.StatusUnauthorized,
			domain.ErrUserLocked:         http.StatusForbidden,
		})
		return
	}
	c.SetCookie("access_token", token, 86400, "/", "", false, true)

	utils.Success(c, gin.H{
		"access_token": token,
	})
}

// ListActions godoc
// @Summary      Danh sách actions
// @Description  Lấy toàn bộ danh sách actions trong hệ thống
// @Tags         auth
// @Produce      json
// @Success      200  {array}   domain.Action
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/actions [get]
func (h *AuthHandler) ListActions(c *gin.Context) {
	actions, err := h.AuthService.GetAllActions(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.Success(c, actions)
}

// ListResources godoc
// @Summary      Danh sách resources
// @Description  Lấy toàn bộ danh sách resources trong hệ thống
// @Tags         auth
// @Produce      json
// @Success      200  {array}   domain.Resource
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/resources [get]
func (h *AuthHandler) ListResources(c *gin.Context) {
	resources, err := h.AuthService.GetAllResources(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.Success(c, resources)
}

// ListPolicies godoc
// @Summary      Danh sách policies (phân trang và tìm kiếm)
// @Description  Lấy danh sách policies với pagination và tìm kiếm
// @Tags         auth
// @Produce      json
// @Param        page    query     int     false  "Trang hiện tại (default: 1)"
// @Param        limit   query     int     false  "Số bản ghi mỗi trang (default: 10, max: 50)"
// @Param        search  query     string  false  "Từ khóa tìm kiếm (theo tên hoặc mô tả)"
// @Success      200     {object}  dto.PolicyPaginationResponse
// @Failure      500     {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/policies [get]
func (h *AuthHandler) ListPolicies(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	search := strings.TrimSpace(c.Query("search"))
	if search == "" {
		search = strings.TrimSpace(c.Query("q"))
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	resp, err := h.AuthService.GetPoliciesPaging(c.Request.Context(), page, limit, search)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// CreatePolicy godoc
// @Summary      Tạo policy mới
// @Description  Tạo một policy phân quyền mới
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request      body     dto.PolicyRequest  true  "Thông tin policy"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/policies [post]
func (h *AuthHandler) CreatePolicy(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	var req dto.PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.AuthService.CreatePolicy(c, req, lang)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.CreateSuccess(c, "Tạo thành công")
}

func (h *AuthHandler) GetPolicyById(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	id := c.Param("id")
	idNum, err := strconv.Atoi(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	policy, err := h.AuthService.GetPolicyById(c.Request.Context(), idNum, lang)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, policy)
}
