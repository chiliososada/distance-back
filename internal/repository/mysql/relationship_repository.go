package mysql

import (
	"context"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"

	"gorm.io/gorm"
)

type relationshipRepository struct {
	db *gorm.DB
}

// NewRelationshipRepository 创建关系仓储实例
func NewRelationshipRepository(db *gorm.DB) repository.RelationshipRepository {
	return &relationshipRepository{db: db}
}

// Create 创建关系
func (r *relationshipRepository) Create(ctx context.Context, relationship *model.UserRelationship) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查是否已存在关系
		var exists model.UserRelationship
		err := tx.Where("follower_uid = ? AND following_uid = ?",
			relationship.FollowerUID, relationship.FollowingUID).
			First(&exists).Error

		if err == nil {
			// 关系已存在，更新状态
			exists.Status = relationship.Status
			if relationship.Status == "accepted" {
				exists.AcceptedAt = &time.Time{}
				*exists.AcceptedAt = time.Now()
			}
			return tx.Save(&exists).Error
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// 创建新关系
		if relationship.Status == "accepted" {
			now := time.Now()
			relationship.AcceptedAt = &now
		}
		return tx.Create(relationship).Error
	})
}

// Update 更新关系
func (r *relationshipRepository) Update(ctx context.Context, relationship *model.UserRelationship) error {
	if relationship.Status == "accepted" && relationship.AcceptedAt == nil {
		now := time.Now()
		relationship.AcceptedAt = &now
	}
	return r.db.WithContext(ctx).Save(relationship).Error
}

// Delete 删除关系
func (r *relationshipRepository) Delete(ctx context.Context, followerUID, followingUID string) error {
	return r.db.WithContext(ctx).
		Where("follower_uid = ? AND following_uid = ?", followerUID, followingUID).
		Delete(&model.UserRelationship{}).Error
}

// GetRelationship 获取两个用户之间的关系
func (r *relationshipRepository) GetRelationship(ctx context.Context, followerUID, followingUID string) (*model.UserRelationship, error) {
	var relationship model.UserRelationship
	err := r.db.WithContext(ctx).
		Where("follower_uid = ? AND following_uid = ?", followerUID, followingUID).
		First(&relationship).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &relationship, nil
}

// GetFollowers 获取用户的粉丝列表
func (r *relationshipRepository) GetFollowers(ctx context.Context, userUID string, status string, offset, limit int) ([]*model.UserRelationship, int64, error) {
	var relationships []*model.UserRelationship
	var total int64

	db := r.db.WithContext(ctx).
		Where("following_uid = ?", userUID)

	if status != "" {
		db = db.Where("status = ?", status)
	}

	// 获取总数
	if err := db.Model(&model.UserRelationship{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取关系列表
	err := db.Preload("Follower"). // 预加载关注者信息
					Order("created_at DESC").
					Offset(offset).
					Limit(limit).
					Find(&relationships).Error

	if err != nil {
		return nil, 0, err
	}

	return relationships, total, nil
}

// GetFollowings 获取用户关注的列表
func (r *relationshipRepository) GetFollowings(ctx context.Context, userUID string, status string, offset, limit int) ([]*model.UserRelationship, int64, error) {
	var relationships []*model.UserRelationship
	var total int64

	db := r.db.WithContext(ctx).
		Where("follower_uid = ?", userUID)

	if status != "" {
		db = db.Where("status = ?", status)
	}

	// 获取总数
	if err := db.Model(&model.UserRelationship{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取关系列表
	err := db.Preload("Following"). // 预加载被关注者信息
					Order("created_at DESC").
					Offset(offset).
					Limit(limit).
					Find(&relationships).Error

	if err != nil {
		return nil, 0, err
	}

	return relationships, total, nil
}

// UpdateStatus 更新关系状态
func (r *relationshipRepository) UpdateStatus(ctx context.Context, followerUID, followingUID string, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "accepted" {
		updates["accepted_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.UserRelationship{}).
		Where("follower_uid = ? AND following_uid = ?", followerUID, followingUID).
		Updates(updates).Error
}

// ExistsRelationship 检查关系是否存在
func (r *relationshipRepository) ExistsRelationship(ctx context.Context, followerUID, followingUID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserRelationship{}).
		Where("follower_uid = ? AND following_uid = ? AND status = ?",
			followerUID, followingUID, "accepted").
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
