package model

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/copier"
)

const (
	// Topic 状态常量
	TopicStatusActive    = "active"
	TopicStatusClosed    = "closed"
	TopicStatusCancelled = "cancelled"

	// 互动类型常量
	InteractionTypeLike     = "like"
	InteractionTypeFavorite = "favorite"
	InteractionTypeShare    = "share"

	// 互动状态常量
	InteractionStatusActive    = "active"
	InteractionStatusCancelled = "cancelled"
)

// Topic 话题模型
type Topic struct {
	BaseModel
	UserUID           string        `gorm:"type:varchar(36);index:idx_user_time" json:"user_uid"`
	Title             string        `gorm:"size:255" json:"title"`
	Content           string        `gorm:"type:text" json:"content"`
	LocationLatitude  *float64      `gorm:"type:decimal(10,8)" json:"location_latitude,omitempty"`
	LocationLongitude *float64      `gorm:"type:decimal(11,8)" json:"location_longitude,omitempty"`
	LikesCount        uint          `gorm:"default:0" json:"likes_count"`
	ParticipantsCount uint          `gorm:"default:0" json:"participants_count"`
	ViewsCount        uint          `gorm:"default:0" json:"views_count"`
	SharesCount       uint          `gorm:"default:0" json:"shares_count"`
	ExpiresAt         time.Time     `gorm:"type:timestamp" json:"expires_at"`
	Status            string        `gorm:"type:enum('active','closed','cancelled');default:'active'" json:"status"`
	User              User          `gorm:"foreignKey:UserUID;references:UID" json:"user"`
	TopicImages       []*TopicImage `gorm:"foreignKey:TopicUID;references:UID" json:"images"`
	Tags              []*Tag        `gorm:"many2many:topic_tags;foreignKey:UID;joinForeignKey:TopicUID;references:UID;joinReferences:TagUID" json:"tags"`
	ChatRoom          *ChatRoom     `gorm:"foreignKey:TopicUID;references:UID" json:"chat_room"`
}

// TopicImage 话题图片模型
type TopicImage struct {
	BaseModel
	TopicUID    string `gorm:"type:varchar(36);index:idx_topic_sort" json:"topic_uid"`
	ImageURL    string `gorm:"size:255" json:"image_url"`
	SortOrder   uint   `gorm:"index:idx_topic_sort" json:"-"`
	ImageWidth  uint   `json:"-"`
	ImageHeight uint   `json:"-"`
	FileSize    uint   `json:"-"`
	Topic       Topic  `gorm:"foreignKey:TopicUID;references:UID" json:"-"`
}

// Tag 标签模型
type Tag struct {
	BaseModel
	Name     string   `gorm:"size:50;uniqueIndex" json:"name"`
	UseCount uint     `gorm:"default:0" json:"-"`
	Topics   []*Topic `gorm:"many2many:topic_tags;foreignKey:UID;joinForeignKey:TagUID;references:UID;joinReferences:TopicUID" json:"-"`
}

// TopicTag 话题标签关联模型 join table

type TopicTag struct {
	TopicUID  string    `gorm:"type:varchar(36);primaryKey" json:"topic_uid"`
	TagUID    string    `gorm:"type:varchar(36);primaryKey" json:"tag_uid"`
	Topic     Topic     `gorm:"foreignKey:TopicUID;references:UID" json:"topic"`
	Tag       Tag       `gorm:"foreignKey:TagUID;references:UID" json:"tag"`
	CreatedAt time.Time `json:"created_at"`
}

// TopicInteraction 话题互动模型
type TopicInteraction struct {
	BaseModel
	TopicUID          string `gorm:"type:varchar(36);uniqueIndex:unique_interaction" json:"topic_uid"`
	UserUID           string `gorm:"type:varchar(36);uniqueIndex:unique_interaction" json:"user_uid"`
	InteractionType   string `gorm:"type:enum('like','favorite','share');uniqueIndex:unique_interaction" json:"interaction_type"`
	InteractionStatus string `gorm:"type:enum('active','cancelled');default:'active'" json:"interaction_status"`
	Topic             Topic  `gorm:"foreignKey:TopicUID;references:UID" json:"topic"`
	User              User   `gorm:"foreignKey:UserUID;references:UID" json:"user"`
}

// Store in redis
type CachedTopic struct {
	BaseModel
	UserUID           string    `json:"user_uid"`
	Title             string    `json:"title"`
	Content           string    `json:"content"`
	LocationLatitude  *float64  `json:"location_latitude,omitempty"`
	LocationLongitude *float64  `json:"location_longitude,omitempty"`
	LikesCount        uint      `json:"likes_count"`
	ParticipantsCount uint      `json:"participants_count"`
	ViewsCount        uint      `json:"views_count"`
	SharesCount       uint      `json:"shares_count"`
	ExpiresAt         time.Time `json:"expires_at"`
	Status            string    `json:"status"`
	User              struct {
		Nickname          string     `json:"nickname"`
		AvatarURL         string     `json:"avatar_url"`
		BirthDate         *time.Time `json:"birth_date,omitempty"`
		Gender            string     `json:"gender"`
		LocationLatitude  *float64   `json:"location_latitude,omitempty"`
		LocationLongitude *float64   `json:"location_longitude,omitempty"`
	} `json:"user"`
	TopicImages []string `json:"topic_images" copier:"-"`
	Tags        []string `json:"tags" copier:"-"`
	ChatID      string   `json:"chat_id" copier:"-"`
}

func (t *Topic) CastToCached() CachedTopic {

	var images []string
	var tags []string
	for _, i := range t.TopicImages {
		images = append(images, i.ImageURL)
	}
	for _, t := range t.Tags {
		tags = append(tags, t.Name)
	}
	var cached CachedTopic

	copier.Copy(&cached, t)
	cached.TopicImages = images
	cached.Tags = tags
	if t.ChatRoom != nil {
		cached.ChatID = t.ChatRoom.UID
	}
	return cached
}

func (t *Topic) MarshalToCached() ([]byte, error) {
	cached := t.CastToCached()
	return json.Marshal(&cached)

}
