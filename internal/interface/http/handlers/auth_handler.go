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

// GetPermissionById godoc
// @Summary      Chi tiết permission theo ID
// @Description  Lấy thông tin chi tiết của một permission kèm danh sách policies theo ID
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của permission"
// @Success      200  {object}  domain.Permission
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/permissions/{id} [get]
func (h *AuthHandler) GetPermissionById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	perm, err := h.AuthService.GetPermissionById(c.Request.Context(), id)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrPermissionNotFound: http.StatusNotFound,
		})
		return
	}

	utils.Success(c, perm)
}

// UpdatePermissionById godoc
// @Summary      Cập nhật permission theo ID
// @Description  Cập nhật thông tin của một permission và danh sách policies theo ID
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "ID của permission"
// @Param        request  body      dto.UpdatePermissionRequest true  "Thông tin cập nhật permission"
// @Success      200      {object}  domain.Permission
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/permissions/{id} [put]
func (h *AuthHandler) UpdatePermissionById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	lang := utils.GetLanguageFromContext(c)
	perm, err := h.AuthService.UpdatePermissionById(c.Request.Context(), id, req, lang)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrPermissionNotFound:      http.StatusNotFound,
			domain.ErrPermissionNameDuplicate: http.StatusConflict,
			domain.ErrPolicyNotFound:          http.StatusBadRequest,
			domain.ErrEmptyPolicies:           http.StatusBadRequest,
		})
		return
	}

	utils.Success(c, perm)
}

// DeletePermissionById godoc
// @Summary      Xóa permission theo ID
// @Description  Xóa một permission và các liên kết policies khỏi hệ thống theo ID
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của permission"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/permissions/{id} [delete]
func (h *AuthHandler) DeletePermissionById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	lang := utils.GetLanguageFromContext(c)
	if err := h.AuthService.DeletePermissionById(c.Request.Context(), id, lang); err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrPermissionNotFound: http.StatusNotFound,
		})
		return
	}

	utils.Success(c, gin.H{
		"message": utils.Translate(c, utils.KeyDeletedSuccess),
	})
}

// CreateRole godoc
// @Summary      Tạo role mới
// @Description  Tạo một role mới liên kết với danh sách permission IDs
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateRoleRequest  true  "Thông tin role và danh sách permission IDs"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/roles [post]
func (h *AuthHandler) CreateRole(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.AuthService.CreateRole(c.Request.Context(), req, lang)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrPermissionNotFound:   http.StatusBadRequest,
			domain.ErrEmptyRolePermissions: http.StatusBadRequest,
			domain.ErrRoleNameDuplicate:    http.StatusConflict,
		})
		return
	}

	utils.Success(c, gin.H{
		"id":      id,
		"message": utils.Translate(c, "Tạo role thành công"),
	})
}

// ListRoles godoc
// @Summary      Danh sách roles (phân trang và tìm kiếm)
// @Description  Lấy danh sách roles kèm permissions với pagination và tìm kiếm
// @Tags         auth
// @Produce      json
// @Param        page    query     int     false  "Trang hiện tại (default: 1)"
// @Param        limit   query     int     false  "Số bản ghi mỗi trang (default: 10, max: 50)"
// @Param        search  query     string  false  "Từ khóa tìm kiếm (theo tên hoặc mô tả)"
// @Success      200     {object}  dto.RolePaginationResponse
// @Failure      500     {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/roles [get]
func (h *AuthHandler) ListRoles(c *gin.Context) {
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

	resp, err := h.AuthService.GetRolesPaging(c.Request.Context(), page, limit, search)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// UpdateRoleById godoc
// @Summary      Cập nhật role theo ID
// @Description  Cập nhật thông tin của một role và danh sách permissions theo ID
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "ID của role"
// @Param        request  body      dto.UpdateRoleRequest  true  "Thông tin cập nhật role"
// @Success      200      {object}  domain.Role
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/roles/{id} [put]
func (h *AuthHandler) UpdateRoleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	lang := utils.GetLanguageFromContext(c)
	role, err := h.AuthService.UpdateRoleById(c.Request.Context(), id, req, lang)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrRoleNotFound:         http.StatusNotFound,
			domain.ErrSystemRoleProtected:  http.StatusForbidden,
			domain.ErrRoleNameDuplicate:    http.StatusConflict,
			domain.ErrPermissionNotFound:   http.StatusBadRequest,
			domain.ErrEmptyRolePermissions: http.StatusBadRequest,
		})
		return
	}

	utils.Success(c, role)
}

// DeleteRoleById godoc
// @Summary      Xóa role theo ID
// @Description  Xóa một vai trò (role) khỏi hệ thống theo ID (không thể xóa vai trò hệ thống hoặc vai trò đang được gán cho người dùng)
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của role cần xóa"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/roles/{id} [delete]
func (h *AuthHandler) DeleteRoleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	lang := utils.GetLanguageFromContext(c)
	if err := h.AuthService.DeleteRoleById(c.Request.Context(), id, lang); err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrRoleNotFound:           http.StatusNotFound,
			domain.ErrSystemRoleCannotDelete: http.StatusForbidden,
			domain.ErrRoleInUse:              http.StatusConflict,
		})
		return
	}

	utils.Success(c, gin.H{
		"message": utils.Translate(c, "Xóa vai trò thành công"),
	})
}

// UpdateRoleStatus godoc
// @Summary      Chuyển trạng thái role theo ID
// @Description  Chuyển đổi (toggle) hoặc cập nhật trạng thái hoạt động (active/inactive) của một vai trò (role) theo ID
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true   "ID của role"
// @Param        request  body      dto.UpdateRoleStatusRequest  false  "Trạng thái mới (nếu không truyền is_active, hệ thống sẽ tự động đảo ngược trạng thái hiện tại)"
// @Success      200      {object}  domain.Role
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/roles/{id}/status [patch]
func (h *AuthHandler) UpdateRoleStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var req dto.UpdateRoleStatusRequest
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	lang := utils.GetLanguageFromContext(c)
	role, err := h.AuthService.UpdateRoleStatus(c.Request.Context(), id, req, lang)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrRoleNotFound:               http.StatusNotFound,
			domain.ErrSystemRoleProtected:        http.StatusForbidden,
			domain.ErrSystemRoleCannotDeactivate: http.StatusForbidden,
		})
		return
	}

	utils.Success(c, role)
}

// DeactivateRoleById godoc
// @Summary      Hủy kích hoạt role theo ID
// @Description  Hủy kích hoạt một vai trò (role) trong hệ thống theo ID (không thể hủy kích hoạt vai trò hệ thống)
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của role cần hủy kích hoạt"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/roles/{id}/deactivate [patch]
func (h *AuthHandler) DeactivateRoleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	lang := utils.GetLanguageFromContext(c)
	if err := h.AuthService.DeactivateRoleById(c.Request.Context(), id, lang); err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrRoleNotFound:               http.StatusNotFound,
			domain.ErrSystemRoleProtected:        http.StatusForbidden,
			domain.ErrSystemRoleCannotDeactivate: http.StatusForbidden,
		})
		return
	}

	utils.Success(c, gin.H{
		"message": utils.Translate(c, "Hủy kích hoạt vai trò thành công"),
	})
}

// ListUsers godoc
// @Summary      Danh sách users (phân trang limit & offset)
// @Description  Lấy danh sách người dùng kèm vai trò (role), hỗ trợ tìm kiếm và phân trang bằng limit & offset
// @Tags         auth
// @Produce      json
// @Param        limit   query     int     false  "Số bản ghi cần lấy (default: 10, max: 50)"
// @Param        offset  query     int     false  "Số bản ghi cần bỏ qua (default: 0)"
// @Param        search  query     string  false  "Từ khóa tìm kiếm (theo username, email hoặc họ và tên)"
// @Success      200     {object}  dto.UserPaginationResponse
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/users [get]
func (h *AuthHandler) ListUsers(c *gin.Context) {
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

	resp, err := h.AuthService.GetUsersPaging(c.Request.Context(), limit, offset, search)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// CreateUser godoc
// @Summary      Tạo người dùng mới
// @Description  Tạo tài khoản người dùng mới cùng với thông tin đăng nhập và lưu lịch sử mật khẩu theo chuẩn OWASP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateUserRequest  true  "Thông tin người dùng mới"
// @Success      201      {object}  utils.Response{data=domain.User}
// @Failure      400      {object}  utils.Response
// @Failure      409      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Security     BearerAuth
// @Router       /auth/users [post]
func (h *AuthHandler) CreateUser(c *gin.Context) {
	lang := utils.GetLanguageFromContext(c)
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	createdBy := c.GetString("username")
	if createdBy == "" {
		createdBy = "system"
	}

	user, err := h.AuthService.CreateUser(c.Request.Context(), req, lang, createdBy)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrUsernameDuplicate: http.StatusConflict,
			domain.ErrEmailDuplicate:    http.StatusConflict,
			domain.ErrRoleNotFound:      http.StatusBadRequest,
			domain.ErrPasswordTooShort:  http.StatusBadRequest,
			domain.ErrUsernameInvalid:   http.StatusBadRequest,
			domain.ErrEmailInvalid:      http.StatusBadRequest,
			domain.ErrFullNameInvalid:   http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusCreated, utils.Response{
		Code:    http.StatusCreated,
		Message: utils.Translate(c, "Tạo người dùng thành công"),
		Data:    user,
	})
}

// GetUserById godoc
// @Summary      Chi tiết người dùng theo ID
// @Description  Lấy thông tin chi tiết người dùng kèm vai trò (role) theo ID
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của người dùng"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/users/{id} [get]
func (h *AuthHandler) GetUserById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID người dùng không hợp lệ")
		return
	}

	user, err := h.AuthService.GetUserById(c.Request.Context(), id)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrUserNotFound: http.StatusNotFound,
		})
		return
	}

	utils.Success(c, user)
}

// UpdateUserById godoc
// @Summary      Cập nhật thông tin người dùng theo ID
// @Description  Chỉnh sửa thông tin hồ sơ, vai trò hoặc trạng thái hoạt động của người dùng theo chuẩn OWASP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "ID của người dùng cần cập nhật"
// @Param        request  body      dto.UpdateUserRequest  true  "Thông tin cập nhật người dùng"
// @Success      200      {object}  domain.User
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/users/{id} [put]
func (h *AuthHandler) UpdateUserById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID người dùng không hợp lệ")
		return
	}

	lang := utils.GetLanguageFromContext(c)
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updatedBy := c.GetString("username")
	if updatedBy == "" {
		updatedBy = "system"
	}

	user, err := h.AuthService.UpdateUserById(c.Request.Context(), id, req, lang, updatedBy)
	if err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrUserNotFound:      http.StatusNotFound,
			domain.ErrUsernameDuplicate: http.StatusConflict,
			domain.ErrEmailDuplicate:    http.StatusConflict,
			domain.ErrRoleNotFound:      http.StatusBadRequest,
			domain.ErrUsernameInvalid:   http.StatusBadRequest,
			domain.ErrEmailInvalid:      http.StatusBadRequest,
			domain.ErrFullNameInvalid:   http.StatusBadRequest,
		})
		return
	}

	utils.Success(c, user)
}

// DeleteUserById godoc
// @Summary      Xóa người dùng theo ID
// @Description  Xóa mềm (soft delete) tài khoản người dùng theo ID theo chuẩn OWASP (không cho phép xóa tài khoản quản trị hệ thống hoặc tự xóa tài khoản của chính mình)
// @Tags         auth
// @Produce      json
// @Param        id   path      int  true  "ID của người dùng cần xóa"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/users/{id} [delete]
func (h *AuthHandler) DeleteUserById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.Error(c, http.StatusBadRequest, "ID người dùng không hợp lệ")
		return
	}

	deletedBy := c.GetString("username")
	if deletedBy == "" {
		deletedBy = "system"
	}

	lang := utils.GetLanguageFromContext(c)
	if err := h.AuthService.DeleteUserById(c.Request.Context(), id, lang, deletedBy); err != nil {
		utils.HandleDomainError(c, err, map[error]int{
			domain.ErrUserNotFound:           http.StatusNotFound,
			domain.ErrSystemUserCannotDelete: http.StatusForbidden,
			domain.ErrCannotDeleteSelf:       http.StatusForbidden,
		})
		return
	}

	utils.Success(c, gin.H{
		"message": utils.Translate(c, "Xóa người dùng thành công"),
	})
}
