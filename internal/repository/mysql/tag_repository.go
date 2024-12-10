package mysql

import (
	"context"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) repository.TagRepository {
	return &tagRepository{db: db}
}

// Create 创建标签
func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// Update 更新标签
func (r *tagRepository) Update(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// GetByUID 根据UID获取标签
func (r *tagRepository) GetByUID(ctx context.Context, uid string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// GetByName 根据名称获取标签
func (r *tagRepository) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// List 获取标签列表
func (r *tagRepository) List(ctx context.Context, offset, limit int) ([]*model.Tag, int64, error) {
	var tags []*model.Tag
	var total int64

	err := r.db.WithContext(ctx).Model(&model.Tag{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&tags).Error
	if err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// ListPopular 获取热门标签列表
func (r *tagRepository) ListPopular(ctx context.Context, limit int) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.WithContext(ctx).
		Order("use_count DESC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

// IncrementUseCount 增加标签使用次数
func (r *tagRepository) IncrementUseCount(ctx context.Context, uid string) error {
	return r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("uid = ?", uid).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", 1)).
		Error
}

// DecrementUseCount 减少标签使用次数
func (r *tagRepository) DecrementUseCount(ctx context.Context, uid string) error {
	return r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("uid = ?", uid).
		UpdateColumn("use_count", gorm.Expr("use_count - ?", 1)).
		Error
}

// Delete 删除标签
func (r *tagRepository) Delete(ctx context.Context, uid string) error {
	return r.db.WithContext(ctx).Where("uid = ?", uid).Delete(&model.Tag{}).Error
}

// BatchCreate 批量创建标签
func (r *tagRepository) BatchCreate(ctx context.Context, tags []string) ([]string, error) {
	var tagUIDs []string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tagName := range tags {
			var tag model.Tag

			// 查找现有标签
			err := tx.Where("name = ?", tagName).First(&tag).Error
			if err == gorm.ErrRecordNotFound {
				newUID := uuid.New().String()
				tag = model.Tag{
					BaseModel: model.BaseModel{
						UID:       newUID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:     tagName,
					UseCount: 0,
				}

				// 打印调试信息
				logger.Info("Creating new tag with fields",
					logger.String("name", tagName),
					logger.String("uid", tag.UID))

				if err := tx.Create(&tag).Error; err != nil {
					logger.Error("Failed to create tag",
						logger.String("name", tagName),
						logger.Any("error", err))
					return err
				}
			} else {
				logger.Info("Found existing tag",
					logger.String("name", tagName),
					logger.String("uid", tag.UID))
			}

			// 确保我们有正确的 UID
			tagUIDs = append(tagUIDs, tag.UID)
		}
		return nil
	})

	logger.Info("Batch create tags result",
		logger.Any("tag_uids", tagUIDs),
		logger.Any("error", err))

	return tagUIDs, err
}
