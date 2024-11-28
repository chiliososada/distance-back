package request

// FollowRequest 关注请求
type FollowRequest struct {
	TargetID uint64 `json:"target_id" binding:"required,min=1"` // 目标用户ID
}

// GetRelationshipsRequest 获取关系列表请求
type GetRelationshipsRequest struct {
	PaginationQuery
	SortQuery
	Status string `form:"status" binding:"omitempty,oneof=pending accepted blocked"` // 关系状态
}

// BlockUserRequest 拉黑用户请求
type BlockUserRequest struct {
	TargetID uint64 `json:"target_id" binding:"required,min=1"` // 目标用户ID
	Reason   string `json:"reason" binding:"omitempty,max=500"` // 拉黑原因
}
