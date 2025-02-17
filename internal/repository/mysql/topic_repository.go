package mysql

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

const (
	RecentTopicCacheThreshold = 1000
)

type topicRepository struct {
	db                 *gorm.DB
	postedTopicCounter atomic.Int64
	recentTopicCache   repository.TopicCache
}

// NewTopicRepository 创建话题仓储实例
func NewTopicRepository(db *gorm.DB) repository.TopicRepository {
	rtc, err := repository.NewRedisRecentTopic(context.Background())
	if err != nil {
		panic(fmt.Sprintf("create recentTopicCache failed: %v\n", err))
	}
	tr := &topicRepository{db: db, recentTopicCache: rtc}

	return tr
}

// ListByTag 获取带有特定标签的话题列表
func (r *topicRepository) ListByTag(ctx context.Context, tagUID string, offset, limit int) ([]*model.Topic, int64, error) {
	var topics []*model.Topic
	var total int64

	db := r.db.WithContext(ctx).
		Joins("JOIN topic_tags ON topic_tags.topic_uid = topics.uid").
		Where("topic_tags.tag_uid = ? AND topics.status = ?", tagUID, "active")

	if err := db.Model(&model.Topic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("User").
		Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("topics.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

// Update 更新话题
func (r *topicRepository) Update(ctx context.Context, topic *model.Topic) error {
	return r.db.WithContext(ctx).Save(topic).Error
}

// Delete 删除话题
//
//	func (r *topicRepository) Delete(ctx context.Context, uid string) error {
//		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
//			// 删除话题相关的所有数据
//			if err := tx.Where("topic_uid = ?", uid).Delete(&model.TopicImage{}).Error; err != nil {
//				return err
//			}
//			if err := tx.Where("topic_uid = ?", uid).Delete(&model.TopicTag{}).Error; err != nil {
//				return err
//			}
//			if err := tx.Where("topic_uid = ?", uid).Delete(&model.TopicInteraction{}).Error; err != nil {
//				return err
//			}
//			// 删除话题
//			return tx.Where("uid = ?", uid).Delete(&model.Topic{}).Error
//		})
//	}
//
// Delete 删除话题（软删除）
func (r *topicRepository) Delete(ctx context.Context, uid string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新话题状态为closed
		if err := tx.Model(&model.Topic{}).
			Where("uid = ?", uid).
			Update("status", "closed").Error; err != nil {
			logger.Error("Failed to update topic status",
				logger.String("topic_uid", uid),
				logger.Any("error", err))
			return err
		}

		logger.Info("Successfully closed topic",
			logger.String("topic_uid", uid))
		return nil
	})
}

// // GetByUID 根据UID获取话题
//
//	func (r *topicRepository) GetByUID(ctx context.Context, uid string) (*model.Topic, error) {
//		var topic model.Topic
//		err := r.db.WithContext(ctx).
//			Preload("User").
//			Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
//				return db.Order("sort_order ASC")
//			}).
//			Preload("Tags"). // 简化预加载
//			Where("uid = ?", uid).
//			First(&topic).Error
//		// 添加日志检查标签数量
//		logger.Info("Retrieved topic with tags",
//			logger.String("topic_uid", uid),
//			logger.Int("tag_count", len(topic.Tags)))
//		if err != nil {
//			if err == gorm.ErrRecordNotFound {
//				return nil, nil
//			}
//			return nil, err
//		}
//		return &topic, nil
//	}
//
// GetByUID 根据UID获取话题
func (r *topicRepository) GetByUID(ctx context.Context, uid string) (*model.Topic, error) {
	var topic model.Topic

	logger.Info("Getting topic by UID in repository", logger.String("uid", uid))

	// 先获取话题基本信息和关联数据
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("uid = ?", uid).
		First(&topic).Error

	if err != nil {
		logger.Error("Failed to get topic basic info",
			logger.String("uid", uid),
			logger.Any("error", err))
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// 单独查询标签
	var tags []*model.Tag
	if err := r.db.WithContext(ctx).
		Table("tags").
		Select("tags.*").
		Joins("INNER JOIN topic_tags ON topic_tags.tag_uid = tags.uid").
		Where("topic_tags.topic_uid = ?", uid).
		Scan(&tags).Error; err != nil {
		logger.Error("Failed to get tags",
			logger.String("uid", uid),
			logger.Any("error", err))
	} else {
		logger.Info("Found tags",
			logger.String("uid", uid),
			logger.Int("tag_count", len(tags)))
		topic.Tags = tags
	}

	// 验证数据
	var tagCount int64
	r.db.WithContext(ctx).
		Table("topic_tags").
		Where("topic_uid = ?", uid).
		Count(&tagCount)

	logger.Info("Topic tags count in database",
		logger.String("uid", uid),
		logger.Int64("tag_count", tagCount),
		logger.Int("tags_in_topic", len(topic.Tags)))

	return &topic, nil
}

// List 获取话题列表
func (r *topicRepository) List(ctx context.Context, offset, limit int) ([]*model.Topic, int64, error) {
	var topics []*model.Topic
	var total int64

	db := r.db.WithContext(ctx).Where("status = ?", "active")

	if err := db.Model(&model.Topic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("User").
		Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Tags"). // 添加这行来预加载标签
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

// ListByUser 获取用户的话题列表
func (r *topicRepository) ListByUser(ctx context.Context, userUID string, offset, limit int) ([]*model.Topic, int64, error) {
	var topics []*model.Topic
	var total int64

	db := r.db.WithContext(ctx).Where("user_uid = ? AND status = ?", userUID, "active")

	if err := db.Model(&model.Topic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("User").
		Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Tags"). // 添加这行来预加载标签
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

// GetNearbyTopics 获取附近的话题
func (r *topicRepository) GetNearbyTopics(ctx context.Context, lat, lng float64, radius float64, offset, limit int) ([]*model.Topic, int64, error) {
	var topics []*model.Topic
	var total int64

	// 使用 MySQL 空间函数计算距离
	distanceSQL := "ST_Distance_Sphere(POINT(location_longitude, location_latitude), POINT(?, ?))"
	db := r.db.WithContext(ctx).
		Where(fmt.Sprintf("%s <= ?", distanceSQL), lng, lat, radius).
		Where("status = ?", "active")

	if err := db.Model(&model.Topic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("User").
		Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Tags"). // 添加这行来预加载标签
		Select("*, "+distanceSQL+" as distance", lng, lat).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&topics).Error

	if err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

// DeleteTopicImages 删除话题的照片
func (r *topicRepository) DeleteTopicImages(ctx context.Context, topicUID string, imageUIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查图片是否属于该话题
		var count int64
		if err := tx.Model(&model.TopicImage{}).
			Where("topic_uid = ? AND uid IN ?", topicUID, imageUIDs).
			Count(&count).Error; err != nil {
			return err
		}

		if int(count) != len(imageUIDs) {
			return fmt.Errorf("some images do not belong to this topic")
		}

		// 删除图片记录
		if err := tx.Where("topic_uid = ? AND uid IN ?", topicUID, imageUIDs).
			Delete(&model.TopicImage{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// AddImages 添加话题图片
func (r *topicRepository) AddImages(ctx context.Context, topicUID string, images []*model.TopicImage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, img := range images {
			img.TopicUID = topicUID
			if img.UID == "" { // 确保有 UID
				img.UID = uuid.New().String()
			}
			if err := tx.Create(img).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetImages 获取话题图片
func (r *topicRepository) GetImages(ctx context.Context, topicUID string) ([]*model.TopicImage, error) {
	var images []*model.TopicImage
	err := r.db.WithContext(ctx).
		Where("topic_uid = ?", topicUID).
		Order("sort_order ASC").
		Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

// AddTags 添加话题标签
func (r *topicRepository) AddTags(ctx context.Context, topicUID string, tagUIDs []string) error {
	logger.Info("Starting AddTags",
		logger.String("topic_uid", topicUID),
		logger.Any("tag_uids", tagUIDs))

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tagUID := range tagUIDs {
			// 首先验证标签是否存在
			var tag model.Tag
			if err := tx.Where("uid = ?", tagUID).First(&tag).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Error("Tag not found in database",
						logger.String("tag_uid", tagUID))
					return fmt.Errorf("tag not found: %s", tagUID)
				}
				return err
			}

			// 创建话题标签关联
			topicTag := model.TopicTag{
				TopicUID:  topicUID,
				TagUID:    tag.UID,
				CreatedAt: time.Now(),
			}

			if err := tx.Create(&topicTag).Error; err != nil {
				logger.Error("Failed to create topic tag",
					logger.String("topic_uid", topicUID),
					logger.String("tag_uid", tag.UID),
					logger.Any("error", err))
				return err
			}

			// 增加标签使用次数
			if err := tx.Model(&model.Tag{}).Where("uid = ?", tag.UID).
				UpdateColumn("use_count", gorm.Expr("use_count + ?", 1)).Error; err != nil {
				logger.Error("Failed to update tag use count",
					logger.String("tag_uid", tag.UID),
					logger.Any("error", err))
				return err
			}

			logger.Info("Successfully added tag and updated use count",
				logger.String("topic_uid", topicUID),
				logger.String("tag_uid", tag.UID))
		}
		return nil
	})
}

// RemoveTags 移除话题标签
func (r *topicRepository) RemoveTags(ctx context.Context, topicUID string, tagUIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("topic_uid = ? AND tag_uid IN ?", topicUID, tagUIDs).
			Delete(&model.TopicTag{}).Error; err != nil {
			return err
		}
		// 减少标签使用次数
		for _, tagUID := range tagUIDs {
			if err := tx.Model(&model.Tag{}).Where("uid = ?", tagUID).
				UpdateColumn("use_count", gorm.Expr("use_count - ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetTags 获取话题标签
func (r *topicRepository) GetTags(ctx context.Context, topicUID string) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.WithContext(ctx).
		Joins("JOIN topic_tags ON topic_tags.tag_uid = tags.uid").
		Where("topic_tags.topic_uid = ?", topicUID).
		Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// BatchCreate 批量创建标签
func (r *topicRepository) BatchCreate(ctx context.Context, tags []string) ([]string, error) {
	var tagUIDs []string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tagName := range tags {
			var tag model.Tag
			// 首先尝试查找现有标签
			err := tx.Where("name = ?", tagName).First(&tag).Error
			if err == gorm.ErrRecordNotFound {
				// 如果标签不存在，创建新标签
				tag = model.Tag{
					BaseModel: model.BaseModel{
						UID: uuid.New().String(),
					},
					Name:     tagName,
					UseCount: 0, // 初始使用次数为 0
				}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			tagUIDs = append(tagUIDs, tag.UID)
		}
		return nil
	})

	return tagUIDs, err
}

func (r *topicRepository) CreateNewTopic(ctx context.Context, userUID string, req *request.CreateTopicRequest) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//create an empty topic to lock the topic id
		if result := tx.Model(&model.Topic{}).Where("uid = ?", req.Uid).
			FirstOrCreate(&topic, model.Topic{
				BaseModel:         model.BaseModel{UID: req.Uid},
				UserUID:           userUID,
				Title:             req.Title,
				Content:           req.Content,
				LocationLatitude:  req.Latitude,
				LocationLongitude: req.Longitude,
				ExpiresAt:         req.ExpiresAt}); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			//topic exists
			return fmt.Errorf("topic id %s exists", req.Uid)
		}

		//create tags
		model_tags := []*model.Tag{}
		topic_tag := []model.TopicTag{}
		for _, tagName := range req.Tags {
			var t model.Tag
			result := tx.Where("name = ?", tagName).First(&t)
			fmt.Printf("result: %+v\n", result)
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				return result.Error
			}
			if result.RowsAffected == 1 {
				if err := tx.Model(&t).Update("use_count", t.UseCount+1).Error; err != nil {
					return err
				}
			} else {
				t = model.Tag{
					Name:     tagName,
					UseCount: 1,
				}
				if err := tx.Create(&t).Error; err != nil {
					return err
				}

			}

			model_tags = append(model_tags, &t)
			topic_tag = append(topic_tag, model.TopicTag{TopicUID: req.Uid, TagUID: t.UID})
		}

		//insert images
		model_images := []model.TopicImage{}
		for _, imageUrl := range req.Images {
			model_images = append(model_images, model.TopicImage{TopicUID: topic.UID, ImageURL: imageUrl})
		}
		if len(model_images) > 0 {
			if result := tx.Create(&model_images); result.Error != nil {
				return result.Error
			} else if result.RowsAffected != int64(len(req.Images)) {
				return fmt.Errorf("insert topic image failed")
			}
		}

		//update tag-topic join table
		// this needs to be done manually since the table is created manually
		if len(topic_tag) > 0 {
			if result := tx.Create(&topic_tag); result.Error != nil {
				fmt.Printf("crearte topic tag failed: %v\n", result.Error)
				return result.Error
			} else if result.RowsAffected != int64(len(topic_tag)) {
				return fmt.Errorf("update topic-tag failed")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	} else {
		updated := r.postedTopicCounter.Add(1)
		if updated == RecentTopicCacheThreshold {
			r.postedTopicCounter.Store(int64(0))
			go r.recentTopicCache.ReLoad(ctx)
		}
		return &topic, nil
	}

}

func (r *topicRepository) FindTopicsBy(c *gin.Context, by request.FindTopicsByRequest) ([]*model.CachedTopic, int, error) {
	var updatedScore int
	switch by.FindBy {
	case request.FindTopicsByRecency:

		topics, score, err := r.findTopicsByRecency(c, by.Max, by.RecencyScore)
		updatedScore = score
		if err == nil {
			return topics, updatedScore, nil
		} else if err.Error() == repository.CheckDBError {
			goto CheckDB
		} else {
			return nil, updatedScore, err
		}
	case request.FindTopicsByPopularity:
		return r.findTopicsByPopularity(c, by.Max)
	default:
		return nil, updatedScore, errors.New(fmt.Sprintf("unimplemented find topics by %v", by))
	}
CheckDB:
	return nil, updatedScore, nil
}

func (r *topicRepository) findTopicsByRecency(c *gin.Context, count int, before int) ([]*model.CachedTopic, int, error) {
	return r.recentTopicCache.Get(c, count, before)

}

func (r *topicRepository) findTopicsByPopularity(c *gin.Context, count int) ([]*model.CachedTopic, int, error) {
	return nil, 0, nil
}

// ListPopular 获取热门标签
func (r *topicRepository) ListPopular(ctx context.Context, limit int) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.WithContext(ctx).
		Order("use_count DESC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

// // AddInteraction 添加话题互动
//
//	func (r *topicRepository) AddInteraction(ctx context.Context, interaction *model.TopicInteraction) error {
//		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
//			if err := tx.Create(interaction).Error; err != nil {
//				return err
//			}
//			// 更新计数
//			return r.UpdateCounts(ctx, interaction.TopicUID)
//		})
//	}
func (r *topicRepository) AddInteraction(ctx context.Context, interaction *model.TopicInteraction) error {
	// 首先在事务外检查是否存在交互
	var existingInteraction model.TopicInteraction
	err := r.db.WithContext(ctx).
		Where("topic_uid = ? AND user_uid = ? AND interaction_type = ?",
			interaction.TopicUID, interaction.UserUID, interaction.InteractionType).
		First(&existingInteraction).Error

	// 开始事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err == gorm.ErrRecordNotFound {
			// 不存在则创建
			if err := tx.Create(interaction).Error; err != nil {
				return fmt.Errorf("failed to create interaction: %v", err)
			}
		} else if err == nil {
			// 存在则更新状态
			existingInteraction.InteractionStatus = interaction.InteractionStatus
			if err := tx.Save(&existingInteraction).Error; err != nil {
				return fmt.Errorf("failed to update interaction: %v", err)
			}
		} else {
			return fmt.Errorf("failed to check interaction: %v", err)
		}

		// 使用单个 SQL 更新点赞数
		updateSQL := `
            UPDATE topics t 
            SET likes_count = (
                SELECT COUNT(*) 
                FROM topic_interactions ti 
                WHERE ti.topic_uid = ? 
                AND ti.interaction_type = 'like' 
                AND ti.interaction_status = 'active'
            )
            WHERE t.uid = ?
        `
		if err := tx.Exec(updateSQL, interaction.TopicUID, interaction.TopicUID).Error; err != nil {
			return fmt.Errorf("failed to update like count: %v", err)
		}

		return nil
	})
}

// RemoveInteraction 移除话题互动
func (r *topicRepository) RemoveInteraction(ctx context.Context, topicUID, userUID string, interactionType string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("topic_uid = ? AND user_uid = ? AND interaction_type = ?",
			topicUID, userUID, interactionType).
			Delete(&model.TopicInteraction{}).Error; err != nil {
			return err
		}
		// 更新计数
		return r.UpdateCounts(ctx, topicUID)
	})
}

// GetInteractions 获取话题互动
func (r *topicRepository) GetInteractions(ctx context.Context, topicUID string, interactionType string) ([]*model.TopicInteraction, error) {
	var interactions []*model.TopicInteraction
	err := r.db.WithContext(ctx).
		Where("topic_uid = ? AND interaction_type = ? AND interaction_status = ?",
			topicUID, interactionType, "active").
		Preload("User").
		Find(&interactions).Error
	if err != nil {
		return nil, err
	}
	return interactions, nil
}

// IncrementViewCount 增加话题浏览次数
func (r *topicRepository) IncrementViewCount(ctx context.Context, topicUID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Topic{}).
		Where("uid = ?", topicUID).
		UpdateColumn("views_count", gorm.Expr("views_count + ?", 1)).
		Error
}

// UpdateCounts 更新话题的各种计数
func (r *topicRepository) UpdateCounts(ctx context.Context, topicUID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新点赞数
		var likesCount int64
		if err := tx.Model(&model.TopicInteraction{}).
			Where("topic_uid = ? AND interaction_type = ? AND interaction_status = ?",
				topicUID, "like", "active").
			Count(&likesCount).Error; err != nil {
			return err
		}

		// 更新分享数
		var sharesCount int64
		if err := tx.Model(&model.TopicInteraction{}).
			Where("topic_uid = ? AND interaction_type = ? AND interaction_status = ?",
				topicUID, "share", "active").
			Count(&sharesCount).Error; err != nil {
			return err
		}

		// 更新参与人数（去重的互动用户数）
		var participantsCount int64
		if err := tx.Model(&model.TopicInteraction{}).
			Where("topic_uid = ? AND interaction_status = ?", topicUID, "active").
			Distinct("user_uid").
			Count(&participantsCount).Error; err != nil {
			return err
		}

		// 更新话题统计数据
		return tx.Model(&model.Topic{}).
			Where("uid = ?", topicUID).
			Updates(map[string]interface{}{
				"likes_count":        likesCount,
				"shares_count":       sharesCount,
				"participants_count": participantsCount,
			}).Error
	})
}
