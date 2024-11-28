package handler

import (
	"strings"
	"time"

	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/pkg/auth"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RegisterUser 处理用户注册
// @Summary 用户注册
// @Description 注册新用户或更新已有用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Firebase ID Token"
// @Param request body request.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=response.UserInfo}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *Handler) RegisterUser(c *gin.Context) {
	// Firebase token 验证
	token := c.GetHeader("Authorization")
	if token == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	decodedToken, err := auth.VerifyIDToken(c.Request.Context(), strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		logger.Error("Firebase token验证失败", logger.Any("error", err))
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.RegisterRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	authUser := auth.NewAuthUserFromToken(decodedToken)
	user, err := h.userService.RegisterOrUpdateUser(c.Request.Context(), authUser)
	if err != nil {
		logger.Error("用户注册失败",
			logger.Any("error", err),
			logger.Any("auth_user", authUser))
		Error(c, err)
		return
	}

	Success(c, response.ToUserInfo(user))
}

// GetProfile 获取个人资料
// @Summary 获取个人资料
// @Description 获取当前登录用户的详细资料
// @Tags 用户管理
// @Produce json
// @Success 200 {object} response.Response{data=response.UserProfile}
// @Failure 401 {object} response.Response
// @Router /api/v1/users/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToUserProfile(user))
}

// UpdateProfile 更新个人资料
// @Summary 更新个人资料
// @Description 更新当前登录用户的个人资料信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body request.UpdateProfileRequest true "更新资料请求"
// @Success 200 {object} response.Response{data=response.UserInfo}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/users/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.New(errors.CodeAuthentication, "Unauthorized"))
		return
	}

	var req request.UpdateProfileRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 将请求转换为模型
	profile := &model.User{
		BaseModel:    model.BaseModel{ID: userID},
		Nickname:     req.Nickname,
		Bio:          req.Bio,
		Gender:       req.Gender,
		Language:     req.Language,
		PrivacyLevel: req.PrivacyLevel,
		Status:       "active",     // 维持当前状态
		UserType:     "individual", // 维持当前类型
	}

	// 处理可选的布尔值字段
	if req.LocationSharing != nil {
		profile.LocationSharing = *req.LocationSharing
	}
	if req.PhotoEnabled != nil {
		profile.PhotoEnabled = *req.PhotoEnabled
	}
	if req.NotificationEnabled != nil {
		profile.NotificationEnabled = *req.NotificationEnabled
	}

	// 处理生日（如果提供）
	if req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			Error(c, errors.New(errors.CodeValidation, "Invalid birth date format"))
			return
		}
		profile.BirthDate = &birthDate
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), userID, profile); err != nil {
		Error(c, err)
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToUserInfo(updatedUser))
}

// UpdateAvatar 更新用户头像
// @Summary 更新用户头像
// @Description 上传并更新当前登录用户的头像
// @Tags 用户管理
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "头像文件"
// @Success 200 {object} response.Response{data=response.UserInfo}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/users/avatar [put]
func (h *Handler) UpdateAvatar(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		Error(c, errors.New(errors.CodeValidation, "Invalid file"))
		return
	}

	avatarFile := &model.File{
		File: file,
		Type: "avatar",
		Name: file.Filename,
		Size: uint(file.Size),
	}

	if err := h.userService.UpdateAvatar(c.Request.Context(), userID, avatarFile); err != nil {
		Error(c, err)
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToUserInfo(updatedUser))
}

// UpdateLocation 更新位置信息
// @Summary 更新位置信息
// @Description 更新当前登录用户的地理位置信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body request.UpdateLocationRequest true "位置信息"
// @Success 200 {object} response.Response{data=response.UserInfo}
// @Failure 400,401 {object} response.Response
// @Router /api/v1/users/location [put]
func (h *Handler) UpdateLocation(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.New(errors.CodeAuthentication, "Unauthorized"))
		return
	}

	var req request.UpdateLocationRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 验证位置参数
	if err := ValidateLocation(req.Latitude, req.Longitude, 0); err != nil {
		Error(c, err)
		return
	}

	// 直接传递经纬度
	if err := h.userService.UpdateLocation(c.Request.Context(), userID, req.Latitude, req.Longitude); err != nil {
		Error(c, err)
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToUserInfo(updatedUser))
}

// GetUserProfile 获取指定用户资料
// @Summary 获取用户资料
// @Description 获取指定用户ID的用户资料
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=response.UserProfile}
// @Failure 400,401,403,404 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
	targetID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	currentUserID := h.GetCurrentUserID(c)
	user, err := h.userService.GetUserByID(c.Request.Context(), targetID)
	if err != nil {
		Error(c, err)
		return
	}

	if user == nil {
		Error(c, errors.ErrUserNotFound)
		return
	}

	// 检查隐私设置
	if user.PrivacyLevel != "public" && currentUserID != targetID {
		// 检查是否是好友
		isFriend, err := h.relationshipService.IsFriend(c.Request.Context(), currentUserID, targetID)
		if err != nil || !isFriend {
			Error(c, errors.ErrForbidden)
			return
		}
	}

	// 获取关系信息
	relationship := &response.Relationship{}
	if currentUserID != 0 && currentUserID != targetID {
		isFollowing, _ := h.relationshipService.IsFollowing(c.Request.Context(), currentUserID, targetID)
		isFollowed, _ := h.relationshipService.IsFollowed(c.Request.Context(), currentUserID, targetID)
		isFriend, _ := h.relationshipService.IsFriend(c.Request.Context(), currentUserID, targetID)
		isBlocked, _ := h.relationshipService.IsBlocked(c.Request.Context(), currentUserID, targetID)

		relationship = &response.Relationship{
			IsFollowing: isFollowing,
			IsFollowed:  isFollowed,
			IsFriend:    isFriend,
			IsBlocked:   isBlocked,
		}
	}

	profile := response.ToUserProfile(user)
	profile.Relationship = relationship

	Success(c, profile)
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 根据关键词搜索用户
// @Tags 用户管理
// @Produce json
// @Param request query request.SearchUsersRequest true "搜索请求"
// @Success 200 {object} response.Response{data=response.UserListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/users/search [get]
func (h *Handler) SearchUsers(c *gin.Context) {
	var req request.SearchUsersRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	users, total, err := h.userService.SearchUsers(
		c.Request.Context(),
		req.Keyword,
		req.Page,
		req.Size,
	)
	if err != nil {
		Error(c, err)
		return
	}

	// 转换响应
	userResponses := make([]*response.UserInfo, len(users))
	for i, user := range users {
		userResponses[i] = response.ToUserInfo(user)
	}

	Success(c, response.NewPaginatedResponse(userResponses, total, req.Page, req.Size))
}

// GetNearbyUsers 获取附近用户
// @Summary 获取附近用户
// @Description 获取指定位置附近的用户列表
// @Tags 用户管理
// @Produce json
// @Param request query request.NearbyUsersRequest true "附近用户请求"
// @Success 200 {object} response.Response{data=response.UserListResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/users/nearby [get]
func (h *Handler) GetNearbyUsers(c *gin.Context) {
	var req request.NearbyUsersRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 验证位置参数
	if err := ValidateLocation(req.Latitude, req.Longitude, req.Radius); err != nil {
		Error(c, err)
		return
	}

	users, total, err := h.userService.GetNearbyUsers(
		c.Request.Context(),
		req.Latitude,
		req.Longitude,
		req.Radius,
		req.Page,
		req.Size,
	)
	if err != nil {
		Error(c, err)
		return
	}

	// 转换响应
	userResponses := make([]*response.UserInfo, len(users))
	for i, user := range users {
		userResponses[i] = response.ToUserInfo(user)
	}

	Success(c, response.NewPaginatedResponse(userResponses, total, req.Page, req.Size))
}

// RegisterDevice 注册设备
// @Summary 注册设备
// @Description 注册用户的设备信息用于消息推送
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body request.UserDeviceRequest true "设备信息"
// @Success 200 {object} response.Response
// @Failure 400,401 {object} response.Response
// @Router /api/v1/users/devices [post]
func (h *Handler) RegisterDevice(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, errors.ErrUnauthorized)
		return
	}

	var req request.UserDeviceRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	device := &model.UserDevice{
		UserID:      userID,
		DeviceToken: req.DeviceToken,
		DeviceType:  req.DeviceType,
		DeviceName:  req.DeviceName,
		DeviceModel: req.DeviceModel,
		OSVersion:   req.OSVersion,
		AppVersion:  req.AppVersion,
		PushEnabled: true,
	}

	if req.PushEnabled != nil {
		device.PushEnabled = *req.PushEnabled
	}

	if err := h.userService.RegisterDevice(c.Request.Context(), userID, device); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}
