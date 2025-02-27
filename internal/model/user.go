package model

import (
	"database/sql"
	"time"

	"github.com/chiliososada/distance-back/internal/api/request"
)

// 用户状态常量
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

// 性别常量
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOther  = "other"
)

// 隐私级别常量
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
	Email               string     `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Gender              string     `gorm:"type:enum('male','female','others','unknown');default:unknown" json:"gender"`
	Bio                 string     `gorm:"type:text" json:"bio"`
	LocationLatitude    *float64   `gorm:"type:decimal(10,8)" json:"location_latitude,omitempty"`
	LocationLongitude   *float64   `gorm:"type:decimal(11,8)" json:"location_longitude,omitempty"`
	Language            string     `gorm:"size:10;default:'zh_CN'" json:"language"`
	Status              string     `gorm:"type:enum('active','inactive','banned');default:'active'" json:"status"`
	PrivacyLevel        string     `gorm:"type:enum('public','friends','private');default:'public'" json:"privacy_level"`
	NotificationEnabled bool       `gorm:"default:true" json:"notification_enabled"`
	LocationSharing     bool       `gorm:"default:true" json:"location_sharing"`
	PhotoEnabled        bool       `gorm:"default:true" json:"photo_enabled"`
	LastActiveAt        *time.Time `json:"last_active_at,omitempty"`
	UserType            string     `gorm:"type:enum('individual','merchant','official','admin');default:'individual'" json:"user_type"`
	// 统计字段
	TopicsCount    int64 `gorm:"-" json:"topics_count"`
	FollowersCount int64 `gorm:"-" json:"followers_count"`
	FollowingCount int64 `gorm:"-" json:"following_count"`
	FriendsCount   int64 `gorm:"-" json:"friends_count"`

	Chat []UserChat `gorm:"foreignKey:UserUID;references:UID" json:"chat"`
}

func (u *User) UpdateFromRequest(req *request.UpdateProfileRequest) {
	if req.Email != nil {
		u.Email = *req.Email
	}
	if req.Nickname != nil {
		u.Nickname = *req.Nickname
	}
	if req.Bio != nil {
		u.Bio = *req.Bio
	}
	if req.Gender != nil {
		u.Gender = *req.Gender
	}
	if req.BirthDate != nil {
		u.BirthDate = *&req.BirthDate
	}
	if req.Language != nil {
		u.Language = *req.Language
	}
	if req.PrivacyLevel != nil {
		u.PrivacyLevel = *req.PrivacyLevel
	}
	if req.LocationSharing != nil {
		u.LocationSharing = *req.LocationSharing
	}
	if req.PhotoEnabled != nil {
		u.PhotoEnabled = *req.PhotoEnabled
	}
	if req.NotificationEnabled != nil {
		u.NotificationEnabled = *req.NotificationEnabled
	}
	if req.AvatarURL != nil {
		u.AvatarURL = *req.AvatarURL
	}
}

// UserAuthentication 用户认证模型
type UserAuthentication struct {
	BaseModel
	UserUID      string         `gorm:"type:varchar(36);uniqueIndex:idx_user_uid" json:"user_uid"`
	FirebaseUID  string         `gorm:"size:128;uniqueIndex" json:"firebase_uid"`
	AuthProvider string         `gorm:"type:enum('password','phone','google','apple','anonymous')" json:"auth_provider"`
	Email        sql.NullString `gorm:"size:100;uniqueIndex" json:"email"`
	PhoneNumber  sql.NullString `gorm:"size:20;uniqueIndex" json:"phone_number"`
	LastSignInAt *time.Time     `json:"last_sign_in_at"`
	User         User           `gorm:"foreignKey:UserUID;references:UID" json:"user"`
}

// UserDevice 用户设备模型
type UserDevice struct {
	BaseModel
	UserUID      string    `gorm:"type:varchar(36);index" json:"user_uid"`
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
	User         User      `gorm:"foreignKey:UserUID;references:UID" json:"user"`
}

// AdminPermission 管理员权限模型
type AdminPermission struct {
	BaseModel
	UserUID        string `gorm:"type:varchar(36);uniqueIndex:uk_admin_permission" json:"user_uid"`
	PermissionType string `gorm:"type:enum('super_admin','user_manage','content_audit','system_config','data_analysis');uniqueIndex:uk_admin_permission" json:"permission_type"`
	Status         bool   `gorm:"default:true" json:"status"`
	User           User   `gorm:"foreignKey:UserUID;references:UID" json:"user"`
}

// UserBan 用户封禁记录模型
type UserBan struct {
	BaseModel
	UserUID     string    `gorm:"type:varchar(36);index:idx_user_status" json:"user_uid"`
	OperatorUID string    `gorm:"type:varchar(36)" json:"operator_uid"`
	Reason      string    `gorm:"type:text" json:"reason"`
	BanStart    time.Time `json:"ban_start"`
	BanEnd      time.Time `json:"ban_end"`
	Status      string    `gorm:"type:enum('active','expired','cancelled');default:'active';index:idx_user_status" json:"status"`
	User        User      `gorm:"foreignKey:UserUID;references:UID" json:"user"`
	Operator    User      `gorm:"foreignKey:OperatorUID;references:UID" json:"operator"`
}

type UserChat struct {
	BaseModel
	UserUID     string   `gorm:"type:varchar(36);uniqueIndex" json:"user_uid"`
	ChatRoomUID string   `gorm:"type:varchar(36);index" json:"chat_room_uid"`
	User        User     `gorm:"foreignKey:UserUID;references:UID" json:"user"`
	ChatRoom    ChatRoom `gorm:"foreignKey:ChatRoomUID;references:UID" json:"chat_room"`
}
