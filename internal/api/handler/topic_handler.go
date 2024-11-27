package handler

import (
	"time"

	"DistanceBack_v1/internal/model"
	"DistanceBack_v1/internal/service"
	"DistanceBack_v1/pkg/logger"

	"github.com/gin-gonic/gin"
)

// CreateTopicRequest 创建话题请求
type CreateTopicRequest struct {
	Title     string    `json:"title" binding:"required,min=1,max=255"`
	Content   string    `json:"content" binding:"required,min=1"`
	Latitude  float64   `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64   `json:"longitude" binding:"required,min=-180,max=180"`
	ExpiresAt time.Time `json:"expires_at" binding:"required,gtfield=time.Now"`
	Tags      []string  `json:"tags" binding:"omitempty,dive,min=1,max=50"`
}

// UpdateTopicRequest 更新话题请求
type UpdateTopicRequest struct {
	Title     string    `json:"title" binding:"required,min=1,max=255"`
	Content   string    `json:"content" binding:"required,min=1"`
	ExpiresAt time.Time `json:"expires_at" binding:"required,gtfield=time.Now"`
}

// CreateTopic 创建话题
func (h *Handler) CreateTopic(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 解析多部分表单数据
	var req CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 处理图片文件
	form, err := c.MultipartForm()
	if err != nil {
		logger.Warn("Failed to get multipart form", logger.Any("error", err))
	}

	var images []*model.File
	if form != nil && form.File["images"] != nil {
		files := form.File["images"]
		images = make([]*model.File, 0, len(files))
		for _, file := range files {
			images = append(images, &model.File{
				File: file,
				Type: "image",
				Name: file.Filename,
				Size: uint(file.Size),
			})
		}
	}

	topic := &model.Topic{
		Title:             req.Title,
		Content:           req.Content,
		LocationLatitude:  req.Latitude,
		LocationLongitude: req.Longitude,
		ExpiresAt:         req.ExpiresAt,
	}

	createdTopic, err := h.topicService.CreateTopic(c, userID, topic, images)
	if err != nil {
		Error(c, err)
		return
	}

	// 添加标签
	if len(req.Tags) > 0 {
		if err := h.topicService.AddTags(c, topic.ID, req.Tags); err != nil {
			logger.Error("Failed to add tags",
				logger.Any("error", err),
				logger.Uint64("topic_id", topic.ID))
		}
	}

	Success(c, createdTopic)
}

// UpdateTopic 更新话题
func (h *Handler) UpdateTopic(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var req UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	topic := &model.Topic{
		BaseModel: model.BaseModel{ID: topicID},
		Title:     req.Title,
		Content:   req.Content,
		ExpiresAt: req.ExpiresAt,
	}

	if err := h.topicService.UpdateTopic(c, userID, topic); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// DeleteTopic 删除话题
func (h *Handler) DeleteTopic(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.topicService.DeleteTopic(c, userID, topicID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetTopic 获取话题详情
func (h *Handler) GetTopic(c *gin.Context) {
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 增加浏览次数
	go h.topicService.ViewTopic(c, topicID)

	topic, err := h.topicService.GetTopicByID(c, topicID)
	if err != nil {
		Error(c, err)
		return
	}
	if topic == nil {
		Error(c, service.ErrTopicNotFound)
		return
	}

	Success(c, topic)
}

// ListTopics 获取话题列表
func (h *Handler) ListTopics(c *gin.Context) {
	var query PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	topics, total, err := h.topicService.ListTopics(c, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"topics": topics,
		"total":  total,
		"page":   query.Page,
		"size":   query.PageSize,
	})
}

// ListUserTopics 获取用户的话题列表
func (h *Handler) ListUserTopics(c *gin.Context) {
	userID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var query PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	topics, total, err := h.topicService.ListUserTopics(c, userID, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"topics": topics,
		"total":  total,
		"page":   query.Page,
		"size":   query.PageSize,
	})
}

// GetNearbyTopics 获取附近的话题
func (h *Handler) GetNearbyTopics(c *gin.Context) {
	var query struct {
		PaginationQuery
		Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90"`
		Longitude float64 `form:"longitude" binding:"required,min=-180,max=180"`
		Radius    float64 `form:"radius" binding:"required,min=0,max=50000"` // 最大50km
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	topics, total, err := h.topicService.GetNearbyTopics(c, query.Latitude, query.Longitude, query.Radius, query.Page, query.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{
		"topics": topics,
		"total":  total,
		"page":   query.Page,
		"size":   query.PageSize,
	})
}

// AddTopicImage 添加话题图片
func (h *Handler) AddTopicImage(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	image := &model.File{
		File: file,
		Type: "image",
		Name: file.Filename,
		Size: uint(file.Size),
	}

	// 将单个 image 包装为切片
	images := []*model.File{image}

	// 调用服务方法
	if err := h.topicService.AddTopicImage(c.Request.Context(), topicID, images); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// AddTopicInteraction 添加话题互动（点赞、收藏、分享）
func (h *Handler) AddTopicInteraction(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	interactionType := c.Param("type")
	if err := h.topicService.AddInteraction(c, userID, topicID, interactionType); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RemoveTopicInteraction 移除话题互动
func (h *Handler) RemoveTopicInteraction(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	interactionType := c.Param("type")
	if err := h.topicService.RemoveInteraction(c, userID, topicID, interactionType); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetTopicInteractions 获取话题互动列表
func (h *Handler) GetTopicInteractions(c *gin.Context) {
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	interactionType := c.Param("type")
	interactions, err := h.topicService.GetInteractions(c, topicID, interactionType)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, interactions)
}

// AddTags 添加话题标签
func (h *Handler) AddTags(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var req struct {
		Tags []string `json:"tags" binding:"required,min=1,dive,min=2,max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.topicService.AddTags(c, topicID, req.Tags); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RemoveTags 移除话题标签
func (h *Handler) RemoveTags(c *gin.Context) {
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	var req struct {
		TagIDs []uint64 `json:"tag_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	if err := h.topicService.RemoveTags(c, topicID, req.TagIDs); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetTopicTags 获取话题标签
func (h *Handler) GetTopicTags(c *gin.Context) {
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	tags, err := h.topicService.GetTopicTags(c, topicID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tags)
}

// GetPopularTags 获取热门标签
func (h *Handler) GetPopularTags(c *gin.Context) {
	var query struct {
		Limit int `form:"limit" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	tags, err := h.topicService.GetPopularTags(c, query.Limit)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, tags)
}
