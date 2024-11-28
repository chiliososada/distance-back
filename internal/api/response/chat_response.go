package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// ChatRoomResponse 聊天室响应
type ChatRoomResponse struct {
	ID           uint64        `json:"id"`
	Name         string        `json:"name"`
	Type         string        `json:"type"` // individual/group
	AvatarURL    string        `json:"avatar_url"`
	Announcement string        `json:"announcement"`
	MembersCount int           `json:"members_count"`
	UnreadCount  int64         `json:"unread_count"`
	LastMessage  *MessageBrief `json:"last_message,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	IsPinned     bool          `json:"is_pinned"`
}

// ChatMemberResponse 聊天室成员响应
type ChatMemberResponse struct {
	UserBrief
	Role     string    `json:"role"` // owner/admin/member
	Nickname string    `json:"nickname"`
	JoinedAt time.Time `json:"joined_at"`
	IsMuted  bool      `json:"is_muted"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID          uint64         `json:"id"`
	ChatRoomID  uint64         `json:"chat_room_id"`
	Sender      UserBrief      `json:"sender"`
	ContentType string         `json:"content_type"` // text/image/file/system
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
	Type string `json:"type"` // image/file
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
	Size uint   `json:"size"`
}

// 使用泛型定义分页响应类型
type ChatRoomListResponse = PaginatedResponse[*ChatRoomResponse]
type MessageListResponse = PaginatedResponse[*MessageResponse]
type ChatMemberListResponse = PaginatedResponse[*ChatMemberResponse]

// 转换方法
func ToChatRoomResponse(room *model.ChatRoom) *ChatRoomResponse {
	if room == nil {
		return nil
	}

	resp := &ChatRoomResponse{
		ID:           room.ID,
		Name:         room.Name,
		Type:         room.Type,
		AvatarURL:    room.AvatarURL,
		Announcement: room.Announcement,
		CreatedAt:    room.CreatedAt,
	}

	return resp
}

func ToMessageResponse(msg *model.Message) *MessageResponse {
	if msg == nil {
		return nil
	}

	resp := &MessageResponse{
		ID:          msg.ID,
		ChatRoomID:  msg.ChatRoomID,
		Sender:      *ToUserBrief(&msg.Sender),
		ContentType: msg.ContentType,
		Content:     msg.Content,
		CreatedAt:   msg.CreatedAt,
	}

	return resp
}

func ToChatMemberResponse(member *model.ChatRoomMember) *ChatMemberResponse {
	if member == nil {
		return nil
	}

	return &ChatMemberResponse{
		UserBrief: *ToUserBrief(&member.User),
		Role:      member.Role,
		Nickname:  member.Nickname,
		JoinedAt:  member.CreatedAt,
		IsMuted:   member.IsMuted,
	}
}
