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

// GetPolicyById godoc
// @Summary      Chi tiết policy theo ID
// @Description  Lấy thông tin chi tiết của một policy theo ID
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của policy"
// @Success      200  {object}  domain.Policy
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/policies/{id} [get]
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

// UpdatePolicyById godoc
// @Summary      Cập nhật policy theo ID
// @Description  Cập nhật thông tin của một policy theo ID
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "ID của policy"
// @Param        request  body      dto.UpdatePolicyRequest  true  "Thông tin cập nhật policy"
// @Success      200      {object}  domain.Policy
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/policies/{id} [put]
func (h *AuthHandler) UpdatePolicyById(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	id := c.Param("id")
	idNum, err := strconv.Atoi(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	var req dto.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	policy, err := h.AuthService.UpdatePolicyById(c.Request.Context(), idNum, req, lang)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, policy)
}


// DeletePolicyById godoc
// @Summary      Xóa policy theo ID
// @Description  Xóa một policy khỏi hệ thống theo ID
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của policy"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/policies/{id} [delete]
func (h *AuthHandler) DeletePolicyById(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	id := c.Param("id")
	idNum, err := strconv.Atoi(id)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.AuthService.DeletePolicyById(c.Request.Context(), idNum, lang); err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "Xóa thành công"})
}


// CreatePermission godoc
// @Summary      Tạo permission mới
// @Description  Tạo một permission mới liên kết với nhiều policy IDs
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreatePermissionRequest  true  "Thông tin permission và danh sách policy IDs"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/permissions [post]
func (h *AuthHandler) CreatePermission(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.AuthService.CreatePermission(c.Request.Context(), req, lang)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrPolicyNotFound:          http.StatusBadRequest,
			domain.ErrEmptyPolicies:           http.StatusBadRequest,
			domain.ErrPermissionNameDuplicate: http.StatusConflict,
		})
		return
	}

	utils.Success(c, gin.H{
		"id":      id,
		"message": "Tạo permission thành công",
	})
}


// ListPermissions godoc
// @Summary      Danh sách permissions (phân trang limit & offset)
// @Description  Lấy danh sách permissions kèm danh sách các policy của mỗi permission, hỗ trợ tìm kiếm và phân trang bằng limit & offset
// @Tags         auth
// @Produce      json
// @Param        limit   query     int     false  "Số bản ghi cần lấy (default: 10, max: 50)"
// @Param        offset  query     int     false  "Số bản ghi cần bỏ qua (default: 0)"
// @Param        search  query     string  false  "Từ khóa tìm kiếm (theo tên hoặc mô tả)"
// @Success      200     {object}  dto.PermissionPaginationResponse
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/permissions [get]
func (h *AuthHandler) ListPermissions(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	search := strings.TrimSpace(c.Query("search"))
	if search == "" {
		search = strings.TrimSpace(c.Query("q"))
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	resp, err := h.AuthService.GetPermissionsPaging(c.Request.Context(), limit, offset, search)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, resp)
}
