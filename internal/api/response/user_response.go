package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// UserBrief 用户简要信息
type UserBrief struct {
	ID        uint64 `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// UserInfo 用户基础信息
type UserInfo struct {
	UserBrief
	Bio              string     `json:"bio"`
	Gender           string     `json:"gender"`
	BirthDate        *time.Time `json:"birth_date,omitempty"`
	Location         *Location  `json:"location,omitempty"`
	PrivacyLevel     string     `json:"privacy_level"`
	IsLocationShared bool       `json:"is_location_shared"`
	IsPhotoEnabled   bool       `json:"is_photo_enabled"`
	LastActiveAt     *time.Time `json:"last_active_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
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
	IsBlocked   bool `json:"is_blocked"`
}

// Location 位置信息
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance,omitempty"`
}

// 使用泛型定义分页响应类型
type UserListResponse = PaginatedResponse[*UserInfo]
type UserBriefListResponse = PaginatedResponse[*UserBrief]

// ToUserInfo 转换为用户信息
func ToUserInfo(user *model.User) *UserInfo {
	if user == nil {
		return nil
	}

	info := &UserInfo{
		UserBrief: UserBrief{
			ID:        user.ID,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
		},
		Bio:              user.Bio,
		Gender:           user.Gender,
		PrivacyLevel:     user.PrivacyLevel,
		IsLocationShared: user.LocationSharing,
		IsPhotoEnabled:   user.PhotoEnabled,
		LastActiveAt:     user.LastActiveAt,
		CreatedAt:        user.CreatedAt,
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

// ToUserProfile 转换为用户完整资料
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

// ToUserBrief 转换为用户简要信息
func ToUserBrief(user *model.User) *UserBrief {
	if user == nil {
		return nil
	}

	return &UserBrief{
		ID:        user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}
}
