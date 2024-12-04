package model

import "time"

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
	UserUID           string       `gorm:"type:varchar(36);index:idx_user_time" json:"user_uid"`
	Title             string       `gorm:"size:255" json:"title"`
	Content           string       `gorm:"type:text" json:"content"`
	LocationLatitude  float64      `gorm:"type:decimal(10,8)" json:"location_latitude"`
	LocationLongitude float64      `gorm:"type:decimal(11,8)" json:"location_longitude"`
	LikesCount        uint         `gorm:"default:0" json:"likes_count"`
	ParticipantsCount uint         `gorm:"default:0" json:"participants_count"`
	ViewsCount        uint         `gorm:"default:0" json:"views_count"`
	SharesCount       uint         `gorm:"default:0" json:"shares_count"`
	ExpiresAt         time.Time    `json:"expires_at"`
	Status            string       `gorm:"type:enum('active','closed','cancelled');default:'active'" json:"status"`
	User              User         `gorm:"foreignKey:UserUID;references:UID" json:"user"`
	TopicImages       []TopicImage `gorm:"foreignKey:TopicUID;references:UID" json:"topic_images"`
	TopicTags         []TopicTag   `gorm:"foreignKey:TopicUID;references:UID" json:"topic_tags"`
}

// TopicImage 话题图片模型
type TopicImage struct {
	BaseModel
	TopicUID    string `gorm:"type:varchar(36);index:idx_topic_sort" json:"topic_uid"`
	ImageURL    string `gorm:"size:255" json:"image_url"`
	SortOrder   uint   `gorm:"index:idx_topic_sort" json:"sort_order"`
	ImageWidth  uint   `json:"image_width"`
	ImageHeight uint   `json:"image_height"`
	FileSize    uint   `json:"file_size"`
	Topic       Topic  `gorm:"foreignKey:TopicUID;references:UID" json:"topic"`
}

// Tag 标签模型
type Tag struct {
	BaseModel
	Name     string `gorm:"size:50;uniqueIndex" json:"name"`
	UseCount uint   `gorm:"default:0" json:"use_count"`
}

// TopicTag 话题标签关联模型
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
