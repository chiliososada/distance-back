package handler

import (
	"DistanceBack_v1/internal/api/request"
	"DistanceBack_v1/internal/api/response"
	"DistanceBack_v1/internal/model"
	"DistanceBack_v1/internal/service"
	"DistanceBack_v1/pkg/logger"

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
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/topics [post]
func (h *Handler) CreateTopic(c *gin.Context) {
	// 1. 获取并验证当前用户
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 解析请求参数
	var req request.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("话题参数解析失败",
			logger.Any("error", err),
			logger.Uint64("user_id", userID))
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的请求参数"))
		return
	}

	// 3. 处理上传的图片
	var images []*model.File
	form, err := c.MultipartForm()
	if err == nil && form != nil && form.File["images"] != nil {
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

	// 4. 构建话题模型
	topic := &model.Topic{
		UserID:            userID,
		Title:             req.Title,
		Content:           req.Content,
		LocationLatitude:  req.Latitude,
		LocationLongitude: req.Longitude,
		ExpiresAt:         req.ExpiresAt,
		Status:            "active", // 设置初始状态
	}

	// 5. 调用服务创建话题
	createdTopic, err := h.topicService.CreateTopic(c, userID, topic, images)
	if err != nil {
		logger.Error("创建话题失败",
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
			logger.Any("topic", topic))
		Error(c, err)
		return
	}

	// 6. 处理标签
	if len(req.Tags) > 0 {
		if err := h.topicService.AddTags(c, createdTopic.ID, req.Tags); err != nil {
			logger.Error("添加标签失败",
				logger.Any("error", err),
				logger.Uint64("topic_id", createdTopic.ID),
				logger.Any("tags", req.Tags))
			// 标签添加失败不影响话题创建的结果
		}
	}

	// 7. 转换并返回响应
	Success(c, response.ToTopicResponse(createdTopic))
}

// UpdateTopic 更新话题
// @Summary 更新话题
// @Description 更新指定话题的内容(仅话题创建者可操作)
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path uint64 true "话题ID"
// @Param request body request.UpdateTopicRequest true "更新内容"
// @Success 200 {object} response.Response{data=response.TopicResponse} "更新成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id} [put]
func (h *Handler) UpdateTopic(c *gin.Context) {
	// 1. 身份验证
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取话题ID
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 3. 解析请求参数
	var req request.UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("更新参数解析失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 4. 构建更新模型
	topic := &model.Topic{
		BaseModel: model.BaseModel{ID: topicID},
		Title:     req.Title,
		Content:   req.Content,
		ExpiresAt: req.ExpiresAt,
	}

	// 5. 执行更新
	if err := h.topicService.UpdateTopic(c, userID, topic); err != nil {
		logger.Error("更新话题失败",
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
			logger.Uint64("topic_id", topicID))
		Error(c, err)
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
// @Param id path uint64 true "话题ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id} [delete]
func (h *Handler) DeleteTopic(c *gin.Context) {
	// 1. 身份验证
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取话题ID
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		logger.Error("话题ID解析失败",
			logger.Any("error", err),
			logger.String("id", c.Param("id")))
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 3. 执行删除操作
	if err := h.topicService.DeleteTopic(c, userID, topicID); err != nil {
		logger.Error("删除话题失败",
			logger.Any("error", err),
			logger.Uint64("user_id", userID),
			logger.Uint64("topic_id", topicID))
		Error(c, err)
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
// @Param id path uint64 true "话题ID"
// @Param Authorization header string false "Bearer 用户令牌(可选)"
// @Success 200 {object} response.Response{data=response.TopicDetailResponse} "话题详情"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id} [get]
func (h *Handler) GetTopic(c *gin.Context) {
	// 1. 获取话题ID
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 2. 异步增加浏览次数
	go h.topicService.ViewTopic(c, topicID)

	// 3. 获取话题信息
	topic, err := h.topicService.GetTopicByID(c, topicID)
	if err != nil {
		logger.Error("获取话题失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
		Error(c, err)
		return
	}

	if topic == nil {
		Error(c, service.ErrTopicNotFound)
		return
	}

	// 4. 获取当前用户的互动状态(如果已登录)
	userID := h.GetCurrentUserID(c)
	var interaction *model.TopicInteraction
	if userID != 0 {
		interactions, _ := h.topicService.GetInteractions(c, topicID, "")
		if len(interactions) > 0 {
			interaction = interactions[0]
		}
	}

	// 5. 转换并返回响应
	Success(c, response.ToTopicDetailResponse(topic, interaction))
}

// ListTopics 获取话题列表
// @Summary 获取话题列表
// @Description 分页获取话题列表
// @Tags 话题
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页大小" minimum(1) maximum(100)
// @Param tag_id query uint64 false "标签ID"
// @Param user_id query uint64 false "用户ID"
// @Success 200 {object} response.Response{data=response.TopicListResponse} "话题列表"
// @Failure 400 {object} response.Response "错误详情"
// @Router /api/v1/topics [get]
func (h *Handler) ListTopics(c *gin.Context) {
	// 1. 获取查询参数
	var query request.TopicListRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 2. 获取话题列表
	topics, total, err := h.topicService.ListTopics(c, query.Page, query.PageSize)
	if err != nil {
		logger.Error("获取话题列表失败",
			logger.Any("error", err),
			logger.Any("query", query))
		Error(c, err)
		return
	}

	// 3. 转换并返回响应
	Success(c, response.ToTopicListResponse(topics, total, query.Page, query.PageSize))
}

// ListUserTopics 获取用户的话题列表
// @Summary 获取用户话题列表
// @Description 分页获取指定用户发布的所有话题
// @Tags 话题
// @Accept json
// @Produce json
// @Param id path uint64 true "用户ID"
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页大小" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.TopicListResponse} "话题列表"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/users/{id} [get]
func (h *Handler) ListUserTopics(c *gin.Context) {
	// 1. 获取目标用户ID
	targetUserID, err := ParseUint64Param(c, "id")
	if err != nil {
		logger.Error("解析用户ID失败",
			logger.Any("error", err),
			logger.String("id", c.Param("id")))
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 2. 获取分页参数
	pagination, err := GetPagination(c)
	if err != nil {
		logger.Error("解析分页参数失败",
			logger.Any("error", err),
			logger.Uint64("target_user_id", targetUserID))
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 3. 获取用户话题列表
	topics, total, err := h.topicService.ListUserTopics(c, targetUserID, pagination.Page, pagination.PageSize)
	if err != nil {
		logger.Error("获取用户话题列表失败",
			logger.Any("error", err),
			logger.Uint64("target_user_id", targetUserID),
			logger.Int("page", pagination.Page),
			logger.Int("page_size", pagination.PageSize))
		Error(c, err)
		return
	}

	// 4. 转换并返回响应
	Success(c, response.ToTopicListResponse(topics, total, pagination.Page, pagination.PageSize))
}

// GetNearbyTopics 获取附近的话题
// GetNearbyTopics 获取附近话题
// @Summary 获取附近话题
// @Description 根据给定坐标获取指定范围内的话题列表
// @Tags 话题
// @Accept json
// @Produce json
// @Param latitude query number true "纬度" minimum(-90) maximum(90)
// @Param longitude query number true "经度" minimum(-180) maximum(180)
// @Param radius query number true "范围(米)" minimum(0) maximum(50000)
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页大小" minimum(1) maximum(100)
// @Success 200 {object} response.Response{data=response.TopicListResponse} "话题列表"
// @Failure 400 {object} response.Response "错误详情"
// @Router /api/v1/topics/nearby [get]
func (h *Handler) GetNearbyTopics(c *gin.Context) {
	// 1. 获取查询参数
	var query request.NearbyTopicsRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 2. 获取附近话题
	topics, total, err := h.topicService.GetNearbyTopics(
		c,
		query.Latitude,
		query.Longitude,
		query.Radius,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		logger.Error("获取附近话题失败",
			logger.Any("error", err),
			logger.Any("query", query))
		Error(c, err)
		return
	}

	// 3. 转换并返回响应
	Success(c, response.ToTopicListResponse(topics, total, query.Page, query.PageSize))
}

// AddTopicImage 添加话题图片
// @Summary 为话题添加图片
// @Description 为指定话题添加一张或多张图片
// @Tags 话题
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path uint64 true "话题ID"
// @Param image formData file true "图片文件"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/images [post]
func (h *Handler) AddTopicImage(c *gin.Context) {
	// 1. 身份验证
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取话题ID
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 3. 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		logger.Error("获取上传文件失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 4. 构建文件模型
	image := &model.File{
		File: file,
		Type: "image",
		Name: file.Filename,
		Size: uint(file.Size),
	}
	images := []*model.File{image}

	// 5. 处理图片上传
	if err := h.topicService.AddTopicImage(c, topicID, images); err != nil {
		logger.Error("添加话题图片失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
		Error(c, err)
		return
	}

	Success(c, nil)
}

// AddTopicInteraction 添加话题互动
// @Summary 添加话题互动
// @Description 为话题添加点赞、收藏或分享等互动
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path uint64 true "话题ID"
// @Param type path string true "互动类型(like/favorite/share)"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/interactions/{type} [post]
func (h *Handler) AddTopicInteraction(c *gin.Context) {
	// 1. 身份验证
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取参数
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	interactionType := c.Param("type")
	if !isValidInteractionType(interactionType) {
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的互动类型"))
		return
	}

	// 3. 添加互动
	if err := h.topicService.AddInteraction(c, userID, topicID, interactionType); err != nil {
		logger.Error("添加话题互动失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID),
			logger.String("type", interactionType))
		Error(c, err)
		return
	}

	Success(c, nil)
}

// RemoveTopicInteraction 移除话题互动
// @Summary 移除话题互动
// @Description 移除指定的话题互动
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path uint64 true "话题ID"
// @Param type path string true "互动类型(like/favorite/share)"
// @Success 200 {object} response.Response "移除成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/interactions/{type} [delete]
func (h *Handler) RemoveTopicInteraction(c *gin.Context) {
	// 1. 身份验证
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取参数
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	interactionType := c.Param("type")
	if !isValidInteractionType(interactionType) {
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的互动类型"))
		return
	}

	// 3. 移除互动
	if err := h.topicService.RemoveInteraction(c, userID, topicID, interactionType); err != nil {
		logger.Error("移除话题互动失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID),
			logger.String("type", interactionType))
		Error(c, err)
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
// @Param id path uint64 true "话题ID"
// @Param type path string true "互动类型(like/favorite/share)"
// @Success 200 {object} response.Response "互动列表"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/interactions/{type} [get]
func (h *Handler) GetTopicInteractions(c *gin.Context) {
	// 1. 获取参数
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	interactionType := c.Param("type")
	if !isValidInteractionType(interactionType) {
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的互动类型"))
		return
	}

	// 2. 获取互动列表
	interactions, err := h.topicService.GetInteractions(c, topicID, interactionType)
	if err != nil {
		logger.Error("获取话题互动失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID),
			logger.String("type", interactionType))
		Error(c, err)
		return
	}

	Success(c, response.ToTopicInteractionsResponse(interactions))
}

// isValidInteractionType 验证互动类型是否有效
func isValidInteractionType(t string) bool {
	return t == "like" || t == "favorite" || t == "share"
}

// AddTags 添加话题标签
// @Summary 添加话题标签
// @Description 为指定话题添加一个或多个标签
// @Tags 话题
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param id path uint64 true "话题ID"
// @Param request body request.AddTagsRequest true "标签信息"
// @Success 200 {object} response.Response "添加成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/tags [post]
func (h *Handler) AddTags(c *gin.Context) {
	// 1. 身份验证
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取话题ID
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		logger.Error("解析话题ID失败",
			logger.Any("error", err),
			logger.String("id", c.Param("id")))
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 3. 解析请求参数
	var req request.AddTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("解析标签请求失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的标签参数"))
		return
	}

	// 4. 添加标签
	if err := h.topicService.AddTags(c, topicID, req.Tags); err != nil {
		logger.Error("添加话题标签失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID),
			logger.Any("tags", req.Tags))
		Error(c, err)
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
// @Param id path uint64 true "话题ID"
// @Param request body request.RemoveTagsRequest true "标签ID列表"
// @Success 200 {object} response.Response "移除成功"
// @Failure 400,401,403,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/tags [delete]
func (h *Handler) RemoveTags(c *gin.Context) {
	// 1. 身份验证
	userID := h.GetCurrentUserID(c)
	if userID == 0 {
		Error(c, service.ErrUnauthorized)
		return
	}

	// 2. 获取话题ID
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		logger.Error("解析话题ID失败",
			logger.Any("error", err),
			logger.String("id", c.Param("id")))
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 3. 解析请求参数
	var req request.RemoveTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("解析标签请求失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
		Error(c, service.NewError(service.CodeInvalidRequest, "无效的标签参数"))
		return
	}

	// 4. 验证权限并移除标签
	if err := h.topicService.RemoveTags(c, topicID, req.TagIDs); err != nil {
		logger.Error("移除话题标签失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID),
			logger.Any("tag_ids", req.TagIDs))
		Error(c, err)
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
// @Param id path uint64 true "话题ID"
// @Success 200 {object} response.Response{data=[]response.TagInfo} "标签列表"
// @Failure 400,404 {object} response.Response "错误详情"
// @Router /api/v1/topics/{id}/tags [get]
func (h *Handler) GetTopicTags(c *gin.Context) {
	// 1. 获取话题ID
	topicID, err := ParseUint64Param(c, "id")
	if err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 2. 获取标签列表
	tags, err := h.topicService.GetTopicTags(c, topicID)
	if err != nil {
		logger.Error("获取话题标签失败",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
		Error(c, err)
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
// @Router /api/v1/topics/tags/popular [get]
func (h *Handler) GetPopularTags(c *gin.Context) {
	// 1. 获取参数
	var query struct {
		Limit int `form:"limit" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, service.ErrInvalidRequest)
		return
	}

	// 2. 获取热门标签
	tags, err := h.topicService.GetPopularTags(c, query.Limit)
	if err != nil {
		logger.Error("获取热门标签失败",
			logger.Any("error", err),
			logger.Int("limit", query.Limit))
		Error(c, err)
		return
	}

	Success(c, response.ToTagInfoList(tags))
}
