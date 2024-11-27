package response

import "time"

// RelationshipResponse 关系响应
type RelationshipResponse struct {
	TargetUser UserBrief  `json:"target_user"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

// RelationshipStatsResponse 关系统计响应
type RelationshipStatsResponse struct {
	FollowersCount    int64 `json:"followers_count"`
	FollowingCount    int64 `json:"following_count"`
	FriendsCount      int64 `json:"friends_count"`
	PendingCount      int64 `json:"pending_count"`
	BlockedUsersCount int64 `json:"blocked_users_count"`
}

// RelationshipStatusResponse 关系状态响应
type RelationshipStatusResponse struct {
	IsFollowing bool `json:"is_following"`
	IsFollowed  bool `json:"is_followed"`
	IsFriend    bool `json:"is_friend"`
	IsBlocked   bool `json:"is_blocked"`
	IsPending   bool `json:"is_pending"`
}

// FriendResponse 好友响应
type FriendResponse struct {
	UserBrief
	CommonFriends     int        `json:"common_friends"`
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty"`
}
