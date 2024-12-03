package request

import (
	"mime/multipart"
	"time"
)

// CreateTopicRequest 创建话题请求
type CreateTopicRequest struct {
	Title     string                  `json:"title" binding:"required,min=1,max=255"`
	Content   string                  `json:"content" binding:"required,min=1"`
	Images    []*multipart.FileHeader `form:"images" binding:"omitempty,dive"`
	Tags      []string                `json:"tags" binding:"omitempty,dive,min=1,max=50"`
	Latitude  float64                 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64                 `json:"longitude" binding:"required,min=-180,max=180"`
	ExpiresAt time.Time               `json:"expires_at" binding:"required,gtfield=time.Now"`
}

// UpdateTopicRequest 更新话题请求
type UpdateTopicRequest struct {
	Title     string    `json:"title" binding:"required,min=1,max=255"`
	Content   string    `json:"content" binding:"required,min=1"`
	ExpiresAt time.Time `json:"expires_at" binding:"required,gtfield=time.Now"`
}

// TopicListRequest 话题列表请求
type TopicListRequest struct {
	PaginationQuery
	TagUID    string `form:"tag_uid" binding:"omitempty,uuid"`
	UserUID   string `form:"user_uid" binding:"omitempty,uuid"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count views_count"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// NearbyTopicsRequest 附近话题请求
type NearbyTopicsRequest struct {
	PaginationQuery
	LocationQuery
}

// TopicImageRequest 话题图片请求
type TopicImageRequest struct {
	Image *multipart.FileHeader `form:"image" binding:"required"`
}

// TopicInteractionRequest 话题互动请求
type TopicInteractionRequest struct {
	InteractionType string `json:"interaction_type" binding:"required,oneof=like favorite share"`
}

// AddTagsRequest 添加标签请求
type AddTagsRequest struct {
	Tags []string `json:"tags" binding:"required,min=1,max=10,dive,min=1,max=50"`
}

// RemoveTagsRequest 移除标签请求
type RemoveTagsRequest struct {
	TagUIDs []string `json:"tag_uids" binding:"required,min=1,dive,uuid"`
}

// GetPopularTagsRequest 获取热门标签请求
type GetPopularTagsRequest struct {
	Limit int `form:"limit" binding:"required,min=1,max=100"`
}

// GetTopicRequest 获取话题请求
type GetTopicRequest struct {
	UIDParam
	IncludeDeleted bool `form:"include_deleted" json:"include_deleted"`
}

// GetTopicInteractionsRequest 获取话题互动列表请求
type GetTopicInteractionsRequest struct {
	PaginationQuery
	TopicUID        string `form:"topic_uid" binding:"required,uuid"`
	InteractionType string `form:"interaction_type" binding:"required,oneof=like favorite share"`
}

// BatchGetTopicsRequest 批量获取话题请求
type BatchGetTopicsRequest struct {
	TopicUIDs []string `json:"topic_uids" binding:"required,min=1,max=100,dive,uuid"`
}
