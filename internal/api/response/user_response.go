package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// UserBrief 用户简要信息
type UserBrief struct {
	UID       string `json:"uid"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// UserInfo 用户基础信息
type UserInfo struct {
	UserBrief
	Bio                 string     `json:"bio"`
	Gender              string     `json:"gender"`
	BirthDate           *time.Time `json:"birth_date,omitempty"`
	Location            *Location  `json:"location,omitempty"`
	Language            string     `json:"language"`
	PrivacyLevel        string     `json:"privacy_level"`
	NotificationEnabled bool       `json:"notification_enabled"`
	LocationSharing     bool       `json:"location_shared"`
	PhotoEnabled        bool       `json:"photo_enabled"`
	LastActiveAt        *time.Time `json:"last_active_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UserType            string     `json:"user_type"`
}

// UserProfile 用户完整资料
type UserProfile struct {
	UserInfo
	Stats        UserStats     `json:"stats"`
	Relationship *Relationship `json:"relationship,omitempty"`
}

// UserStats 用户统计信息
type UserStats struct {
	TopicsCount    int64 `json:"topics_count"`
	FollowersCount int64 `json:"followers_count"`
	FollowingCount int64 `json:"following_count"`
	FriendsCount   int64 `json:"friends_count"`
}

// Relationship 关系信息
type Relationship struct {
	IsFollowing bool `json:"is_following"`
	IsFollowed  bool `json:"is_followed"`
	IsFriend    bool `json:"is_friend"`
	IsRejected  bool `json:"is_rejected"`
}

type LoginInfo struct {
	CsrfToken   string `json:"csrf_token"`
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
	PhotoUrl    string `json:"photo_url"`
	Email       string `json:"email"`
}

// 使用泛型定义分页响应类型
type UserListResponse = PaginatedResponse[*UserInfo]
type UserBriefListResponse = PaginatedResponse[*UserBrief]

// Convert Functions

func ToUserBrief(user *model.User) *UserBrief {
	if user == nil {
		return nil
	}

	return &UserBrief{
		UID:       user.UID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}
}

func ToUserInfo(user *model.User) *UserInfo {
	if user == nil {
		return nil
	}

	info := &UserInfo{
		UserBrief: UserBrief{
			UID:       user.UID,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
		},
		Bio:                 user.Bio,
		Gender:              user.Gender,
		Language:            user.Language,
		PrivacyLevel:        user.PrivacyLevel,
		NotificationEnabled: user.NotificationEnabled,
		LocationSharing:     user.LocationSharing,
		PhotoEnabled:        user.PhotoEnabled,
		LastActiveAt:        user.LastActiveAt,
		CreatedAt:           user.CreatedAt,
		UserType:            user.UserType,
	}

	if user.BirthDate != nil {
		info.BirthDate = user.BirthDate
	}

	if user.LocationLatitude != 0 || user.LocationLongitude != 0 {
		info.Location = &Location{
			Latitude:  user.LocationLatitude,
			Longitude: user.LocationLongitude,
		}
	}

	return info
}

func ToUserProfile(user *model.User) *UserProfile {
	if user == nil {
		return nil
	}

	return &UserProfile{
		UserInfo: *ToUserInfo(user),
		Stats: UserStats{
			TopicsCount:    user.TopicsCount,
			FollowersCount: user.FollowersCount,
			FollowingCount: user.FollowingCount,
			FriendsCount:   user.FriendsCount,
		},
	}
}
