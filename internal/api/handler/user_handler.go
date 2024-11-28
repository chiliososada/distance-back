package handler

import (
	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/service"
	"github.com/chiliososada/distance-back/pkg/auth"
	"github.com/chiliososada/distance-back/pkg/logger"

	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterUser 处理用户注册请求
// @Summary 用户注册
// @Description 注册新用户或更新已有用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Firebase ID Token"
// @Param request body request.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *Handler) RegisterUser(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("请求数据绑定失败", logger.Any("error", err))
		Error(c, service.ErrInvalidRequest)
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		logger.Error("未找到Authorization header")
		Error(c, service.ErrUnauthorized)
		return
	}

	decodedToken, err := auth.VerifyIDToken(c.Request.Context(), strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		logger.Error("Firebase token验证失败", logger.Any("error", err))
		Error(c, service.ErrUnauthorized)
		return
	}

	authUser := auth.NewAuthUserFromToken(decodedToken)
	user, err := h.userService.RegisterOrUpdateUser(c, authUser)
	if err != nil {
		logger.Error("用户注册失败",
			logger.Any("error", err),
			logger.Any("auth_user", authUser))
		Error(c, err)
		return
	}

	Success(c, response.ToResponse(user))
}

// GetProfile 获取当前用户的个人资料
// @Summary 获取个人资料
// @Description 获取当前登录用户的详细资料
// @Tags 用户管理
// @Produce json
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/users/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	user, err := h.userService.GetUserByID(c, userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToResponse(user))
}

// UpdateProfile 更新用户个人资料
// @Summary 更新个人资料
// @Description 更新当前登录用户的个人资料信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body request.UpdateProfileRequest true "更新资料请求"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 400,401 {object} response.ErrorResponse
// @Router /api/v1/users/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	profile := &model.User{
		Nickname:        req.Nickname,
		Bio:             req.Bio,
		Gender:          req.Gender,
		Language:        req.Language,
		PrivacyLevel:    req.PrivacyLevel,
		LocationSharing: req.LocationSharing,
		PhotoEnabled:    req.PhotoEnabled,
	}

	if err := h.userService.UpdateProfile(c, userID, profile); err != nil {
		Error(c, err)
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(c, userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToResponse(updatedUser))
}

// UpdateAvatar 更新用户头像
// @Summary 更新用户头像
// @Description 上传并更新当前登录用户的头像
// @Tags 用户管理
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "头像文件"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 400,401 {object} response.ErrorResponse
// @Router /api/v1/users/avatar [put]
func (h *Handler) UpdateAvatar(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	avatarFile := &model.File{
		File: file,
		Type: "avatar",
	}

	if err := h.userService.UpdateAvatar(c, userID, avatarFile); err != nil {
		Error(c, err)
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(c, userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToResponse(updatedUser))
}

// UpdateLocation 更新用户位置信息
// @Summary 更新位置信息
// @Description 更新当前登录用户的地理位置信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body request.UpdateLocationRequest true "位置信息"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 400,401 {object} response.ErrorResponse
// @Router /api/v1/users/location [put]
func (h *Handler) UpdateLocation(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	var req request.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.userService.UpdateLocation(c, userID, req.Latitude, req.Longitude); err != nil {
		Error(c, err)
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(c, userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToResponse(updatedUser))
}

// GetUserProfile 获取指定用户的资料
// @Summary 获取用户资料
// @Description 获取指定用户ID的用户资料
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=response.UserResponse}
// @Failure 400,401,403,404 {object} response.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
	targetID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	currentUserID := h.GetCurrentUserID(c)
	user, err := h.userService.GetUserByID(c, targetID)
	if err != nil {
		Error(c, err)
		return
	}

	if user == nil {
		Error(c, service.ErrUserNotFound)
		return
	}

	// 检查隐私设置
	if user.PrivacyLevel != "public" && currentUserID != targetID {
		// 检查是否是好友
		isFriend, err := h.relationshipService.IsFriend(c, currentUserID, targetID)
		if err != nil || !isFriend {
			Error(c, service.ErrForbidden)
			return
		}
	}

	Success(c, response.ToResponse(user))
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 根据关键词搜索用户
// @Tags 用户管理
// @Produce json
// @Param request query request.SearchUserRequest true "搜索请求"
// @Success 200 {object} response.Response{data=response.PaginatedResponse{list=[]response.UserResponse}}
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/users/search [get]
func (h *Handler) SearchUsers(c *gin.Context) {
	var req request.SearchUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	users, total, err := h.userService.SearchUsers(c, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	// 转换为响应格式
	userResponses := make([]*response.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = response.ToResponse(user)
	}

	Success(c, response.NewPaginated(userResponses, total, req.Page, req.PageSize))
}

// GetNearbyUsers 获取附近的用户
// @Summary 获取附近用户
// @Description 获取指定位置附近的用户列表
// @Tags 用户管理
// @Produce json
// @Param request query request.NearbyUsersRequest true "附近用户请求"
// @Success 200 {object} response.Response{data=response.PaginatedResponse{list=[]response.UserResponse}}
// @Failure 400 {object} response.ErrorResponse
// @Router /api/v1/users/nearby [get]
func (h *Handler) GetNearbyUsers(c *gin.Context) {
	var req request.NearbyUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	users, total, err := h.userService.GetNearbyUsers(c, req.Latitude, req.Longitude, req.Radius, req.Page, req.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	// 转换为响应格式
	userResponses := make([]*response.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = response.ToResponse(user)
	}

	Success(c, response.NewPaginated(userResponses, total, req.Page, req.PageSize))
}

// RegisterDevice 注册用户设备
// @Summary 注册设备
// @Description 注册用户的设备信息用于消息推送
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param device body model.UserDevice true "设备信息"
// @Success 200 {object} response.Response
// @Failure 400,401 {object} response.ErrorResponse
// @Router /api/v1/users/devices [post]
func (h *Handler) RegisterDevice(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	var device model.UserDevice
	if err := c.ShouldBindJSON(&device); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.userService.RegisterDevice(c, userID, &device); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}
