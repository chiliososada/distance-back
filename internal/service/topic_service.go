package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/cache"
	"github.com/chiliososada/distance-back/pkg/constants"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/chiliososada/distance-back/pkg/storage"
)

type TopicService struct {
	topicRepo    repository.TopicRepository
	userRepo     repository.UserRepository
	relationRepo repository.RelationshipRepository
	storage      storage.Storage
}

// NewTopicService 创建话题服务实例
func NewTopicService(
	topicRepo repository.TopicRepository,
	userRepo repository.UserRepository,
	relationRepo repository.RelationshipRepository,
	storage storage.Storage,
) *TopicService {
	return &TopicService{
		topicRepo:    topicRepo,
		userRepo:     userRepo,
		relationRepo: relationRepo,
		storage:      storage,
	}
}

// CreateTopic 创建话题
func (s *TopicService) CreateTopic(ctx context.Context, userID uint64, topic *model.Topic, images []*model.File) (*model.Topic, error) {
	// 验证用户状态
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.Status != "active" {
		return nil, ErrInvalidUserStatus
	}

	// 设置话题基本信息
	topic.UserID = userID
	topic.Status = "active"
	if topic.ExpiresAt.IsZero() {
		topic.ExpiresAt = time.Now().Add(24 * time.Hour) // 默认24小时后过期
	}

	// 创建话题
	if err := s.topicRepo.Create(ctx, topic); err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	// 处理图片
	if len(images) > 0 {
		topicImages := make([]*model.TopicImage, 0, len(images))
		for i, img := range images {
			// 上传图片
			fileURL, err := s.storage.UploadFile(ctx, img.File, storage.TopicDirectory)
			if err != nil {
				logger.Error("failed to upload topic image",
					logger.Any("error", err),
					logger.Uint64("topic_id", topic.ID),
					logger.Int("image_index", i))
				continue
			}

			// 创建图片记录
			topicImage := &model.TopicImage{
				TopicID:     topic.ID,
				ImageURL:    fileURL,
				SortOrder:   uint(i),
				ImageWidth:  uint(img.Width),  // 将int转换为uint
				ImageHeight: uint(img.Height), // 将int转换为uint
				FileSize:    img.Size,
			}
			topicImages = append(topicImages, topicImage)
		}

		// 保存图片记录
		if err := s.topicRepo.AddImages(ctx, topic.ID, topicImages); err != nil {
			logger.Error("failed to save topic images",
				logger.Any("error", err),
				logger.Uint64("topic_id", topic.ID))
		}
	}

	// 缓存话题信息
	cacheKey := cache.TopicKey(topic.ID)
	if err := cache.Set(cacheKey, topic, cache.DefaultExpiration); err != nil {
		logger.Warn("failed to cache topic", logger.Any("error", err))
	}

	return topic, nil
}

// UpdateTopic 更新话题
func (s *TopicService) UpdateTopic(ctx context.Context, userID uint64, topic *model.Topic) error {
	// 获取原话题信息
	existingTopic, err := s.GetTopicByID(ctx, topic.ID)
	if err != nil {
		return err
	}
	if existingTopic == nil {
		return ErrTopicNotFound
	}

	// 验证权限
	if existingTopic.UserID != userID {
		return ErrForbidden
	}

	// 检查话题状态
	if existingTopic.Status != "active" {
		return ErrInvalidTopicStatus
	}

	// 只更新允许修改的字段
	existingTopic.Title = topic.Title
	existingTopic.Content = topic.Content
	existingTopic.ExpiresAt = topic.ExpiresAt

	// 保存更新
	if err := s.topicRepo.Update(ctx, existingTopic); err != nil {
		return fmt.Errorf("failed to update topic: %w", err)
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topic.ID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete topic cache", logger.Any("error", err))
	}

	return nil
}

// DeleteTopic 删除话题
func (s *TopicService) DeleteTopic(ctx context.Context, userID, topicID uint64) error {
	// 获取话题信息
	topic, err := s.GetTopicByID(ctx, topicID)
	if err != nil {
		return err
	}
	if topic == nil {
		return ErrTopicNotFound
	}

	// 验证权限
	if topic.UserID != userID {
		return ErrForbidden
	}

	// 删除话题
	if err := s.topicRepo.Delete(ctx, topicID); err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topicID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete topic cache", logger.Any("error", err))
	}

	return nil
}

// GetTopicByID 获取话题详情
func (s *TopicService) GetTopicByID(ctx context.Context, topicID uint64) (*model.Topic, error) {
	// 尝试从缓存获取
	cacheKey := cache.TopicKey(topicID)
	var cachedTopic model.Topic
	err := cache.Get(cacheKey, &cachedTopic)
	if err == nil {
		return &cachedTopic, nil
	}

	// 从数据库获取
	topic, err := s.topicRepo.GetByID(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic: %w", err)
	}
	if topic == nil {
		return nil, nil
	}

	// 缓存话题信息
	if err := cache.Set(cacheKey, topic, cache.DefaultExpiration); err != nil {
		logger.Warn("failed to cache topic", logger.Any("error", err))
	}

	return topic, nil
}

// ViewTopic 查看话题（增加浏览次数）
func (s *TopicService) ViewTopic(ctx context.Context, topicID uint64) error {
	// 增加浏览次数
	if err := s.topicRepo.IncrementViewCount(ctx, topicID); err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topicID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete topic cache", logger.Any("error", err))
	}

	return nil
}

// ListTopics 获取话题列表
func (s *TopicService) ListTopics(ctx context.Context, page, pageSize int) ([]*model.Topic, int64, error) {
	offset := (page - 1) * pageSize
	return s.topicRepo.List(ctx, offset, pageSize)
}

// ListUserTopics 获取用户的话题列表
func (s *TopicService) ListUserTopics(ctx context.Context, userID uint64, page, pageSize int) ([]*model.Topic, int64, error) {
	offset := (page - 1) * pageSize
	return s.topicRepo.ListByUser(ctx, userID, offset, pageSize)
}

// GetNearbyTopics 获取附近的话题
func (s *TopicService) GetNearbyTopics(ctx context.Context, lat, lng float64, radius float64, page, pageSize int) ([]*model.Topic, int64, error) {
	offset := (page - 1) * pageSize
	return s.topicRepo.GetNearbyTopics(ctx, lat, lng, radius, offset, pageSize)
}

// AddInteraction 添加话题互动（点赞、收藏、分享）
func (s *TopicService) AddInteraction(ctx context.Context, userID, topicID uint64, interactionType string) error {
	// 检查话题是否存在
	topic, err := s.GetTopicByID(ctx, topicID)
	if err != nil {
		return err
	}
	if topic == nil {
		return ErrTopicNotFound
	}

	// 验证互动类型
	if !isValidInteractionType(interactionType) {
		return ErrInvalidInteraction
	}

	// 创建互动记录
	interaction := &model.TopicInteraction{
		TopicID:           topicID,
		UserID:            userID,
		InteractionType:   interactionType,
		InteractionStatus: "active",
	}

	// 保存互动
	if err := s.topicRepo.AddInteraction(ctx, interaction); err != nil {
		return fmt.Errorf("failed to add interaction: %w", err)
	}

	return nil
}

// RemoveInteraction 移除话题互动
func (s *TopicService) RemoveInteraction(ctx context.Context, userID, topicID uint64, interactionType string) error {
	return s.topicRepo.RemoveInteraction(ctx, topicID, userID, interactionType)
}

// GetInteractions 获取话题互动列表
func (s *TopicService) GetInteractions(ctx context.Context, topicID uint64, interactionType string) ([]*model.TopicInteraction, error) {
	return s.topicRepo.GetInteractions(ctx, topicID, interactionType)
}

// CleanExpiredTopics 清理过期话题
func (s *TopicService) CleanExpiredTopics(ctx context.Context) error {
	// 这个方法可以定期调用，用于清理过期的话题
	// 实现略...
	return nil
}

// 辅助函数

// isValidInteractionType 检查互动类型是否有效
func isValidInteractionType(interactionType string) bool {
	validTypes := map[string]bool{
		"like":     true,
		"favorite": true,
		"share":    true,
	}
	return validTypes[interactionType]
}

// AddTags 添加话题标签
func (s *TopicService) AddTags(ctx context.Context, topicID uint64, tags []string) error {
	// 验证标签数量
	if len(tags) > constants.MaxTagsPerTopic {
		return ErrTooManyTags
	}

	// 验证标签名称
	for _, tag := range tags {
		if len(tag) < constants.MinTagLength || len(tag) > constants.MaxTagLength {
			return ErrInvalidTagName
		}
	}

	// 批量创建或获取标签ID
	tagIDs, err := s.topicRepo.BatchCreate(ctx, tags)
	if err != nil {
		return fmt.Errorf("failed to create tags: %w", err)
	}

	// 添加话题-标签关联
	if err := s.topicRepo.AddTags(ctx, topicID, tagIDs); err != nil {
		return fmt.Errorf("failed to add topic tags: %w", err)
	}

	return nil
}

// RemoveTags 移除话题标签
func (s *TopicService) RemoveTags(ctx context.Context, topicID uint64, tagIDs []uint64) error {
	return s.topicRepo.RemoveTags(ctx, topicID, tagIDs)
}

// GetTopicTags 获取话题标签
func (s *TopicService) GetTopicTags(ctx context.Context, topicID uint64) ([]*model.Tag, error) {
	return s.topicRepo.GetTags(ctx, topicID)
}

// GetPopularTags 获取热门标签
func (s *TopicService) GetPopularTags(ctx context.Context, limit int) ([]*model.Tag, error) {
	// 尝试从缓存获取
	var tags []*model.Tag
	cacheKey := cache.PopularTagsKey()
	err := cache.Get(cacheKey, &tags)
	if err == nil {
		return tags, nil
	}

	// 从数据库获取
	tags, err = s.topicRepo.ListPopular(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 设置缓存
	if err := cache.Set(cacheKey, tags, constants.TagCacheExpiration); err != nil {
		logger.Warn("failed to cache popular tags", logger.Any("error", err))
	}

	return tags, nil
}

// AddTopicImage 添加话题图片
func (s *TopicService) AddTopicImage(ctx context.Context, topicID uint64, images []*model.File) error {
	if len(images) == 0 {
		return nil
	}

	// 检查话题是否存在
	topic, err := s.GetTopicByID(ctx, topicID)
	if err != nil {
		return err
	}
	if topic == nil {
		return ErrTopicNotFound
	}

	topicImages := make([]*model.TopicImage, 0, len(images))
	for i, img := range images {
		// 上传图片
		fileURL, err := s.storage.UploadFile(ctx, img.File, storage.TopicDirectory)
		if err != nil {
			logger.Error("failed to upload topic image",
				logger.Any("error", err),
				logger.Uint64("topic_id", topicID),
				logger.Int("image_index", i))
			continue
		}

		// 创建图片记录，注意类型转换
		topicImage := &model.TopicImage{
			TopicID:     topicID,
			ImageURL:    fileURL,
			SortOrder:   uint(i),
			ImageWidth:  uint(img.Width),  // 将int转换为uint
			ImageHeight: uint(img.Height), // 将int转换为uint
			FileSize:    img.Size,
		}
		topicImages = append(topicImages, topicImage)
	}

	// 保存图片记录
	if err := s.topicRepo.AddImages(ctx, topicID, topicImages); err != nil {
		return fmt.Errorf("failed to save topic images: %w", err)
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topicID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete topic cache",
			logger.Any("error", err),
			logger.Uint64("topic_id", topicID))
	}

	return nil
}

// validateImage 验证图片
func (s *TopicService) validateImage(image *model.File) error {
	// 验证文件类型
	if image.Type != "image" {
		return ErrFileTypeNotSupported
	}

	// 验证文件大小
	if image.Size > uint(storage.MaxFileSize) {
		return ErrFileTooLarge
	}

	// 验证图片尺寸
	if image.Width > storage.MaxImageDimension || image.Height > storage.MaxImageDimension {
		return ErrImageDimensionsTooLarge
	}

	return nil
}
