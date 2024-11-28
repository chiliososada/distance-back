// Topic request structures
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
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"page_size" binding:"required,min=1,max=100"`
	TagID     uint64 `form:"tag_id" binding:"omitempty,min=1"`
	UserID    uint64 `form:"user_id" binding:"omitempty,min=1"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=created_at likes_count views_count"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// NearbyTopicsRequest 附近话题请求
type NearbyTopicsRequest struct {
	Page      int     `form:"page" binding:"required,min=1"`
	PageSize  int     `form:"page_size" binding:"required,min=1,max=100"`
	Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"longitude" binding:"required,min=-180,max=180"`
	Radius    float64 `form:"radius" binding:"required,min=0,max=50000"`
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
	TagIDs []uint64 `json:"tag_ids" binding:"required,min=1,dive,min=1"`
}

// GetPopularTagsRequest 获取热门标签请求
type GetPopularTagsRequest struct {
	Limit int `form:"limit" binding:"required,min=1,max=100"`
}
