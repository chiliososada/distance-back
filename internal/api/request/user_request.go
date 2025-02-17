package request

import "time"

type LoginRequest struct {
	IdToken string `json:"id_token" binding:"required"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Nickname   string     `json:"nickname" binding:"required,min=2,max=50"`
	BirthDate  string     `json:"birth_date" binding:"required,datetime=2006-01-02"`
	Bio        string     `json:"bio" binding:"omitempty,max=500"`
	Gender     string     `json:"gender" binding:"omitempty,oneof=male female other"`
	Language   string     `json:"language" binding:"omitempty,len=5"`
	DeviceInfo DeviceInfo `json:"device_info" binding:"required"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceType   string `json:"device_type" binding:"required,oneof=ios android web"`
	DeviceToken  string `json:"device_token" binding:"required"`
	DeviceName   string `json:"device_name" binding:"required,max=100"`
	DeviceModel  string `json:"device_model" binding:"omitempty,max=50"`
	OSVersion    string `json:"os_version" binding:"omitempty,max=20"`
	AppVersion   string `json:"app_version" binding:"required,max=20"`
	PushProvider string `json:"push_provider" binding:"required,oneof=fcm apns web"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Email               *string    `json:"email" binding:"required,min=1"`
	Nickname            *string    `json:"nickname" binding:"omitempty,min=2,max=50"`
	Bio                 *string    `json:"bio" binding:"omitempty,max=500"`
	Gender              *string    `json:"gender" binding:"omitempty,oneof=male female other"`
	BirthDate           *time.Time `json:"birth_date" binding:"omitempty"`
	Language            *string    `json:"language" binding:"omitempty,len=5"` // 如: zh_CN
	PrivacyLevel        *string    `json:"privacy_level" binding:"omitempty,oneof=public friends private"`
	LocationSharing     *bool      `json:"location_sharing,omitempty"`
	PhotoEnabled        *bool      `json:"photo_enabled,omitempty"`
	NotificationEnabled *bool      `json:"notification_enabled,omitempty"`
	AvatarURL           *string    `json:"avatar_url,omitempty"`
}

// UpdateLocationRequest 更新位置请求
type UpdateLocationRequest struct {
	LocationPerson
	LocationSharing bool `json:"location_sharing"`
}

// SearchUsersRequest 搜索用户请求
type SearchUsersRequest struct {
	PaginationQuery
	SortQuery
	Keyword string `form:"keyword" binding:"required,min=1,max=50" json:"keyword"`
}

// NearbyUsersRequest 查询附近用户请求
type NearbyUsersRequest struct {
	PaginationQuery
	LocationQuery
}

// UserDeviceRequest 用户设备注册请求
type UserDeviceRequest struct {
	DeviceToken  string `json:"device_token" binding:"required"`
	DeviceType   string `json:"device_type" binding:"required,oneof=ios android web"`
	DeviceName   string `json:"device_name" binding:"required,max=100"`
	DeviceModel  string `json:"device_model" binding:"max=50"`
	OSVersion    string `json:"os_version" binding:"max=20"`
	AppVersion   string `json:"app_version" binding:"required,max=20"`
	BrowserInfo  string `json:"browser_info" binding:"omitempty,max=200"`
	PushProvider string `json:"push_provider" binding:"required,oneof=fcm apns web"`
	PushEnabled  *bool  `json:"push_enabled,omitempty"`
}

// GetUserRequest 获取用户信息请求
type GetUserRequest struct {
	UIDParam
	IncludeDeleted bool `form:"include_deleted" json:"include_deleted"` // 是否包含已删除的用户
}

// GetUserByFirebaseRequest 根据Firebase UID获取用户请求
type GetUserByFirebaseRequest struct {
	FirebaseUID string `json:"firebase_uid" binding:"required"`
}

// BatchGetUsersRequest 批量获取用户信息请求
type BatchGetUsersRequest struct {
	UIDs []string `json:"uids" binding:"required,min=1,max=100,dive,uuid"` // 最多100个用户
}
