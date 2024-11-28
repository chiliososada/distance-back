package model

import "time"

const (
	// Topic 状态
	TopicStatusActive    = "active"
	TopicStatusClosed    = "closed"
	TopicStatusCancelled = "cancelled"

	// 互动类型
	InteractionTypeLike     = "like"
	InteractionTypeFavorite = "favorite"
	InteractionTypeShare    = "share"

	// 互动状态
	InteractionStatusActive    = "active"
	InteractionStatusCancelled = "cancelled"
)

// Topic 话题模型
type Topic struct {
	BaseModel
	UserID            uint64       `gorm:"index:idx_user_time" json:"user_id"`
	Title             string       `gorm:"size:255" json:"title"`
	Content           string       `gorm:"type:text" json:"content"`
	LocationLatitude  float64      `gorm:"type:decimal(10,8)" json:"location_latitude"`
	LocationLongitude float64      `gorm:"type:decimal(11,8)" json:"location_longitude"`
	LikesCount        uint         `gorm:"default:0" json:"likes_count"`        // 点赞数
	ParticipantsCount uint         `gorm:"default:0" json:"participants_count"` // 参与人数
	ViewsCount        uint         `gorm:"default:0" json:"views_count"`        // 浏览数
	SharesCount       uint         `gorm:"default:0" json:"shares_count"`       // 分享数
	ExpiresAt         time.Time    `json:"expires_at"`                          // 过期时间
	Status            string       `gorm:"type:enum('active','closed','cancelled');default:'active'" json:"status"`
	User              User         `gorm:"foreignKey:UserID" json:"user"`
	TopicImages       []TopicImage `gorm:"foreignKey:TopicID" json:"topic_images"`
	TopicTags         []TopicTag   `gorm:"foreignKey:TopicID" json:"topic_tags"`
}

// TopicImage 话题图片模型
type TopicImage struct {
	BaseModel
	TopicID     uint64 `gorm:"index:idx_topic_sort" json:"topic_id"`
	ImageURL    string `gorm:"size:255" json:"image_url"`
	SortOrder   uint   `gorm:"index:idx_topic_sort" json:"sort_order"`
	ImageWidth  uint   `json:"image_width"`
	ImageHeight uint   `json:"image_height"`
	FileSize    uint   `json:"file_size"`
	Topic       Topic  `gorm:"foreignKey:TopicID" json:"topic"`
}

// Tag 标签模型
type Tag struct {
	BaseModel
	Name     string `gorm:"size:50;uniqueIndex" json:"name"`
	UseCount uint   `gorm:"default:0" json:"use_count"` // 使用次数
}

// TopicTag 话题标签关联模型
type TopicTag struct {
	TopicID   uint64    `gorm:"primaryKey" json:"topic_id"`
	TagID     uint64    `gorm:"primaryKey" json:"tag_id"`
	Topic     Topic     `gorm:"foreignKey:TopicID" json:"topic"`
	Tag       Tag       `gorm:"foreignKey:TagID" json:"tag"`
	CreatedAt time.Time `json:"created_at"`
}

// TopicInteraction 话题互动模型
type TopicInteraction struct {
	BaseModel
	TopicID           uint64 `gorm:"uniqueIndex:unique_interaction" json:"topic_id"`
	UserID            uint64 `gorm:"uniqueIndex:unique_interaction" json:"user_id"`
	InteractionType   string `gorm:"type:enum('like','favorite','share');uniqueIndex:unique_interaction" json:"interaction_type"`
	InteractionStatus string `gorm:"type:enum('active','cancelled');default:'active'" json:"interaction_status"`
	Topic             Topic  `gorm:"foreignKey:TopicID" json:"topic"`
	User              User   `gorm:"foreignKey:UserID" json:"user"`
}

// TopicReport 话题举报模型
type TopicReport struct {
	BaseModel
	TopicID      uint64 `json:"topic_id"`
	ReporterID   uint64 `json:"reporter_id"`
	ReasonType   string `gorm:"type:enum('spam','abuse','copyright','other')" json:"reason_type"`
	ReasonDetail string `gorm:"type:text" json:"reason_detail"`
	Status       string `gorm:"type:enum('pending','processing','resolved','rejected');default:'pending'" json:"status"`
	HandlerID    uint64 `json:"handler_id"`                     // 处理人ID
	HandleResult string `gorm:"type:text" json:"handle_result"` // 处理结果
	Topic        Topic  `gorm:"foreignKey:TopicID" json:"topic"`
	Reporter     User   `gorm:"foreignKey:ReporterID" json:"reporter"`
	Handler      User   `gorm:"foreignKey:HandlerID" json:"handler"`
}
