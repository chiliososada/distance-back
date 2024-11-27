package request

// FollowRequest 关注请求
type FollowRequest struct {
	TargetID uint64 `json:"target_id" binding:"required"`
}

// GetRelationshipsRequest 获取关系列表请求
type GetRelationshipsRequest struct {
	Pagination
	Status string `form:"status" binding:"omitempty,oneof=pending accepted"`
}

// BlockUserRequest 拉黑用户请求
type BlockUserRequest struct {
	TargetID uint64 `json:"target_id" binding:"required"`
	Reason   string `json:"reason" binding:"max=500"`
}
