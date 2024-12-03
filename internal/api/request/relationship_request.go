package request

// FollowRequest 关注请求
type FollowRequest struct {
	TargetUID string `json:"target_uid" binding:"required,uuid"` // 目标用户UID
}

// GetRelationshipsRequest 获取关系列表请求
type GetRelationshipsRequest struct {
	PaginationQuery
	SortQuery
	Status string `form:"status" binding:"omitempty,oneof=pending accepted blocked"` // 关系状态
}

// BlockUserRequest 拉黑用户请求
type BlockUserRequest struct {
	TargetUID string `json:"target_uid" binding:"required,uuid"` // 目标用户UID
	Reason    string `json:"reason" binding:"omitempty,max:500"` // 拉黑原因
}

// CheckRelationshipRequest 检查关系请求
type CheckRelationshipRequest struct {
	UIDParam // 继承UUID参数结构
}

// BatchCheckRelationshipRequest 批量检查关系请求
type BatchCheckRelationshipRequest struct {
	TargetUIDs []string `json:"target_uids" binding:"required,min=1,max=100,dive,uuid"`
}

// UpdateRelationshipRequest 更新关系请求
type UpdateRelationshipRequest struct {
	Status string `json:"status" binding:"required,oneof=pending accepted blocked"`
}

// GetMutualFriendsRequest 获取共同好友请求
type GetMutualFriendsRequest struct {
	PaginationQuery
	TargetUID string `form:"target_uid" binding:"required,uuid"`
}

// GetRelationshipStatsRequest 获取关系统计请求
type GetRelationshipStatsRequest struct {
	UserUID string `form:"user_uid" binding:"required,uuid"`
}
