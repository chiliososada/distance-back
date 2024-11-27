package mysql

import (
	"DistanceBack_v1/internal/model"
	"DistanceBack_v1/internal/repository"
	"context"

	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) repository.TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) Update(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id uint64) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(ctx).First(&tag, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

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

func (r *tagRepository) ListPopular(ctx context.Context, limit int) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.WithContext(ctx).
		Order("use_count DESC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

func (r *tagRepository) IncrementUseCount(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", 1)).
		Error
}

func (r *tagRepository) DecrementUseCount(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count - ?", 1)).
		Error
}

func (r *tagRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Tag{}, id).Error
}

func (r *tagRepository) BatchCreate(ctx context.Context, tags []string) ([]uint64, error) {
	var tagIds []uint64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tagName := range tags {
			var tag model.Tag
			err := tx.Where("name = ?", tagName).First(&tag).Error
			if err == gorm.ErrRecordNotFound {
				tag = model.Tag{
					Name: tagName,
				}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			tagIds = append(tagIds, tag.ID)
		}
		return nil
	})
	return tagIds, err
}
