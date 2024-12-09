package handler

import (
	"encoding/json"
	"time"

	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/api/response"
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
)

// CreateTopic 创建新话题
// @Summary 创建话题
// @Description 创建一个新的话题,支持添加图片和标签
// @Tags 话题
// @Accept multipart/form-data,json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param request body request.CreateTopicRequest true "话题信息"
// @Param images formData file false "话题图片文件(可多张)"
// @Success 200 {object} response.Response{data=response.TopicResponse} "创建成功"
// @Failure 400,401,403,500 {object} response.Response "错误详情"
// @Router /api/v1/topics [post]
func (h *Handler) CreateTopic(c *gin.Context) {
	// 身份验证
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// 参数验证
	var req request.CreateTopicRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Error("创建话题参数验证失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err))
		Error(c, errors.ErrValidation.WithDetails(err.Error()))
		return
	}
	// 手动解析 tags 字段
	if tagsStr := c.PostForm("Tags"); tagsStr != "" {
		var tags []string
		if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
			logger.Error("Failed to parse tags JSON",
				logger.Any("error", err),
				logger.String("tags_str", tagsStr))
		} else {
			req.Tags = tags
		}
	}
	// 手动验证 ExpiresAt
	if !req.ExpiresAt.After(time.Now()) {
		logger.Error("ExpiresAt is not in the future",
			logger.Time("expires_at", req.ExpiresAt),
			logger.Time("current_time", time.Now()))
		Error(c, errors.ErrValidation.WithDetails("ExpiresAt must be in the future"))
		return
	}
	// 处理图片
	var images []*model.File
	form, err := c.MultipartForm()
	if err == nil && form != nil && form.File["images"] != nil {
		files := form.File["images"]
		images = make([]*model.File, len(files))
		for i, file := range files {
			images[i] = &model.File{
				File: file,
				Type: "image",
				Name: file.Filename,
				Size: uint(file.Size),
			}
		}
	}

	// 构建话题模型
	topic := &model.Topic{
		UserUID:           userUID,
		Title:             req.Title,
		Content:           req.Content,
		LocationLatitude:  req.Latitude,
		LocationLongitude: req.Longitude,
		ExpiresAt:         req.ExpiresAt,
	}

	// 创建话题
	createdTopic, err := h.topicService.CreateTopic(c.Request.Context(), userUID, topic, images)
	if err != nil {
		logger.Error("创建话题参数验证失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.Time("expires_at", req.ExpiresAt),
		)
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}
	// 打印请求数据
	logger.Info("Received create topic request",
		logger.Any("request", req))
	// 处理标签
	if len(req.Tags) > 0 {
		logger.Info("Adding tags to topic",
			logger.Any("tags", req.Tags),
			logger.String("topic_uid", topic.UID))

		if err := h.topicService.AddTags(c.Request.Context(), topic.UID, req.Tags); err != nil {
			logger.Error("failed to add tags",
				logger.Any("error", err),
				logger.String("topic_uid", topic.UID),
				logger.Any("tags", req.Tags))
		}
	}

	Success(c, response.ToTopicResponse(createdTopic))
}

// UpdateTopic 更新话题
// @Summary 更新话题
// @Description 更新指定话题的内容(仅话题创建者可操作)
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "话题UUID"
// @Param request body request.UpdateTopicRequest true "更新内容"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id} [put]
func (h *Handler) UpdateTopic(c *gin.Context) {
	// 身份验证
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// 获取话题UUID
	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	// 参数验证
	var req request.UpdateTopicRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Error("更新话题参数验证失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err))
		Error(c, errors.ErrValidation.WithDetails(err.Error()))
		return
	}

	// 构建更新模型
	topic := &model.Topic{
		BaseModel: model.BaseModel{
			UID: topicUID,
		},
		Title:     req.Title,
		Content:   req.Content,
		ExpiresAt: req.ExpiresAt,
	}

	// 执行更新
	if err := h.topicService.UpdateTopic(c.Request.Context(), userUID, topic); err != nil {
		logger.Error("更新话题失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID),
			logger.String("topic_uid", topicUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, nil)
}

// DeleteTopic 删除话题
// @Summary 删除话题
// @Description 删除指定的话题(仅话题创建者可操作)
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "话题UUID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id} [delete]
func (h *Handler) DeleteTopic(c *gin.Context) {
	// 身份验证
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// 获取话题UUID
	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	// 执行删除
	if err := h.topicService.DeleteTopic(c.Request.Context(), userUID, topicUID); err != nil {
		logger.Error("删除话题失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID),
			logger.String("topic_uid", topicUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, nil)
}

// GetTopic 获取话题详情
// @Summary 获取话题详情
// @Description 获取指定话题的详细信息
// @Tags 话题
// @Accept json
// @Produce json
// @Param id path string true "话题UUID"
// @Success 200 {object} response.Response{data=response.TopicDetailResponse} "话题详情"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id} [get]
func (h *Handler) GetTopic(c *gin.Context) {
	// 获取话题UUID
	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	// 异步增加浏览次数
	go h.topicService.ViewTopic(c.Request.Context(), topicUID)

	// 获取话题信息
	topic, err := h.topicService.GetTopicByUID(c.Request.Context(), topicUID)
	if err != nil {
		logger.Error("获取话题失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	if topic == nil {
		Error(c, errors.ErrTopicNotFound)
		return
	}

	// 获取当前用户的互动状态
	userUID := h.GetCurrentUserUID(c)
	var interactions []*model.TopicInteraction
	if userUID != "" {
		interactions, _ = h.topicService.GetInteractions(c.Request.Context(), topicUID, "")
	}

	Success(c, response.ToTopicDetailResponse(topic, interactions))
}

// ListTopics 获取话题列表
// @Summary 获取话题列表
// @Description 分页获取话题列表
// @Tags 话题
// @Accept json
// @Produce json
// @Param request query request.TopicListRequest true "查询参数"
// @Success 200 {object} response.Response{data=response.TopicListResponse} "话题列表"
// @Failure 400 {object} response.Response "错误详情"
// @Router /api/v1/topics [get]
func (h *Handler) ListTopics(c *gin.Context) {
	var req request.TopicListRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	topics, total, err := h.topicService.ListTopics(
		c.Request.Context(),
		req.Page,
		req.Size,
	)
	if err != nil {
		logger.Error("获取话题列表失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, response.ToTopicListResponse(topics, total, req.Page, req.Size))
}

// GetNearbyTopics 获取附近话题
// @Summary 获取附近话题
// @Description 根据给定坐标获取指定范围内的话题列表
// @Tags 话题
// @Accept json
// @Produce json
// @Param request query request.NearbyTopicsRequest true "附近话题请求"
// @Success 200 {object} response.Response{data=response.TopicListResponse}
// @Failure 400 {object} response.Response "错误详情"
// @Router /api/v1/topics/nearby [get]
func (h *Handler) GetNearbyTopics(c *gin.Context) {
	var req request.NearbyTopicsRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	// 验证位置参数
	if err := ValidateLocation(req.Latitude, req.Longitude, req.Radius); err != nil {
		Error(c, err)
		return
	}

	topics, total, err := h.topicService.GetNearbyTopics(
		c.Request.Context(),
		req.Latitude,
		req.Longitude,
		req.Radius,
		req.Page,
		req.Size,
	)
	if err != nil {
		logger.Error("获取附近话题失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, response.ToTopicListResponse(topics, total, req.Page, req.Size))
}

// AddTopicInteraction 添加话题互动
// @Summary 添加话题互动
// @Description 为话题添加点赞、收藏或分享等互动
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "话题UUID"
// @Param type path string true "互动类型(like/favorite/share)"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/interactions/{type} [post]
func (h *Handler) AddTopicInteraction(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	interactionType := c.Param("type")
	if !isValidInteractionType(interactionType) {
		Error(c, errors.ErrValidation.WithDetails("无效的互动类型"))
		return
	}

	if err := h.topicService.AddInteraction(c.Request.Context(), userUID, topicUID, interactionType); err != nil {
		logger.Error("添加话题互动失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID),
			logger.String("type", interactionType))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, nil)
}

// GetTopicInteractions 获取话题互动列表
// @Summary 获取话题互动列表
// @Description 获取指定话题的特定类型互动列表
// @Tags 话题
// @Accept json
// @Produce json
// @Param id path string true "话题UUID"
// @Param type path string true "互动类型(like/favorite/share)"
// @Success 200 {object} response.Response{data=[]response.TopicInteractionResponse} "互动列表"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/interactions/{type} [get]
func (h *Handler) GetTopicInteractions(c *gin.Context) {
	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	interactionType := c.Param("type")
	if !isValidInteractionType(interactionType) {
		Error(c, errors.ErrValidation.WithDetails("无效的互动类型"))
		return
	}

	interactions, err := h.topicService.GetInteractions(c.Request.Context(), topicUID, interactionType)
	if err != nil {
		logger.Error("获取话题互动失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID),
			logger.String("type", interactionType))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, response.ToInteractionListResponse(interactions, int64(len(interactions)), 1, len(interactions)))
}

// AddTags 添加话题标签
// @Summary 添加话题标签
// @Description 为指定话题添加一个或多个标签
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "话题UUID"
// @Param request body request.AddTagsRequest true "标签信息"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/tags [post]
func (h *Handler) AddTags(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	var req request.AddTagsRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	if err := h.topicService.AddTags(c.Request.Context(), topicUID, req.Tags); err != nil {
		logger.Error("添加话题标签失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, nil)
}

// RemoveTags 移除话题标签
// @Summary 移除话题标签
// @Description 从指定话题移除一个或多个标签
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "话题UUID"
// @Param request body request.RemoveTagsRequest true "标签UUID列表"
// @Success 200 {object} response.Response "移除成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/tags [delete]
func (h *Handler) RemoveTags(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	var req request.RemoveTagsRequest
	if err := BindAndValidate(c, &req); err != nil {
		Error(c, err)
		return
	}

	if err := h.topicService.RemoveTags(c.Request.Context(), topicUID, req.TagUIDs); err != nil {
		logger.Error("移除话题标签失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, nil)
}

// GetTopicTags 获取话题标签
// @Summary 获取话题标签
// @Description 获取指定话题的所有标签
// @Tags 话题
// @Accept json
// @Produce json
// @Param id path string true "话题UUID"
// @Success 200 {object} response.Response{data=[]response.TagInfo} "标签列表"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/tags [get]
func (h *Handler) GetTopicTags(c *gin.Context) {
	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	tags, err := h.topicService.GetTopicTags(c.Request.Context(), topicUID)
	if err != nil {
		logger.Error("获取话题标签失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, response.ToTagInfoList(tags))
}

// GetPopularTags 获取热门标签
// @Summary 获取热门标签
// @Description 获取使用次数最多的标签列表
// @Tags 话题
// @Accept json
// @Produce json
// @Param limit query int false "返回数量" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=[]response.TagInfo} "标签列表"
// @Failure 400 {object} response.Response "错误详情"
// @Router /api/v1/tags/popular [get]
func (h *Handler) GetPopularTags(c *gin.Context) {
	var req request.GetPopularTagsRequest
	if err := BindQuery(c, &req); err != nil {
		Error(c, err)
		return
	}

	tags, err := h.topicService.GetPopularTags(c.Request.Context(), req.Limit)
	if err != nil {
		logger.Error("获取热门标签失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.Int("limit", req.Limit))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, response.ToTagInfoList(tags))
}

// isValidInteractionType 验证互动类型是否有效
func isValidInteractionType(t string) bool {
	validTypes := map[string]bool{
		"like":     true,
		"favorite": true,
		"share":    true,
	}
	return validTypes[t]
}

// ListUserTopics 获取用户的话题列表
// @Summary 获取用户话题列表
// @Description 分页获取指定用户发布的所有话题
// @Tags 话题
// @Accept json
// @Produce json
// @Param user_uid path string true "用户UUID"
// @Param page query int true "页码" minimum(1)
// @Param size query int true "每页大小" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.TopicListResponse} "话题列表"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/users/{user_uid} [get]
func (h *Handler) ListUserTopics(c *gin.Context) {
	userUID, err := ParseUUID(c, "user_uid")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的用户ID"))
		return
	}

	page, size, err := GetPagination(c)
	if err != nil {
		Error(c, err)
		return
	}

	topics, total, err := h.topicService.ListUserTopics(c.Request.Context(), userUID, page, size)
	if err != nil {
		logger.Error("获取用户话题列表失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("user_uid", userUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, response.ToTopicListResponse(topics, total, page, size))
}

// AddTopicImage 添加话题图片
// @Summary 为话题添加图片
// @Description 为指定话题添加一张或多张图片
// @Tags 话题
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "话题UUID"
// @Param images formData file true "图片文件(可多张)"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/images [post]
func (h *Handler) AddTopicImage(c *gin.Context) {
	// 身份验证
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	// 获取话题UUID
	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		Error(c, errors.ErrInvalidFileUpload.WithDetails("上传文件解析失败"))
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		Error(c, errors.ErrValidation.WithDetails("未上传任何图片"))
		return
	}

	// 构造图片文件列表
	images := make([]*model.File, len(files))
	for i, file := range files {
		images[i] = &model.File{
			File: file,
			Type: "image",
			Name: file.Filename,
			Size: uint(file.Size),
		}
	}

	// 添加图片
	if err := h.topicService.AddTopicImage(c.Request.Context(), userUID, topicUID, images); err != nil {
		logger.Error("添加话题图片失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, nil)
}

// RemoveTopicInteraction 移除话题互动
// @Summary 移除话题互动
// @Description 移除指定话题的点赞、收藏或分享等互动
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path string true "话题UUID"
// @Param type path string true "互动类型(like/favorite/share)"
// @Success 200 {object} response.Response "移除成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/interactions/{type} [delete]
func (h *Handler) RemoveTopicInteraction(c *gin.Context) {
	userUID := h.GetCurrentUserUID(c)
	if userUID == "" {
		Error(c, errors.ErrUnauthorized)
		return
	}

	topicUID, err := ParseUUID(c, "id")
	if err != nil {
		Error(c, errors.ErrValidation.WithDetails("无效的话题ID"))
		return
	}

	interactionType := c.Param("type")
	if !isValidInteractionType(interactionType) {
		Error(c, errors.ErrValidation.WithDetails("无效的互动类型"))
		return
	}

	if err := h.topicService.RemoveInteraction(c.Request.Context(), userUID, topicUID, interactionType); err != nil {
		logger.Error("移除话题互动失败",
			logger.String("path", c.Request.URL.Path),
			logger.Any("error", err),
			logger.String("topic_uid", topicUID),
			logger.String("type", interactionType))
		Error(c, errors.ErrOperation.WithDetails(err.Error()))
		return
	}

	Success(c, nil)
}
