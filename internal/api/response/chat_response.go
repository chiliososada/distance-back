package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// ChatRoomResponse 聊天室响应
type ChatRoomResponse struct {
	UID          string        `json:"uid"`
	Name         string        `json:"name"`
	Type         string        `json:"type"` // individual/group/merchant/official
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
	UID         string         `json:"uid"`
	ChatRoomUID string         `json:"chat_room_uid"`
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
	UID  string `json:"uid"`
	Type string `json:"type"` // image/file
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
	Size uint   `json:"size"`
}

// 使用泛型定义分页响应类型
type ChatRoomListResponse = PaginatedResponse[*ChatRoomResponse]
type MessageListResponse = PaginatedResponse[*MessageResponse]
type ChatMemberListResponse = PaginatedResponse[*ChatMemberResponse]

// Convert Functions

func ToChatRoomResponse(room *model.ChatRoom) *ChatRoomResponse {
	if room == nil {
		return nil
	}

	resp := &ChatRoomResponse{
		UID:          room.UID,
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
		UID:         msg.UID,
		ChatRoomUID: msg.ChatRoomUID,
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

// 分页响应转换函数
func ToChatRoomListResponse(rooms []*model.ChatRoom, total int64, page, size int) *ChatRoomListResponse {
	list := make([]*ChatRoomResponse, len(rooms))
	for i, room := range rooms {
		list[i] = ToChatRoomResponse(room)
	}
	return NewPaginatedResponse(list, total, page, size)
}

func ToMessageListResponse(messages []*model.Message, total int64, page, size int) *MessageListResponse {
	list := make([]*MessageResponse, len(messages))
	for i, msg := range messages {
		list[i] = ToMessageResponse(msg)
	}
	return NewPaginatedResponse(list, total, page, size)
}

func ToChatMemberListResponse(members []*model.ChatRoomMember, total int64, page, size int) *ChatMemberListResponse {
	list := make([]*ChatMemberResponse, len(members))
	for i, member := range members {
		list[i] = ToChatMemberResponse(member)
	}
	return NewPaginatedResponse(list, total, page, size)
}
