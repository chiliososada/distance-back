package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// RelationshipResponse 关系响应
type RelationshipResponse struct {
	TargetUser UserBrief  `json:"target_user"`           // 目标用户
	Status     string     `json:"status"`                // 关系状态(pending/accepted/blocked)
	CreatedAt  time.Time  `json:"created_at"`            // 创建时间
	AcceptedAt *time.Time `json:"accepted_at,omitempty"` // 接受时间
}

// RelationshipStatsResponse 关系统计响应
type RelationshipStatsResponse struct {
	FollowersCount    int64 `json:"followers_count"`     // 粉丝数
	FollowingCount    int64 `json:"following_count"`     // 关注数
	FriendsCount      int64 `json:"friends_count"`       // 好友数
	PendingCount      int64 `json:"pending_count"`       // 待处理数
	BlockedUsersCount int64 `json:"blocked_users_count"` // 拉黑数
}

// RelationshipStatusResponse 关系状态响应
type RelationshipStatusResponse struct {
	IsFollowing bool `json:"is_following"` // 是否关注
	IsFollowed  bool `json:"is_followed"`  // 是否被关注
	IsFriend    bool `json:"is_friend"`    // 是否好友
	IsBlocked   bool `json:"is_blocked"`   // 是否拉黑
	// IsPending   bool `json:"is_pending"`   // 是否待处理
}

// 使用泛型定义分页响应类型
type RelationshipListResponse = PaginatedResponse[*RelationshipResponse]

// ToRelationshipResponse 转换为关系响应
func ToRelationshipResponse(relationship *model.UserRelationship) *RelationshipResponse {
	if relationship == nil {
		return nil
	}

	return &RelationshipResponse{
		TargetUser: UserBrief{
			ID:        relationship.FollowingID,
			Nickname:  relationship.Following.Nickname,
			AvatarURL: relationship.Following.AvatarURL,
		},
		Status:     relationship.Status,
		CreatedAt:  relationship.CreatedAt,
		AcceptedAt: relationship.AcceptedAt,
	}
}

// FriendResponse 好友响应
type FriendResponse struct {
	UserBrief
	CommonFriends     int        `json:"common_friends"`
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty"`
}
