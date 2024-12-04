package request

// CreateGroupRequest 创建群聊请求
type CreateGroupRequest struct {
	Name           string   `json:"name" binding:"required,min=1,max=100"`
	InitialMembers []string `json:"initial_members" binding:"required,min=1,dive,uuid"`
	Announcement   string   `json:"announcement" binding:"max=500"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ContentType string `json:"content_type" binding:"required,oneof=text image file system"`
	Content     string `json:"content" binding:"required"`
}

// UpdateRoomRequest 更新聊天室请求
type UpdateRoomRequest struct {
	Name         string `json:"name" binding:"omitempty,min=1,max=100"`
	Announcement string `json:"announcement" binding:"max=500"`
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserUID string `json:"user_uid" binding:"required,uuid"`
	Role    string `json:"role" binding:"required,oneof=member admin"`
}

// GetMessagesRequest 获取消息请求
type GetMessagesRequest struct {
	PaginationQuery
	BeforeUID string `form:"before_uid" binding:"omitempty,uuid"`
}

// UpdateMemberRequest 更新成员请求
type UpdateMemberRequest struct {
	Role     string `json:"role" binding:"required,oneof=member admin"`
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	IsMuted  *bool  `json:"is_muted,omitempty"`
}

// UpdateMemberRoleRequest 更新成员角色请求
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin member"`
}

// GetRoomMembersRequest 获取聊天室成员列表请求
type GetRoomMembersRequest struct {
	PaginationQuery
	RoomUID string `form:"room_uid" binding:"required,uuid"`
}

// GetUnreadCountRequest 获取未读消息数请求
type GetUnreadCountRequest struct {
	RoomUID string `form:"room_uid" binding:"required,uuid"`
}

// MarkMessagesAsReadRequest 标记消息已读请求
type MarkMessagesAsReadRequest struct {
	RoomUID    string `json:"room_uid" binding:"required,uuid"`
	MessageUID string `json:"message_uid" binding:"required,uuid"`
}

// PinRoomRequest 置顶聊天室请求
type PinRoomRequest struct {
	RoomUID string `json:"room_uid" binding:"required,uuid"`
}

// GetRoomInfoRequest 获取聊天室信息请求
type GetRoomInfoRequest struct {
	UIDParam
	IncludeMembers bool `form:"include_members" json:"include_members"`
}

// ListRoomsRequest 获取聊天室列表请求
type ListRoomsRequest struct {
	PaginationQuery
	Type     string `form:"type" binding:"omitempty,oneof=individual group merchant official"`
	Status   string `form:"status" binding:"omitempty,oneof=active closed cancelled"`
	IsPinned *bool  `form:"is_pinned" json:"is_pinned,omitempty"`
}

// BatchGetRoomsRequest 批量获取聊天室请求
type BatchGetRoomsRequest struct {
	RoomUIDs []string `json:"room_uids" binding:"required,min=1,max=100,dive,uuid"`
}

// SearchMessagesRequest 搜索消息请求
type SearchMessagesRequest struct {
	PaginationQuery
	RoomUID string `form:"room_uid" binding:"required,uuid"`
	Keyword string `form:"keyword" binding:"required,min=1,max=50"`
}
