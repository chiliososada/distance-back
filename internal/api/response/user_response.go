package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// UserResponse 用户信息响应
type UserResponse struct {
	ID               uint64     `json:"id"`
	Nickname         string     `json:"nickname"`
	AvatarURL        string     `json:"avatar_url"`
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

// UserProfileResponse 用户详细资料响应
type UserProfileResponse struct {
	UserResponse
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
	Distance  float64 `json:"distance,omitempty"` // 米
}

// ToResponse 将用户模型转换为响应
func ToResponse(user *model.User) *UserResponse {
	if user == nil {
		return nil
	}

	resp := &UserResponse{
		ID:               user.ID,
		Nickname:         user.Nickname,
		AvatarURL:        user.AvatarURL,
		Bio:              user.Bio,
		Gender:           user.Gender,
		PrivacyLevel:     user.PrivacyLevel,
		IsLocationShared: user.LocationSharing,
		IsPhotoEnabled:   user.PhotoEnabled,
		LastActiveAt:     user.LastActiveAt,
		CreatedAt:        user.CreatedAt,
	}

	if user.BirthDate != nil {
		resp.BirthDate = user.BirthDate
	}

	if user.LocationLatitude != 0 || user.LocationLongitude != 0 {
		resp.Location = &Location{
			Latitude:  user.LocationLatitude,
			Longitude: user.LocationLongitude,
		}
	}

	return resp
}
