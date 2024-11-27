package handler

import (
	"DistanceBack_v1/internal/model"
	"DistanceBack_v1/internal/service"
	"DistanceBack_v1/pkg/auth"
	"DistanceBack_v1/pkg/logger"

	"github.com/gin-gonic/gin"
)

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Nickname        string `json:"nickname" binding:"required,min=2,max=50"`
	Bio             string `json:"bio" binding:"max=500"`
	Gender          string `json:"gender" binding:"omitempty,oneof=male female other"`
	BirthDate       string `json:"birth_date" binding:"omitempty,datetime=2006-01-02"`
	Language        string `json:"language" binding:"omitempty,min=2,max=10"`
	PrivacyLevel    string `json:"privacy_level" binding:"omitempty,oneof=public friends private"`
	LocationSharing bool   `json:"location_sharing"`
	PhotoEnabled    bool   `json:"photo_enabled"`
}

// UpdateLocationRequest 更新位置请求
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

// RegisterUser 注册用户
func (h *Handler) RegisterUser(c *gin.Context) {
	// 从上下文获取Firebase用户信息
	firebaseUser, exists := c.Get("firebase_user")
	if !exists {
		Error(c, service.ErrUnauthorized)
		return
	}

	user, err := h.userService.RegisterOrUpdateUser(c, firebaseUser.(*auth.AuthUser))
	if err != nil {
		logger.Error("Failed to register user",
			logger.Any("error", err),
			logger.Any("firebase_user", firebaseUser))
		Error(c, err)
		return
	}

	Success(c, user)
}

// GetProfile 获取个人资料
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

	Success(c, user)
}

// UpdateProfile 更新个人资料
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	var req UpdateProfileRequest
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

	Success(c, nil)
}

// UpdateAvatar 更新头像
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

	Success(c, nil)
}

// UpdateLocation 更新位置
func (h *Handler) UpdateLocation(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.userService.UpdateLocation(c, userID, req.Latitude, req.Longitude); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetUserProfile 获取用户资料
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

	Success(c, user)
}

// SearchUsers 搜索用户
func (h *Handler) SearchUsers(c *gin.Context) {
	var query struct {
		PaginationQuery
		Keyword string `form:"keyword" binding:"required,min=1,max=50"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	users, total, err := h.userService.SearchUsers(c, query.Keyword, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"users": users,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

// GetNearbyUsers 获取附近的用户
func (h *Handler) GetNearbyUsers(c *gin.Context) {
	var query struct {
		PaginationQuery
		Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90"`
		Longitude float64 `form:"longitude" binding:"required,min=-180,max=180"`
		Radius    float64 `form:"radius" binding:"required,min=0,max=50000"` // 最大50km
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	users, total, err := h.userService.GetNearbyUsers(c, query.Latitude, query.Longitude, query.Radius, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"users": users,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

// RegisterDevice 注册设备
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
