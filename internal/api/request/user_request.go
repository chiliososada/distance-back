package request

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Nickname    string `json:"nickname" binding:"required,min=2,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=32"`
	DeviceType  string `json:"device_type" binding:"required,oneof=ios android web"`
	DeviceToken string `json:"device_token" binding:"required"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Nickname     string `json:"nickname" binding:"omitempty,min=2,max=50"`
	Bio          string `json:"bio" binding:"omitempty,max=500"`
	Gender       string `json:"gender" binding:"omitempty,oneof=male female other"`
	BirthDate    string `json:"birth_date" binding:"omitempty,datetime=2006-01-02"`
	Language     string `json:"language" binding:"omitempty,len=5"`
	PrivacyLevel string `json:"privacy_level" binding:"omitempty,oneof=public friends private"`
}

// UpdateLocationRequest 更新位置请求
type UpdateLocationRequest struct {
	Location
	LocationSharing bool `json:"location_sharing"`
}

// SearchUserRequest 搜索用户请求
type SearchUserRequest struct {
	Pagination
	Keyword string `json:"keyword" form:"keyword" binding:"required,min=1,max=50"`
}

// NearbyUsersRequest 查询附近用户请求
type NearbyUsersRequest struct {
	Pagination
	Location
}
