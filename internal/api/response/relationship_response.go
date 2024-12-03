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
}

// FriendResponse 好友响应
type FriendResponse struct {
	UserBrief
	CommonFriends     int        `json:"common_friends"`                // 共同好友数
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty"` // 最后互动时间
}

// 使用泛型定义分页响应类型
type RelationshipListResponse = PaginatedResponse[*RelationshipResponse]
type FriendListResponse = PaginatedResponse[*FriendResponse]

// Convert Functions

func ToRelationshipResponse(relationship *model.UserRelationship) *RelationshipResponse {
	if relationship == nil {
		return nil
	}

	return &RelationshipResponse{
		TargetUser: UserBrief{
			UID:       relationship.FollowingUID,
			Nickname:  relationship.Following.Nickname,
			AvatarURL: relationship.Following.AvatarURL,
		},
		Status:     relationship.Status,
		CreatedAt:  relationship.CreatedAt,
		AcceptedAt: relationship.AcceptedAt,
	}
}

func ToFriendResponse(user *model.User, commonFriends int, lastInteractionAt *time.Time) *FriendResponse {
	if user == nil {
		return nil
	}

	return &FriendResponse{
		UserBrief: UserBrief{
			UID:       user.UID,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
		},
		CommonFriends:     commonFriends,
		LastInteractionAt: lastInteractionAt,
	}
}

// 分页响应转换函数
func ToRelationshipListResponse(relationships []*model.UserRelationship, total int64, page, size int) *RelationshipListResponse {
	list := make([]*RelationshipResponse, len(relationships))
	for i, rel := range relationships {
		list[i] = ToRelationshipResponse(rel)
	}
	return NewPaginatedResponse(list, total, page, size)
}

func ToFriendListResponse(users []*model.User, commonFriends map[string]int, lastInteractions map[string]*time.Time, total int64, page, size int) *FriendListResponse {
	list := make([]*FriendResponse, len(users))
	for i, user := range users {
		list[i] = ToFriendResponse(
			user,
			commonFriends[user.UID],
			lastInteractions[user.UID],
		)
	}
	return NewPaginatedResponse(list, total, page, size)
}
