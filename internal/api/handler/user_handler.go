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
// @Summary 用户注册或更新
// @Description 注册新用户或更新已有用户信息（Firebase认证后）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Firebase ID Token"
// @Param request body request.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=response.UserInfo}
// @Failure 400,401,500 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *Handler) RegisterUser(c *gin.Context) {
	// 1. 验证 Firebase token
	token := c.GetHeader("Authorization")
	if token == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// 2. 解析 token
	decodedToken, err := auth.VerifyIDToken(c.Request.Context(), strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		logger.Error("Failed to verify Firebase token",
			logger.Any("error", err))
		Error(c, errors.ErrTokenInvalid)
		return
	}

	// 3. 绑定并验证请求参数
	var req request.RegisterRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 4. 创建认证用户对象
	authUser := auth.NewAuthUserFromToken(decodedToken)

	// 5. 补充用户信息
	if authUser.DisplayName == "" {
		authUser.DisplayName = req.Nickname
	}

	// 6. 注册或更新用户
	user, err := h.userService.RegisterOrUpdateUser(c.Request.Context(), authUser)
	if err != nil {
		logger.Error("Failed to register/update user",
			logger.Any("auth_user", authUser),
			logger.Any("error", err))
		Error(c, errors.ErrAuthentication)
		return
	}
	// 解析出生日期
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		Error(c, errors.New(errors.CodeValidation, "Invalid birth date format"))
		return
	}

	// 更新用户的其他信息
	profileUpdate := &model.User{
		Nickname:  req.Nickname,
		Gender:    req.Gender,
		BirthDate: &birthDate,
		Language:  req.Language,
	}

	if err := h.userService.UpdateProfile(c.Request.Context(), user.UID, profileUpdate); err != nil {
		Error(c, err)
		return
	}
	// 7. 注册设备
	device := &model.UserDevice{
		UserUID:      user.UID,
		DeviceToken:  req.DeviceInfo.DeviceToken,
		DeviceType:   req.DeviceInfo.DeviceType,
		DeviceName:   req.DeviceInfo.DeviceName,
		DeviceModel:  req.DeviceInfo.DeviceModel,
		OSVersion:    req.DeviceInfo.OSVersion,
		AppVersion:   req.DeviceInfo.AppVersion,
		PushProvider: req.DeviceInfo.PushProvider,
		PushEnabled:  true,
		IsActive:     true,
		LastActiveAt: time.Now(),
	}

	if err := h.userService.RegisterDevice(c.Request.Context(), user.UID, device); err != nil {
		logger.Warn("Failed to register device",
			logger.String("user_uid", user.UID),
			logger.Any("error", err))
		// 不返回错误，因为设备注册失败不应影响用户注册
	}
	// 获取最新的用户信息
	updatedUser, err := h.userService.GetUserByUID(c.Request.Context(), user.UID)
	if err != nil {
		Error(c, err)
		return
	}

	// 8. 返回用户信息
	Success(c, response.ToUserInfo(updatedUser))
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
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	user, err := h.userService.GetUserByUID(c.Request.Context(), userUID)
	if err != nil {
		Error(c, errors.ErrUserProfileFetchFailed.WithDetails(err.Error()))
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
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	logger.Info("Updating profile for user", logger.String("uid", userUID))

	var req request.UpdateProfileRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 获取现有用户
	currentUser, err := h.userService.GetUserByUID(c.Request.Context(), userUID)
	if err != nil {
		logger.Error("Failed to get user",
			logger.String("uid", userUID),
			logger.Any("error", err))
		Error(c, err)
		return
	}
	if currentUser == nil {
		Error(c, errors.ErrUserNotFound)
		return
	}

	logger.Info("Found existing user", logger.Any("user", currentUser))

	// 只更新请求中的字段
	if req.Nickname != "" {
		currentUser.Nickname = req.Nickname
	}
	if req.Bio != "" {
		currentUser.Bio = req.Bio
	}
	if req.Gender != "" {
		currentUser.Gender = req.Gender
	}
	if req.Language != "" {
		currentUser.Language = req.Language
	}
	if req.PrivacyLevel != "" {
		currentUser.PrivacyLevel = req.PrivacyLevel
	}
	if req.LocationSharing != nil {
		currentUser.LocationSharing = *req.LocationSharing
	}
	if req.PhotoEnabled != nil {
		currentUser.PhotoEnabled = *req.PhotoEnabled
	}
	if req.NotificationEnabled != nil {
		currentUser.NotificationEnabled = *req.NotificationEnabled
	}
	if req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			Error(c, errors.New(errors.CodeValidation, "Invalid birth date format"))
			return
		}
		currentUser.BirthDate = &birthDate
	}

	logger.Info("Updating user with data", logger.Any("updated_user", currentUser))

	// 更新用户信息
	if err := h.userService.UpdateProfile(c.Request.Context(), userUID, currentUser); err != nil {
		logger.Error("Failed to update user",
			logger.String("uid", userUID),
			logger.Any("error", err))
		Error(c, err)
		return
	}

	// 获取更新后的信息
	updatedUser, err := h.userService.GetUserByUID(c.Request.Context(), userUID)
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
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		Error(c, errors.ErrInvalidFileUpload)
		return
	}

	avatar := &model.File{
		File: file,
		Type: "avatar",
		Name: file.Filename,
		Size: uint(file.Size),
	}

	if err := h.userService.UpdateAvatar(c.Request.Context(), userUID, avatar); err != nil {
		Error(c, errors.ErrAvatarUpdateFailed.WithDetails(err.Error()))
		return
	}

	// 获取更新后的用户信息
	updatedUser, err := h.userService.GetUserByUID(c.Request.Context(), userUID)
	if err != nil {
		Error(c, errors.ErrUserNotFound)
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
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
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

	updatedUser, err := h.userService.UpdateLocation(c.Request.Context(), userUID, req.Latitude, req.Longitude)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, response.ToUserInfo(updatedUser))
}

// GetUserProfile 获取指定用户资料
// @Summary 获取用户资料
// @Description 获取指定用户的资料信息
// @Tags 用户管理
// @Produce json
// @Param uid path string true "用户UID"
// @Success 200 {object} response.Response{data=response.UserProfile}
// @Failure 400,404 {object} response.Response
// @Router /api/v1/users/{uid} [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
	targetUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	currentUserUID := h.GetCurrentUserUID(c)
	user, err := h.userService.GetUserByUID(c.Request.Context(), targetUID)
	if err != nil {
		Error(c, errors.ErrUserProfileUpdateFailed.WithDetails(err.Error()))
		return
	}

	if user == nil {
		Error(c, errors.ErrUserNotFound)
		return
	}

	// 检查隐私设置
	if user.PrivacyLevel != "public" && currentUserUID != targetUID {
		// 检查是否是好友
		isFriend, err := h.relationshipService.IsFriend(c.Request.Context(), currentUserUID, targetUID)
		if err != nil || !isFriend {
			Error(c, errors.ErrRelationExists)
			return
		}
	}

	// 构建关系信息
	relationship := &response.Relationship{}
	if currentUserUID != "" && currentUserUID != targetUID {
		isFollowing, _ := h.relationshipService.IsFollowing(c.Request.Context(), currentUserUID, targetUID)
		isFollowed, _ := h.relationshipService.IsFollowed(c.Request.Context(), currentUserUID, targetUID)
		isFriend, _ := h.relationshipService.IsFriend(c.Request.Context(), currentUserUID, targetUID)
		isRejected, _ := h.relationshipService.IsRejected(c.Request.Context(), currentUserUID, targetUID)

		relationship = &response.Relationship{
			IsFollowing: isFollowing,
			IsFollowed:  isFollowed,
			IsFriend:    isFriend,
			IsRejected:  isRejected,
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
		Error(c, errors.ErrSearchFailed)
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
	// 获取当前用户UID
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// 参数验证
	var req request.UserDeviceRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 验证设备令牌不能为空
	if req.DeviceToken == "" {
		Error(c, errors.ErrDeviceTokenRequired)
		return
	}

	// 构造设备信息
	device := &model.UserDevice{
		DeviceToken:  req.DeviceToken,
		DeviceType:   req.DeviceType,
		DeviceName:   req.DeviceName,
		DeviceModel:  req.DeviceModel,
		OSVersion:    req.OSVersion,
		AppVersion:   req.AppVersion,
		BrowserInfo:  req.BrowserInfo,
		PushProvider: req.PushProvider,
		PushEnabled:  true,
		IsActive:     true,
	}

	// 处理可选的推送开关设置
	if req.PushEnabled != nil {
		device.PushEnabled = *req.PushEnabled
	}

	// 注册设备
	if err := h.userService.RegisterDevice(c.Request.Context(), userUID, device); err != nil {
		logger.Error("注册设备失败",
			logger.String("user_uid", userUID),
			logger.String("device_token", req.DeviceToken),
			logger.Any("error", err))
		Error(c, errors.ErrDeviceRegisterFailed)
		return
	}

	Success(c, nil)
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
		Error(c, errors.ErrNearbyUsersFailed)
		return
	}

	// 转换响应
	userResponses := make([]*response.UserInfo, len(users))
	for i, user := range users {
		userResponses[i] = response.ToUserInfo(user)
	}

	Success(c, response.NewPaginatedResponse(userResponses, total, req.Page, req.Size))
}

// ListUsers 获取用户列表（管理员接口）
// @Summary 获取用户列表
// @Description 分页获取用户列表（仅管理员可用）
// @Tags 用户管理
// @Produce json
// @Param page query int true "页码"
// @Param size query int true "每页数量"
// @Success 200 {object} response.Response{data=response.UserListResponse}
// @Failure 400,401,403 {object} response.Response
// @Router /api/v1/admin/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	// 验证管理员权限
	adminUID := h.GetCurrentUserUID(c)
	if adminUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// TODO: 添加管理员权限验证

	page, size, err := GetPagination(c)
	if err != nil {
		Error(c, err)
		return
	}

	users, total, err := h.userService.ListUsers(c.Request.Context(), page, size)
	if err != nil {
		Error(c, errors.ErrUserListFailed)
		return
	}

	// 转换响应
	userResponses := make([]*response.UserInfo, len(users))
	for i, user := range users {
		userResponses[i] = response.ToUserInfo(user)
	}

	Success(c, response.NewPaginatedResponse(userResponses, total, page, size))
}
