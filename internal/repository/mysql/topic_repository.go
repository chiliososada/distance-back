package mysql

import (
	"context"
	"fmt"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"

	"gorm.io/gorm"
)

type topicRepository struct {
	db *gorm.DB
}

// NewTopicRepository 创建话题仓储实例
func NewTopicRepository(db *gorm.DB) repository.TopicRepository {
	return &topicRepository{db: db}
}

// Create 创建话题
func (r *topicRepository) Create(ctx context.Context, topic *model.Topic) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建话题
		if err := tx.Create(topic).Error; err != nil {
			return err
		}

		return nil
	})
}

// Update 更新话题
func (r *topicRepository) Update(ctx context.Context, topic *model.Topic) error {
	return r.db.WithContext(ctx).Save(topic).Error
}

// Delete 删除话题
func (r *topicRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除话题相关的所有数据
		if err := tx.Where("topic_id = ?", id).Delete(&model.TopicImage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("topic_id = ?", id).Delete(&model.TopicTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("topic_id = ?", id).Delete(&model.TopicInteraction{}).Error; err != nil {
			return err
		}
		// 删除话题
		return tx.Delete(&model.Topic{}, id).Error
	})
}

// GetByID 根据ID获取话题
func (r *topicRepository) GetByID(ctx context.Context, id uint64) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		First(&topic, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
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
func (r *topicRepository) ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*model.Topic, int64, error) {
	var topics []*model.Topic
	var total int64

	db := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, "active")

	if err := db.Model(&model.Topic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("User").
		Preload("TopicImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
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

// AddImages 添加话题图片
func (r *topicRepository) AddImages(ctx context.Context, topicID uint64, images []*model.TopicImage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, img := range images {
			img.TopicID = topicID
			if err := tx.Create(img).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetImages 获取话题图片
func (r *topicRepository) GetImages(ctx context.Context, topicID uint64) ([]*model.TopicImage, error) {
	var images []*model.TopicImage
	err := r.db.WithContext(ctx).
		Where("topic_id = ?", topicID).
		Order("sort_order ASC").
		Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

// AddTags 添加话题标签
func (r *topicRepository) AddTags(ctx context.Context, topicID uint64, tagIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tagID := range tagIDs {
			topicTag := &model.TopicTag{
				TopicID: topicID,
				TagID:   tagID,
			}
			if err := tx.Create(topicTag).Error; err != nil {
				return err
			}
			// 增加标签使用次数
			if err := tx.Model(&model.Tag{}).Where("id = ?", tagID).
				UpdateColumn("use_count", gorm.Expr("use_count + ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchCreate 添加多个标签
func (r *topicRepository) BatchCreate(ctx context.Context, tags []string) ([]uint64, error) {
	var tagIDs []uint64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tagName := range tags {
			var tag model.Tag
			// 首先尝试查找现有标签
			err := tx.Where("name = ?", tagName).First(&tag).Error
			if err == gorm.ErrRecordNotFound {
				// 如果标签不存在，创建新标签
				tag = model.Tag{
					Name:     tagName,
					UseCount: 1,
				}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				// 如果标签存在，增加使用次数
				if err := tx.Model(&tag).
					UpdateColumn("use_count", gorm.Expr("use_count + ?", 1)).Error; err != nil {
					return err
				}
			}
			tagIDs = append(tagIDs, tag.ID)
		}
		return nil
	})

	return tagIDs, err
}

// RemoveTags 移除话题标签
func (r *topicRepository) RemoveTags(ctx context.Context, topicID uint64, tagIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("topic_id = ? AND tag_id IN ?", topicID, tagIDs).
			Delete(&model.TopicTag{}).Error; err != nil {
			return err
		}
		// 减少标签使用次数
		for _, tagID := range tagIDs {
			if err := tx.Model(&model.Tag{}).Where("id = ?", tagID).
				UpdateColumn("use_count", gorm.Expr("use_count - ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetTags 获取话题标签
func (r *topicRepository) GetTags(ctx context.Context, topicID uint64) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.WithContext(ctx).
		Joins("JOIN topic_tags ON topic_tags.tag_id = tags.id").
		Where("topic_tags.topic_id = ?", topicID).
		Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// ListPopular 获取常用标签
func (r *topicRepository) ListPopular(ctx context.Context, limit int) ([]*model.Tag, error) {
	var tags []*model.Tag

	err := r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Order("use_count DESC").
		Limit(limit).
		Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

// ListByTag 获取带有特定标签的话题列表
func (r *topicRepository) ListByTag(ctx context.Context, tagID uint64, offset, limit int) ([]*model.Topic, int64, error) {
	var topics []*model.Topic
	var total int64

	db := r.db.WithContext(ctx).
		Joins("JOIN topic_tags ON topic_tags.topic_id = topics.id").
		Where("topic_tags.tag_id = ? AND topics.status = ?", tagID, "active")

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

// AddInteraction 添加话题互动
func (r *topicRepository) AddInteraction(ctx context.Context, interaction *model.TopicInteraction) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(interaction).Error; err != nil {
			return err
		}
		// 更新计数
		return r.UpdateCounts(ctx, interaction.TopicID)
	})
}

// RemoveInteraction 移除话题互动
func (r *topicRepository) RemoveInteraction(ctx context.Context, topicID, userID uint64, interactionType string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("topic_id = ? AND user_id = ? AND interaction_type = ?",
			topicID, userID, interactionType).
			Delete(&model.TopicInteraction{}).Error; err != nil {
			return err
		}
		// 更新计数
		return r.UpdateCounts(ctx, topicID)
	})
}

// GetInteractions 获取话题互动
func (r *topicRepository) GetInteractions(ctx context.Context, topicID uint64, interactionType string) ([]*model.TopicInteraction, error) {
	var interactions []*model.TopicInteraction
	err := r.db.WithContext(ctx).
		Where("topic_id = ? AND interaction_type = ? AND interaction_status = ?",
			topicID, interactionType, "active").
		Preload("User").
		Find(&interactions).Error
	if err != nil {
		return nil, err
	}
	return interactions, nil
}

// IncrementViewCount 增加话题浏览次数
func (r *topicRepository) IncrementViewCount(ctx context.Context, topicID uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.Topic{}).
		Where("id = ?", topicID).
		UpdateColumn("views_count", gorm.Expr("views_count + ?", 1)).
		Error
}

// UpdateCounts 更新话题的各种计数
func (r *topicRepository) UpdateCounts(ctx context.Context, topicID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新点赞数
		var likesCount int64
		if err := tx.Model(&model.TopicInteraction{}).
			Where("topic_id = ? AND interaction_type = ? AND interaction_status = ?",
				topicID, "like", "active").
			Count(&likesCount).Error; err != nil {
			return err
		}

		// 更新分享数
		var sharesCount int64
		if err := tx.Model(&model.TopicInteraction{}).
			Where("topic_id = ? AND interaction_type = ? AND interaction_status = ?",
				topicID, "share", "active").
			Count(&sharesCount).Error; err != nil {
			return err
		}

		// 更新参与人数（去重的互动用户数）
		var participantsCount int64
		if err := tx.Model(&model.TopicInteraction{}).
			Where("topic_id = ? AND interaction_status = ?", topicID, "active").
			Distinct("user_id").
			Count(&participantsCount).Error; err != nil {
			return err
		}

		// 更新话题统计数据
		return tx.Model(&model.Topic{}).
			Where("id = ?", topicID).
			Updates(map[string]interface{}{
				"likes_count":        likesCount,
				"shares_count":       sharesCount,
				"participants_count": participantsCount,
			}).Error
	})
}
