package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"time"
)

// CreateTopicRequest 创建话题请求
type CreateTopicRequest struct {
	Uid       string    `json:"uid" binding:"required"`
	Title     string    `json:"title" binding:"required,min=1,max=255"`
	Content   string    `json:"content" binding:"omitempty,max=4096"` //must present, can be empty, <=4096
	Images    []string  `json:"images" binding:"required,dive,min=1"` //must present, can be empty, entries cannot be empty
	Tags      []string  `json:"tags" binding:"required,dive,min=1,max=50"`
	Latitude  *float64  `json:"latitude" binding:"omitempty,min=-90,max=90"` //absent means unknown, can be 0
	Longitude *float64  `json:"longitude" binding:"omitempty,min=-180,max=180"`
	ExpiresAt time.Time `json:"expires_at" binding:"required"`
}

// UpdateTopicRequest 更新话题请求
type UpdateTopicRequest struct {
	Title     string                  `json:"title" binding:"required,min=1,max=255"`
	Content   string                  `json:"content" binding:"required,min=1"`
	ExpiresAt time.Time               `json:"expires_at" binding:"omitempty"`
	Tags      []string                `json:"tags,omitempty" binding:"omitempty,dive,min=1,max=50"`
	Images    []*multipart.FileHeader `form:"images" binding:"omitempty,dive"`
	// 可以添加删除图片的字段
	RemoveImageUIDs []string `json:"remove_image_uids,omitempty" binding:"omitempty,dive,uuid"`
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

type FindTopicsBy int

const (
	FindTopicsByUnknown FindTopicsBy = iota
	FindTopicsByRecency
	FindTopicsByPopularity
)

var byToString = map[FindTopicsBy]string{
	FindTopicsByUnknown:    "unknown",
	FindTopicsByRecency:    "recent",
	FindTopicsByPopularity: "popular",
}
var stringToBy = map[string]FindTopicsBy{
	"recent":  FindTopicsByRecency,
	"popular": FindTopicsByPopularity,
}

func (a *FindTopicsBy) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if by := stringToBy[s]; by == FindTopicsByUnknown {
		return errors.New("unknown find topics by parameter")
	} else {
		*a = by
		return nil
	}

}
func (a FindTopicsBy) MarshalJSON() ([]byte, error) {
	if a == FindTopicsByUnknown {
		return nil, errors.New(fmt.Sprintf("unknown find topic by: %v", a))
	} else {
		return json.Marshal(byToString[a])
	}
}

type FindTopicsByRequest struct {
	FindBy       FindTopicsBy `json:"findby"`
	Max          int          `json:"max"`
	RecencyScore int          `json:"recency,omitempty"`
}
