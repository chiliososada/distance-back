package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// UserBrief 用户简要信息
type UserBrief struct {
	ID        uint64 `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// TagInfo 标签信息
type TagInfo struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`      // 标签名称
	UseCount uint   `json:"use_count"` // 使用次数
}

// TopicImage 话题图片
type TopicImage struct {
	ID     uint64 `json:"id"`
	URL    string `json:"url"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Size   uint   `json:"size"`
}

// TopicResponse 话题基本响应
type TopicResponse struct {
	ID                uint64       `json:"id"`
	Title             string       `json:"title"`
	Content           string       `json:"content"`
	Location          *Location    `json:"location,omitempty"`
	User              *UserBrief   `json:"user"`
	Images            []TopicImage `json:"images,omitempty"`
	Tags              []TagInfo    `json:"tags,omitempty"`
	LikesCount        uint         `json:"likes_count"`
	ViewsCount        uint         `json:"views_count"`
	SharesCount       uint         `json:"shares_count"`
	ParticipantsCount uint         `json:"participants_count"`
	ExpiresAt         time.Time    `json:"expires_at"`
	Status            string       `json:"status"`
	CreatedAt         time.Time    `json:"created_at"`
	HasLiked          bool         `json:"has_liked"`
	HasFavorited      bool         `json:"has_favorited"`
	Distance          float64      `json:"distance,omitempty"`
}

// TopicInteractionListResponse 话题互动列表响应
type TopicInteractionListResponse struct {
	Interactions []*TopicInteractionResponse `json:"interactions"`
	Total        int64                       `json:"total"`
	Page         int                         `json:"page"`
	PageSize     int                         `json:"page_size"`
}

// TopicDetailResponse 话题详情响应
type TopicDetailResponse struct {
	TopicResponse
	UserInteraction *UserInteraction `json:"user_interaction,omitempty"`
}

// UserInteraction 用户与话题的互动状态
type UserInteraction struct {
	IsLiked     bool `json:"is_liked"`
	IsFavorited bool `json:"is_favorited"`
	IsShared    bool `json:"is_shared"`
}

// TopicListResponse 话题列表响应
type TopicListResponse struct {
	Topics     []*TopicResponse `json:"topics"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// TopicInteractionResponse 话题互动响应
type TopicInteractionResponse struct {
	ID                uint64     `json:"id"`
	TopicID           uint64     `json:"topic_id"`
	UserID            uint64     `json:"user_id"`
	User              *UserBrief `json:"user"`
	InteractionType   string     `json:"interaction_type"`   // like, favorite, share
	InteractionStatus string     `json:"interaction_status"` // active, cancelled
	CreatedAt         time.Time  `json:"created_at"`
}

// ToTopicResponse 将话题模型转换为响应
func ToTopicResponse(topic *model.Topic) *TopicResponse {
	if topic == nil {
		return nil
	}

	resp := &TopicResponse{
		ID:                topic.ID,
		Title:             topic.Title,
		Content:           topic.Content,
		LikesCount:        topic.LikesCount,
		ViewsCount:        topic.ViewsCount,
		SharesCount:       topic.SharesCount,
		ParticipantsCount: topic.ParticipantsCount,
		Status:            topic.Status,
		ExpiresAt:         topic.ExpiresAt,
		CreatedAt:         topic.CreatedAt,
	}

	// 位置信息
	if topic.LocationLatitude != 0 || topic.LocationLongitude != 0 {
		resp.Location = &Location{
			Latitude:  topic.LocationLatitude,
			Longitude: topic.LocationLongitude,
		}
	}

	// 用户信息
	if topic.User.ID != 0 {
		resp.User = &UserBrief{
			ID:        topic.User.ID,
			Nickname:  topic.User.Nickname,
			AvatarURL: topic.User.AvatarURL,
		}
	}

	return resp
}

// ToTopicDetailResponse 将话题模型转换为详情响应
func ToTopicDetailResponse(topic *model.Topic, interaction *model.TopicInteraction) *TopicDetailResponse {
	if topic == nil {
		return nil
	}

	detail := &TopicDetailResponse{
		TopicResponse: *ToTopicResponse(topic),
	}

	if interaction != nil {
		detail.UserInteraction = &UserInteraction{
			IsLiked:     interaction.InteractionType == "like" && interaction.InteractionStatus == "active",
			IsFavorited: interaction.InteractionType == "favorite" && interaction.InteractionStatus == "active",
			IsShared:    interaction.InteractionType == "share" && interaction.InteractionStatus == "active",
		}
	}

	return detail
}

// ToTopicListResponse 转换话题列表响应
func ToTopicListResponse(topics []*model.Topic, total int64, page, pageSize int) *TopicListResponse {
	topicResponses := make([]*TopicResponse, len(topics))
	for i, topic := range topics {
		topicResponses[i] = ToTopicResponse(topic)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &TopicListResponse{
		Topics:     topicResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// ToTopicInteractionResponse 将互动模型转换为响应
func ToTopicInteractionResponse(interaction *model.TopicInteraction) *TopicInteractionResponse {
	if interaction == nil {
		return nil
	}

	resp := &TopicInteractionResponse{
		ID:                interaction.ID,
		TopicID:           interaction.TopicID,
		UserID:            interaction.UserID,
		InteractionType:   interaction.InteractionType,
		InteractionStatus: interaction.InteractionStatus,
		CreatedAt:         interaction.CreatedAt,
	}

	// 转换用户信息
	if interaction.User.ID != 0 {
		resp.User = &UserBrief{
			ID:        interaction.User.ID,
			Nickname:  interaction.User.Nickname,
			AvatarURL: interaction.User.AvatarURL,
		}
	}

	return resp
}

// ToTopicInteractionsResponse 将互动列表转换为响应
func ToTopicInteractionsResponse(interactions []*model.TopicInteraction) []*TopicInteractionResponse {
	if interactions == nil {
		return nil
	}

	responses := make([]*TopicInteractionResponse, 0, len(interactions))
	for _, interaction := range interactions {
		if resp := ToTopicInteractionResponse(interaction); resp != nil {
			responses = append(responses, resp)
		}
	}

	return responses
}

// ToTagInfo 将标签模型转换为标签信息响应
func ToTagInfo(tag *model.Tag) *TagInfo {
	if tag == nil {
		return nil
	}

	return &TagInfo{
		ID:       tag.ID,
		Name:     tag.Name,
		UseCount: tag.UseCount,
	}
}

// ToTagInfoList 将标签列表转换为标签信息响应列表
func ToTagInfoList(tags []*model.Tag) []*TagInfo {
	if tags == nil {
		return nil
	}

	responses := make([]*TagInfo, 0, len(tags))
	for _, tag := range tags {
		if resp := ToTagInfo(tag); resp != nil {
			responses = append(responses, resp)
		}
	}

	return responses
}

// NewTopicInteractionListResponse 创建新的话题互动列表响应
func NewTopicInteractionListResponse(interactions []*model.TopicInteraction, total int64, page, pageSize int) *TopicInteractionListResponse {
	return &TopicInteractionListResponse{
		Interactions: ToTopicInteractionsResponse(interactions),
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}
}
