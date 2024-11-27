package request

// CreateGroupRequest 创建群聊请求
type CreateGroupRequest struct {
	Name           string   `json:"name" binding:"required,min=1,max=100"`
	InitialMembers []uint64 `json:"initial_members" binding:"required,min=1,dive,min=1"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ContentType string `json:"content_type" binding:"required,oneof=text image file system"`
	Content     string `json:"content" binding:"required"`
}

// UpdateRoomRequest 更新聊天室请求
type UpdateRoomRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=100"`
	Announcement string `json:"announcement" binding:"max=500"`
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=member admin"`
}

// GetMessagesRequest 获取消息请求
type GetMessagesRequest struct {
	BeforeID uint64 `form:"before_id" binding:"omitempty,min=1"`
	Limit    int    `form:"limit" binding:"required,min=1,max=50"`
}
