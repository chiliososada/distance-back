package model

import (
	"database/sql"
	"time"
)

// 用户状态
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

// 性别
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"
)

// 隐私级别
const (
	PrivacyPublic  = "public"
	PrivacyFriends = "friends"
	PrivacyPrivate = "private"
)

// User 用户模型
type User struct {
	BaseModel
	Nickname            string     `gorm:"size:50" json:"nickname"`
	AvatarURL           string     `gorm:"size:255" json:"avatar_url"`
	BirthDate           *time.Time `json:"birth_date"`
	Gender              string     `gorm:"type:enum('male','female','other')" json:"gender"`
	Bio                 string     `gorm:"type:text" json:"bio"`
	LocationLatitude    float64    `gorm:"type:decimal(10,8)" json:"location_latitude"`
	LocationLongitude   float64    `gorm:"type:decimal(11,8)" json:"location_longitude"`
	Language            string     `gorm:"size:10;default:'zh_CN'" json:"language"`
	Status              string     `gorm:"type:enum('active','inactive','banned');default:'active'" json:"status"`
	PrivacyLevel        string     `gorm:"type:enum('public','friends','private');default:'public'" json:"privacy_level"`
	NotificationEnabled bool       `gorm:"default:true" json:"notification_enabled"`
	LocationSharing     bool       `gorm:"default:true" json:"location_sharing"`
	PhotoEnabled        bool       `gorm:"default:true" json:"photo_enabled"`
	LastActiveAt        *time.Time `json:"last_active_at"`
	UserType            string     `gorm:"type:enum('individual','merchant','official','admin');default:'individual'" json:"user_type"`
	// 添加统计字段
	TopicsCount    int64 `gorm:"-" json:"topics_count"`    // 话题数
	FollowersCount int64 `gorm:"-" json:"followers_count"` // 粉丝数
	FollowingCount int64 `gorm:"-" json:"following_count"` // 关注数
	FriendsCount   int64 `gorm:"-" json:"friends_count"`   // 好友数
}

// UserAuthentication 用户认证模型
type UserAuthentication struct {
	BaseModel
	UserID       uint64         `gorm:"uniqueIndex:idx_user_id" json:"user_id"`
	FirebaseUID  string         `gorm:"size:128;uniqueIndex" json:"firebase_uid"`
	AuthProvider string         `gorm:"type:enum('password','phone','google','apple','anonymous')" json:"auth_provider"`
	Email        sql.NullString `gorm:"size:100;uniqueIndex" json:"email"`
	PhoneNumber  sql.NullString `gorm:"size:20;uniqueIndex" json:"phone_number"`
	LastSignInAt *time.Time     `json:"last_sign_in_at"`
	User         User           `gorm:"foreignKey:UserID" json:"user"`
}

// UserDevice 用户设备模型
type UserDevice struct {
	BaseModel
	UserID       uint64    `gorm:"index" json:"user_id"`
	DeviceToken  string    `gorm:"size:255;uniqueIndex" json:"device_token"`
	PushProvider string    `gorm:"type:enum('fcm','apns','web')" json:"push_provider"`
	PushEnabled  bool      `gorm:"default:true" json:"push_enabled"`
	DeviceType   string    `gorm:"type:enum('ios','android','web')" json:"device_type"`
	DeviceName   string    `gorm:"size:100" json:"device_name"`
	DeviceModel  string    `gorm:"size:50" json:"device_model"`
	OSVersion    string    `gorm:"size:20" json:"os_version"`
	AppVersion   string    `gorm:"size:20" json:"app_version"`
	BrowserInfo  string    `gorm:"size:200" json:"browser_info"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	LastActiveAt time.Time `json:"last_active_at"`
	BadgeCount   uint      `gorm:"default:0" json:"badge_count"`
	User         User      `gorm:"foreignKey:UserID" json:"user"`
}

// AdminPermission 管理员权限模型
type AdminPermission struct {
	BaseModel
	UserID         uint64 `gorm:"uniqueIndex:uk_admin_permission" json:"user_id"`
	PermissionType string `gorm:"type:enum('super_admin','user_manage','content_audit','system_config','data_analysis');uniqueIndex:uk_admin_permission" json:"permission_type"`
	Status         bool   `gorm:"default:true" json:"status"`
	User           User   `gorm:"foreignKey:UserID" json:"user"`
}

// UserBan 用户封禁记录模型
type UserBan struct {
	BaseModel
	UserID     uint64    `gorm:"index:idx_user_status" json:"user_id"`
	OperatorID uint64    `json:"operator_id"`
	Reason     string    `gorm:"type:text" json:"reason"`
	BanStart   time.Time `json:"ban_start"`
	BanEnd     time.Time `json:"ban_end"`
	Status     string    `gorm:"type:enum('active','expired','cancelled');default:'active';index:idx_user_status" json:"status"`
	User       User      `gorm:"foreignKey:UserID" json:"user"`
	Operator   User      `gorm:"foreignKey:OperatorID" json:"operator"`
}
