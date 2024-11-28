package response

import (
	"time"

	"github.com/chiliososada/distance-back/internal/model"
)

// TopicResponse 话题响应基础结构
type TopicResponse struct {
	ID                uint64      `json:"id"`
	Title             string      `json:"title"`
	Content           string      `json:"content"`
	User              *UserBrief  `json:"user"` // 使用已有的 UserBrief
	Images            []ImageInfo `json:"images,omitempty"`
	Tags              []TagInfo   `json:"tags,omitempty"`
	Location          *Location   `json:"location"` // 使用已有的 Location
	Distance          float64     `json:"distance,omitempty"`
	LikesCount        uint        `json:"likes_count"`
	ViewsCount        uint        `json:"views_count"`
	SharesCount       uint        `json:"shares_count"`
	ParticipantsCount uint        `json:"participants_count"`
	Status            string      `json:"status"`
	ExpiresAt         time.Time   `json:"expires_at"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// ImageInfo 图片信息
type ImageInfo struct {
	ID        uint64    `json:"id"`
	URL       string    `json:"url"`
	Width     uint      `json:"width"`
	Height    uint      `json:"height"`
	Size      uint      `json:"size"`
	SortOrder uint      `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// TagInfo 标签信息
type TagInfo struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	UseCount uint   `json:"use_count"`
}

// TopicDetailResponse 话题详情响应
type TopicDetailResponse struct {
	TopicResponse
	Interactions *InteractionInfo `json:"interactions,omitempty"`
}

// InteractionInfo 互动信息
type InteractionInfo struct {
	HasLiked     bool `json:"has_liked"`
	HasFavorited bool `json:"has_favorited"`
	HasShared    bool `json:"has_shared"`
}

// TopicInteractionResponse 话题互动响应
type TopicInteractionResponse struct {
	ID              uint64     `json:"id"`
	TopicID         uint64     `json:"topic_id"`
	User            *UserBrief `json:"user"`
	InteractionType string     `json:"interaction_type"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
}

// 使用泛型定义分页响应类型
type TopicListResponse = PaginatedResponse[*TopicResponse]
type TopicInteractionListResponse = PaginatedResponse[*TopicInteractionResponse]
type TagListResponse = PaginatedResponse[*TagInfo]

// Convert Functions

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
		UpdatedAt:         topic.UpdatedAt,
		Location: &Location{
			Latitude:  topic.LocationLatitude,
			Longitude: topic.LocationLongitude,
		},
	}

	// Convert user info
	if topic.User.ID != 0 {
		resp.User = ToUserBrief(&topic.User)
	}

	// Convert images
	if len(topic.TopicImages) > 0 {
		resp.Images = make([]ImageInfo, len(topic.TopicImages))
		for i, img := range topic.TopicImages {
			resp.Images[i] = ImageInfo{
				ID:        img.ID,
				URL:       img.ImageURL,
				Width:     img.ImageWidth,
				Height:    img.ImageHeight,
				Size:      img.FileSize,
				SortOrder: img.SortOrder,
				CreatedAt: img.CreatedAt,
			}
		}
	}

	return resp
}

func ToTopicDetailResponse(topic *model.Topic, interactions []*model.TopicInteraction) *TopicDetailResponse {
	if topic == nil {
		return nil
	}

	resp := &TopicDetailResponse{
		TopicResponse: *ToTopicResponse(topic),
		Interactions:  &InteractionInfo{},
	}

	for _, interaction := range interactions {
		switch interaction.InteractionType {
		case "like":
			resp.Interactions.HasLiked = true
		case "favorite":
			resp.Interactions.HasFavorited = true
		case "share":
			resp.Interactions.HasShared = true
		}
	}

	return resp
}

func ToTopicInteractionResponse(interaction *model.TopicInteraction) *TopicInteractionResponse {
	if interaction == nil {
		return nil
	}

	return &TopicInteractionResponse{
		ID:              interaction.ID,
		TopicID:         interaction.TopicID,
		User:            ToUserBrief(&interaction.User),
		InteractionType: interaction.InteractionType,
		Status:          interaction.InteractionStatus,
		CreatedAt:       interaction.CreatedAt,
	}
}

func ToTopicListResponse(topics []*model.Topic, total int64, page, size int) *TopicListResponse {
	list := make([]*TopicResponse, len(topics))
	for i, topic := range topics {
		list[i] = ToTopicResponse(topic)
	}
	return NewPaginatedResponse(list, total, page, size)
}

func ToInteractionListResponse(interactions []*model.TopicInteraction, total int64, page, size int) *TopicInteractionListResponse {
	list := make([]*TopicInteractionResponse, len(interactions))
	for i, interaction := range interactions {
		list[i] = ToTopicInteractionResponse(interaction)
	}
	return NewPaginatedResponse(list, total, page, size)
}

func ToTagInfoList(tags []*model.Tag) []*TagInfo {
	if tags == nil {
		return nil
	}

	list := make([]*TagInfo, len(tags))
	for i, tag := range tags {
		list[i] = &TagInfo{
			ID:       tag.ID,
			Name:     tag.Name,
			UseCount: tag.UseCount,
		}
	}
	return list
}

func ToTagListResponse(tags []*model.Tag, total int64, page, size int) *TagListResponse {
	tagInfos := ToTagInfoList(tags)
	return NewPaginatedResponse(tagInfos, total, page, size)
}
