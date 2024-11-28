// internal/api/request/topic_request.go
package request

import "time"

// CreateTopicRequest 创建话题请求
type CreateTopicRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Content string `json:"content" binding:"required,min=1"`
	Location
	ExpiresAt time.Time `json:"expires_at" binding:"required,gtfield=time.Now"`
	Tags      []string  `json:"tags" binding:"omitempty,dive,min=1,max=50"`
}

// UpdateTopicRequest 更新话题请求
type UpdateTopicRequest struct {
	Title     string    `json:"title" binding:"required,min=1,max=255"`
	Content   string    `json:"content" binding:"required,min=1"`
	ExpiresAt time.Time `json:"expires_at" binding:"required,gtfield=time.Now"`
}

// TopicListRequest 话题列表请求
type TopicListRequest struct {
	Pagination
	Sort
	TagID  uint64 `form:"tag_id" binding:"omitempty,min=1"`
	UserID uint64 `form:"user_id" binding:"omitempty,min=1"`
}

// NearbyTopicsRequest 附近话题请求
type NearbyTopicsRequest struct {
	Pagination
	Location
}

// TopicInteractionRequest 话题互动请求
type TopicInteractionRequest struct {
	InteractionType string `json:"interaction_type" binding:"required,oneof=like favorite share"`
}

// AddTagsRequest 添加标签请求
type AddTagsRequest struct {
	Tags []string `json:"tags" binding:"required,min=1,dive,min=1,max=50"`
}

// RemoveTagsRequest 移除标签请求
type RemoveTagsRequest struct {
	TagIDs []uint64 `json:"tag_ids" binding:"required,min=1"`
}
