package response

import "time"

// ChatRoomResponse 聊天室响应
type ChatRoomResponse struct {
	ID           uint64        `json:"id"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	AvatarURL    string        `json:"avatar_url"`
	Announcement string        `json:"announcement"`
	MembersCount int           `json:"members_count"`
	UnreadCount  int64         `json:"unread_count"`
	LastMessage  *MessageBrief `json:"last_message"`
	CreatedAt    time.Time     `json:"created_at"`
	IsPinned     bool          `json:"is_pinned"`
}

// ChatMemberResponse 聊天室成员响应
type ChatMemberResponse struct {
	UserBrief
	Role     string    `json:"role"`
	Nickname string    `json:"nickname"`
	JoinedAt time.Time `json:"joined_at"`
	IsMuted  bool      `json:"is_muted"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID          uint64         `json:"id"`
	ChatRoomID  uint64         `json:"chat_room_id"`
	Sender      UserBrief      `json:"sender"`
	ContentType string         `json:"content_type"`
	Content     string         `json:"content"`
	Media       []MessageMedia `json:"media,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// MessageBrief 消息简要信息
type MessageBrief struct {
	ContentType string    `json:"content_type"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

// MessageMedia 消息媒体信息
type MessageMedia struct {
	ID   uint64 `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
	Size uint   `json:"size"`
}
