package model

import (
	"time"
)

// ChatRoom 聊天室模型
type ChatRoom struct {
	BaseModel
	Name         string `gorm:"size:100" json:"name"`
	Type         string `gorm:"type:enum('individual','group','merchant','official')" json:"type"`
	TopicUID     string `gorm:"type:varchar(36)" json:"topic_uid"`
	AvatarURL    string `gorm:"size:255" json:"avatar_url"`
	Announcement string `gorm:"type:text" json:"announcement"`
	// Topic        *Topic `gorm:"foreignKey:TopicUID;references:UID" json:"topic"`
	// 添加与 ChatRoomMember 的关联
	ChatRoomMembers []*ChatRoomMember `gorm:"foreignKey:ChatRoomUID;references:UID" json:"chat_room_members"`
}

// PrivateChatView 私聊视图模型
type PrivateChatView struct {
	RoomUID  string `gorm:"column:room_uid"`
	RoomName string `gorm:"column:room_name"`
	RoomType string `gorm:"column:room_type"`
	UserUID1 string `gorm:"column:user_uid1"`
	UserUID2 string `gorm:"column:user_uid2"`
}

// ChatRoomMember 聊天室成员模型
type ChatRoomMember struct {
	BaseModel
	ChatRoomUID       string   `gorm:"type:varchar(36);uniqueIndex:unique_member" json:"chat_room_uid"`
	UserUID           string   `gorm:"type:varchar(36);uniqueIndex:unique_member" json:"user_uid"`
	Role              string   `gorm:"type:enum('owner','admin','member');default:'member'" json:"role"`
	Nickname          string   `gorm:"size:50" json:"nickname"`
	LastReadMessageID string   `gorm:"type:varchar(36);default:'00000000-0000-0000-0000-000000000000'" json:"last_read_message_id"`
	IsMuted           bool     `gorm:"default:false" json:"is_muted"`
	ChatRoom          ChatRoom `gorm:"foreignKey:ChatRoomUID;references:UID" json:"chat_room"`
	User              User     `gorm:"foreignKey:UserUID;references:UID" json:"user"`
}

// Message 消息模型
type Message struct {
	BaseModel
	ChatRoomUID string   `gorm:"type:varchar(36);index:idx_chat_room_time" json:"chat_room_uid"`
	SenderUID   string   `gorm:"type:varchar(36)" json:"sender_uid"`
	ContentType string   `gorm:"type:enum('text','image','file','system');default:'text'" json:"content_type"`
	Content     string   `gorm:"type:text" json:"content"`
	ChatRoom    ChatRoom `gorm:"foreignKey:ChatRoomUID;references:UID" json:"chat_room"`
	Sender      User     `gorm:"foreignKey:SenderUID;references:UID" json:"sender"`
}

// MessageMedia 消息媒体模型
type MessageMedia struct {
	BaseModel
	MessageUID string  `gorm:"type:varchar(36)" json:"message_uid"`
	MediaType  string  `gorm:"type:enum('image','file')" json:"media_type"`
	MediaURL   string  `gorm:"size:255" json:"media_url"`
	FileName   string  `gorm:"size:255" json:"file_name"`
	FileSize   uint    `json:"file_size"`
	Message    Message `gorm:"foreignKey:MessageUID;references:UID" json:"message"`
}

// PinnedChatRoom 聊天室置顶模型
type PinnedChatRoom struct {
	UserUID     string    `gorm:"type:varchar(36);primaryKey" json:"user_uid"`
	ChatRoomUID string    `gorm:"type:varchar(36);primaryKey" json:"chat_room_uid"`
	PinnedAt    time.Time `json:"pinned_at"`
	User        User      `gorm:"foreignKey:UserUID;references:UID" json:"user"`
	ChatRoom    ChatRoom  `gorm:"foreignKey:ChatRoomUID;references:UID" json:"chat_room"`
}
