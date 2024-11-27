package constants

import "time"

const (
	// 环境相关
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTesting     = "testing"

	// 用户相关
	DefaultAvatarURL = "https://example.com/default-avatar.png"
	MaxNicknameLen   = 50
	MaxBioLen        = 500

	// 时间相关
	TimeFormat     = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
	TimeFormatNano = "2006-01-02 15:04:05.999999999"

	// 文件相关
	MaxFileSize       = 10 * 1024 * 1024 // 10MB
	MaxImageSize      = 5 * 1024 * 1024  // 5MB
	MaxImageWidth     = 4096             // 最大图片宽度
	MaxImageHeight    = 4096             // 最大图片高度
	AllowedImageTypes = ".jpg,.jpeg,.png,.gif,.webp"

	// 分页相关
	DefaultPageSize = 10
	MaxPageSize     = 100
	DefaultPage     = 1

	// 标签相关
	MaxTagLength    = 50
	MaxTagsPerTopic = 10
	MinTagLength    = 2

	// 距离相关（米）
	NearbyDistance    = 5000    // 5公里
	MaxSearchDistance = 50000   // 50公里
	LocationPrecision = 0.00001 // 经纬度精度

	// 缓存相关
	DefaultCacheExpiration = 3600 // 1小时
	UserCacheKeyPrefix     = "user:"
	TopicCacheKeyPrefix    = "topic:"

	// 消息相关
	MaxMessageLength   = 1000
	MaxChatRoomMembers = 500
	// 缓存时间
	TagCacheExpiration      = 24 * time.Hour
	UserCacheExpiration     = 30 * time.Minute
	TopicCacheExpiration    = time.Hour
	LocationCacheExpiration = 5 * time.Minute

	// 其他限制
	MaxSearchKeywordLength = 50

	// 状态相关
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusDeleted  = "deleted"
)

// 用户类型
type UserType string

const (
	UserTypeIndividual UserType = "individual"
	UserTypeMerchant   UserType = "merchant"
	UserTypeOfficial   UserType = "official"
	UserTypeAdmin      UserType = "admin"
)

// 通知类型
type NotificationType string

const (
	NotificationTypeSystem   NotificationType = "system"
	NotificationTypeTopic    NotificationType = "topic"
	NotificationTypeChat     NotificationType = "chat"
	NotificationTypeMerchant NotificationType = "merchant"
)
