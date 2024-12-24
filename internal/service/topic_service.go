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
	"github.com/google/uuid"
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
func (s *TopicService) CreateTopic(ctx context.Context, userUID string, topic *model.Topic, images []*model.File, tags []string) (*model.Topic, error) {
	// 验证用户状态
	user, err := s.userRepo.GetByUID(ctx, userUID)
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
	topic.UserUID = userUID
	topic.Status = "active"
	if topic.ExpiresAt.IsZero() {
		topic.ExpiresAt = time.Now().Add(24 * time.Hour) // 默认24小时后过期
	}
	if topic.UID == "" {
		topic.UID = uuid.New().String()
	}

	// 创建话题
	if err := s.topicRepo.Create(ctx, topic); err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	// 处理图片
	if len(images) > 0 {
		topicImages := make([]*model.TopicImage, 0, len(images))
		for i, img := range images {
			fileURL, err := s.storage.UploadFile(ctx, img.File, storage.TopicDirectory)
			if err != nil {
				logger.Error("failed to upload topic image",
					logger.Any("error", err),
					logger.String("topic_uid", topic.UID),
					logger.Int("image_index", i))
				continue
			}

			topicImage := &model.TopicImage{
				TopicUID:    topic.UID,
				ImageURL:    fileURL,
				SortOrder:   uint(i),
				ImageWidth:  uint(img.Width),
				ImageHeight: uint(img.Height),
				FileSize:    img.Size,
			}
			topicImages = append(topicImages, topicImage)
		}

		logger.Info("Saving topic images",
			logger.String("topic_uid", topic.UID),
			logger.Int("image_count", len(topicImages)))

		if err := s.topicRepo.AddImages(ctx, topic.UID, topicImages); err != nil {
			logger.Error("failed to save topic images",
				logger.Any("error", err),
				logger.String("topic_uid", topic.UID))
			return nil, fmt.Errorf("failed to save topic images: %w", err)
		}
	}

	// 处理标签
	if len(tags) > 0 {
		logger.Info("Processing tags",
			logger.Any("tags", tags),
			logger.String("topic_uid", topic.UID))

		// 先创建或获取标签
		tagUIDs, err := s.topicRepo.BatchCreate(ctx, tags)
		if err != nil {
			logger.Error("Failed to create/get tags",
				logger.Any("error", err),
				logger.Any("tags", tags))
			return nil, fmt.Errorf("failed to process tags: %w", err)
		}

		// 添加标签关联
		if err := s.topicRepo.AddTags(ctx, topic.UID, tagUIDs); err != nil {
			logger.Error("Failed to add tags",
				logger.Any("error", err),
				logger.String("topic_uid", topic.UID),
				logger.Any("tag_uids", tagUIDs))
			return nil, fmt.Errorf("failed to add tags: %w", err)
		}

		logger.Info("Successfully added tags",
			logger.String("topic_uid", topic.UID),
			logger.Any("tags", tags),
			logger.Any("tag_uids", tagUIDs))
	}

	// 重新获取完整信息
	updatedTopic, err := s.topicRepo.GetByUID(ctx, topic.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated topic: %w", err)
	}

	// 缓存话题信息
	cacheKey := cache.TopicKey(topic.UID)
	if err := cache.Set(cacheKey, updatedTopic, cache.DefaultExpiration); err != nil {
		logger.Warn("failed to cache topic", logger.Any("error", err))
	}

	logger.Info("Retrieved updated topic",
		logger.String("topic_uid", topic.UID),
		logger.Int("image_count", len(updatedTopic.TopicImages)),
		logger.Int("tag_count", len(updatedTopic.Tags)))

	return updatedTopic, nil
}

// UpdateTopic 更新话题
func (s *TopicService) UpdateTopic(
	ctx context.Context,
	userUID string,
	topic *model.Topic,
	newImages []*model.File,
	removeImageUIDs []string,
	tags []string,
) error {
	// 获取原话题信息
	existingTopic, err := s.GetTopicByUID(ctx, topic.UID)
	if err != nil {
		return fmt.Errorf("failed to get topic: %w", err)
	}
	if existingTopic == nil {

		return fmt.Errorf("failed to get existingTopic: %w", err)
	}

	// 验证权限
	if existingTopic.UserUID != userUID {
		return fmt.Errorf("TopicNotFound: %w", err)
	}

	// 检查话题状态
	if existingTopic.Status != model.TopicStatusActive {
		return fmt.Errorf("TopicNotActive: %w", err)
	}

	// 更新基本信息
	existingTopic.Title = topic.Title
	existingTopic.Content = topic.Content
	existingTopic.ExpiresAt = topic.ExpiresAt

	// 保存基本信息更新
	if err := s.topicRepo.Update(ctx, existingTopic); err != nil {
		return fmt.Errorf("failed to update topic: %w", err)
	}

	// 处理要删除的图片
	if len(removeImageUIDs) > 0 {
		logger.Info("Deleting images",
			logger.Any("remove_image_uids", removeImageUIDs))
		// 获取要删除的图片信息
		images, err := s.topicRepo.GetImages(ctx, topic.UID)
		if err != nil {
			return fmt.Errorf("failed to get images: %w", err)
		}

		// 找出需要删除的图片URL
		var imageURLsToDelete []string
		for _, image := range images {
			for _, uidToRemove := range removeImageUIDs {
				if image.UID == uidToRemove {
					imageURLsToDelete = append(imageURLsToDelete, image.ImageURL)
				}
			}
		}

		// 删除存储服务中的图片
		for _, url := range imageURLsToDelete {
			if err := storage.GetStorage().DeleteFile(ctx, url); err != nil {
				logger.Error("Failed to delete image file",
					logger.String("url", url),
					logger.Any("error", err))
			}
		}

		// 删除数据库中的图片关联
		if err := s.topicRepo.DeleteTopicImages(ctx, topic.UID, removeImageUIDs); err != nil {
			return fmt.Errorf("failed to remove images: %w", err)
		}
	}

	// 处理新图片
	if len(newImages) > 0 {
		// 上传图片并创建图片记录
		topicImages := make([]*model.TopicImage, len(newImages))
		for i, img := range newImages {
			// 构建存储路径
			directory := fmt.Sprintf("topics/%s", topic.UID)
			// 上传文件
			imageURL, err := storage.GetStorage().UploadFile(ctx, img.File, directory)
			if err != nil {
				return fmt.Errorf("failed to upload image: %w", err)
			}

			// 创建图片记录
			topicImages[i] = &model.TopicImage{
				TopicUID:    topic.UID,
				ImageURL:    imageURL,
				ImageWidth:  uint(img.Width),
				ImageHeight: uint(img.Height),
				FileSize:    img.Size,
				SortOrder:   uint(i),
			}
		}

		// 保存图片记录
		if err := s.topicRepo.AddImages(ctx, topic.UID, topicImages); err != nil {
			return fmt.Errorf("failed to add images: %w", err)
		}
	}

	// 处理标签更新
	if tags != nil {
		// 获取现有标签
		existingTags, err := s.topicRepo.GetTags(ctx, topic.UID)
		if err != nil {
			return fmt.Errorf("failed to get existing tags: %w", err)
		}

		// 删除现有标签
		existingTagUIDs := make([]string, len(existingTags))
		for i, tag := range existingTags {
			existingTagUIDs[i] = tag.UID
		}

		if err := s.topicRepo.RemoveTags(ctx, topic.UID, existingTagUIDs); err != nil {
			return fmt.Errorf("failed to remove tags: %w", err)
		}

		// 添加新标签
		if len(tags) > 0 {
			// 批量创建或获取标签
			tagUIDs, err := s.topicRepo.BatchCreate(ctx, tags)
			if err != nil {
				return fmt.Errorf("failed to create tags: %w", err)
			}

			// 建立标签关联
			if err := s.topicRepo.AddTags(ctx, topic.UID, tagUIDs); err != nil {
				return fmt.Errorf("failed to add tags: %w", err)
			}
		}
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topic.UID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("Failed to delete topic cache",
			logger.String("topic_uid", topic.UID),
			logger.Any("error", err))
	}

	return nil
}

// DeleteTopic 删除话题
func (s *TopicService) DeleteTopic(ctx context.Context, userUID string, topicUID string) error {
	// 获取话题信息
	topic, err := s.GetTopicByUID(ctx, topicUID)
	if err != nil {
		return err
	}
	if topic == nil {
		return ErrTopicNotFound
	}

	// 验证权限
	if topic.UserUID != userUID {
		return ErrForbidden
	}

	// 删除话题
	if err := s.topicRepo.Delete(ctx, topicUID); err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topicUID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete topic cache", logger.Any("error", err))
	}

	return nil
}

// GetTopicByUID 获取话题详情
func (s *TopicService) GetTopicByUID(ctx context.Context, topicUID string) (*model.Topic, error) {
	// 尝试从缓存获取
	// cacheKey := cache.TopicKey(topicUID)
	// var cachedTopic model.Topic
	// err := cache.Get(cacheKey, &cachedTopic)
	// if err == nil {
	// 	return &cachedTopic, nil
	// }

	// 从数据库获取
	topic, err := s.topicRepo.GetByUID(ctx, topicUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic: %w", err)
	}
	if topic == nil {
		return nil, nil
	}

	// 缓存话题信息
	// if err := cache.Set(cacheKey, topic, cache.DefaultExpiration); err != nil {
	// 	logger.Warn("failed to cache topic", logger.Any("error", err))
	// }

	return topic, nil
}

// ViewTopic 查看话题（增加浏览次数）
func (s *TopicService) ViewTopic(ctx context.Context, topicUID string) error {
	// 增加浏览次数
	if err := s.topicRepo.IncrementViewCount(ctx, topicUID); err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topicUID)
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
func (s *TopicService) ListUserTopics(ctx context.Context, userUID string, page, pageSize int) ([]*model.Topic, int64, error) {
	offset := (page - 1) * pageSize
	return s.topicRepo.ListByUser(ctx, userUID, offset, pageSize)
}

// GetNearbyTopics 获取附近的话题
func (s *TopicService) GetNearbyTopics(ctx context.Context, lat, lng float64, radius float64, page, pageSize int) ([]*model.Topic, int64, error) {
	offset := (page - 1) * pageSize
	return s.topicRepo.GetNearbyTopics(ctx, lat, lng, radius, offset, pageSize)
}

// AddInteraction 添加话题互动（点赞、收藏、分享）
func (s *TopicService) AddInteraction(ctx context.Context, userUID string, topicUID string, interactionType string) error {
	// 检查话题是否存在
	topic, err := s.GetTopicByUID(ctx, topicUID)
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
		TopicUID:          topicUID,
		UserUID:           userUID,
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
func (s *TopicService) RemoveInteraction(ctx context.Context, userUID string, topicUID string, interactionType string) error {
	return s.topicRepo.RemoveInteraction(ctx, topicUID, userUID, interactionType)
}

// GetInteractions 获取话题互动列表
func (s *TopicService) GetInteractions(ctx context.Context, topicUID string, interactionType string) ([]*model.TopicInteraction, error) {
	return s.topicRepo.GetInteractions(ctx, topicUID, interactionType)
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
func (s *TopicService) AddTags(ctx context.Context, topicUID string, tags []string) error {
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

	// 批量创建或获取标签UID
	tagUIDs, err := s.topicRepo.BatchCreate(ctx, tags)
	if err != nil {
		return fmt.Errorf("failed to create tags: %w", err)
	}

	// 添加话题-标签关联
	if err := s.topicRepo.AddTags(ctx, topicUID, tagUIDs); err != nil {
		return fmt.Errorf("failed to add topic tags: %w", err)
	}

	return nil
}

// RemoveTags 移除话题标签
func (s *TopicService) RemoveTags(ctx context.Context, topicUID string, tagUIDs []string) error {
	return s.topicRepo.RemoveTags(ctx, topicUID, tagUIDs)
}

// GetTopicTags 获取话题标签
func (s *TopicService) GetTopicTags(ctx context.Context, topicUID string) ([]*model.Tag, error) {
	return s.topicRepo.GetTags(ctx, topicUID)
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
func (s *TopicService) AddTopicImage(ctx context.Context, userUID string, topicUID string, images []*model.File) error {
	// 验证话题是否存在且属于该用户
	topic, err := s.GetTopicByUID(ctx, topicUID)
	if err != nil {
		return fmt.Errorf("failed to get topic: %w", err)
	}
	if topic == nil {
		return ErrTopicNotFound
	}
	if topic.UserUID != userUID {
		return ErrForbidden
	}

	// 检查话题状态
	if topic.Status != model.TopicStatusActive {
		return ErrInvalidTopicStatus
	}

	// 处理图片上传
	topicImages := make([]*model.TopicImage, 0, len(images))
	for i, img := range images {
		// 上传图片
		fileURL, err := s.storage.UploadFile(ctx, img.File, storage.TopicDirectory)
		if err != nil {
			logger.Error("failed to upload topic image",
				logger.Any("error", err),
				logger.String("topic_uid", topicUID),
				logger.Int("image_index", i))
			continue
		}

		// 创建图片记录
		topicImage := &model.TopicImage{
			TopicUID:    topicUID,
			ImageURL:    fileURL,
			SortOrder:   uint(i),
			ImageWidth:  uint(img.Width),
			ImageHeight: uint(img.Height),
			FileSize:    img.Size,
		}
		topicImages = append(topicImages, topicImage)
	}

	// 保存图片记录
	if err := s.topicRepo.AddImages(ctx, topicUID, topicImages); err != nil {
		return fmt.Errorf("failed to save topic images: %w", err)
	}

	// 清除缓存
	cacheKey := cache.TopicKey(topicUID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete topic cache", logger.Any("error", err))
	}

	return nil
}
