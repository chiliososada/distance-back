package model

import "time"

// ChatRoom 聊天室模型
type ChatRoom struct {
	BaseModel
	Name         string  `gorm:"size:100" json:"name"`
	Type         string  `gorm:"type:enum('individual','group','merchant','official')" json:"type"`
	TopicID      *uint64 `json:"topic_id"`
	AvatarURL    string  `gorm:"size:255" json:"avatar_url"`
	Announcement string  `gorm:"type:text" json:"announcement"`
	Topic        *Topic  `gorm:"foreignKey:TopicID" json:"topic"`
}

// ChatRoomMember 聊天室成员模型
type ChatRoomMember struct {
	BaseModel
	ChatRoomID        uint64   `gorm:"uniqueIndex:unique_member" json:"chat_room_id"`
	UserID            uint64   `gorm:"uniqueIndex:unique_member" json:"user_id"`
	Role              string   `gorm:"type:enum('owner','admin','member');default:'member'" json:"role"`
	Nickname          string   `gorm:"size:50" json:"nickname"`
	LastReadMessageID uint64   `gorm:"default:0" json:"last_read_message_id"`
	IsMuted           bool     `gorm:"default:false" json:"is_muted"`
	ChatRoom          ChatRoom `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
	User              User     `gorm:"foreignKey:UserID" json:"user"`
}

// Message 消息模型
type Message struct {
	BaseModel
	ChatRoomID  uint64   `gorm:"index:idx_chat_room_time" json:"chat_room_id"`
	SenderID    uint64   `json:"sender_id"`
	ContentType string   `gorm:"type:enum('text','image','file','system');default:'text'" json:"content_type"`
	Content     string   `gorm:"type:text" json:"content"`
	ChatRoom    ChatRoom `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
	Sender      User     `gorm:"foreignKey:SenderID" json:"sender"`
}

// MessageMedia 消息媒体模型
type MessageMedia struct {
	BaseModel
	MessageID uint64  `json:"message_id"`
	MediaType string  `gorm:"type:enum('image','file')" json:"media_type"`
	MediaURL  string  `gorm:"size:255" json:"media_url"`
	FileName  string  `gorm:"size:255" json:"file_name"`
	FileSize  uint    `json:"file_size"`
	Message   Message `gorm:"foreignKey:MessageID" json:"message"`
}

// PinnedChatRoom 聊天室置顶模型
type PinnedChatRoom struct {
	UserID     uint64    `gorm:"primaryKey" json:"user_id"`
	ChatRoomID uint64    `gorm:"primaryKey" json:"chat_room_id"`
	PinnedAt   time.Time `json:"pinned_at"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	ChatRoom   ChatRoom  `gorm:"foreignKey:ChatRoomID" json:"chat_room"`
}
