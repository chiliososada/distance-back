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
		err := tx.Where("follower_id = ? AND following_id = ?",
			relationship.FollowerID, relationship.FollowingID).
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
func (r *relationshipRepository) Delete(ctx context.Context, followerID, followingID uint64) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&model.UserRelationship{}).Error
}

// GetRelationship 获取两个用户之间的关系
func (r *relationshipRepository) GetRelationship(ctx context.Context, followerID, followingID uint64) (*model.UserRelationship, error) {
	var relationship model.UserRelationship
	err := r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
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
func (r *relationshipRepository) GetFollowers(ctx context.Context, userID uint64, status string, offset, limit int) ([]*model.UserRelationship, int64, error) {
	var relationships []*model.UserRelationship
	var total int64

	db := r.db.WithContext(ctx).
		Where("following_id = ?", userID)

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
func (r *relationshipRepository) GetFollowings(ctx context.Context, userID uint64, status string, offset, limit int) ([]*model.UserRelationship, int64, error) {
	var relationships []*model.UserRelationship
	var total int64

	db := r.db.WithContext(ctx).
		Where("follower_id = ?", userID)

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
func (r *relationshipRepository) UpdateStatus(ctx context.Context, followerID, followingID uint64, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "accepted" {
		updates["accepted_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.UserRelationship{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Updates(updates).Error
}

// ExistsRelationship 检查关系是否存在
func (r *relationshipRepository) ExistsRelationship(ctx context.Context, followerID, followingID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserRelationship{}).
		Where("follower_id = ? AND following_id = ? AND status = ?",
			followerID, followingID, "accepted").
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// // GetMutualFollowers 获取共同关注者（好友）
// func (r *relationshipRepository) GetMutualFollowers(ctx context.Context, userID1, userID2 uint64, offset, limit int) ([]*model.User, int64, error) {
// 	var users []*model.User
// 	var total int64

// 	// 子查询：找到互相关注的用户
// 	subQuery := r.db.Model(&model.UserRelationship{}).
// 		Select("r1.follower_id").
// 		Table("user_relationships as r1").
// 		Joins("JOIN user_relationships as r2 ON r1.follower_id = r2.following_id AND r1.following_id = r2.follower_id").
// 		Where("r1.following_id = ? AND r2.following_id = ? AND r1.status = ? AND r2.status = ?",
// 			userID1, userID2, "accepted", "accepted")

// 	// 获取总数
// 	if err := r.db.Model(&model.User{}).
// 		Where("id IN (?)", subQuery).
// 		Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	// 获取用户列表
// 	err := r.db.Where("id IN (?)", subQuery).
// 		Offset(offset).
// 		Limit(limit).
// 		Find(&users).Error

// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	return users, total, nil
// }

// // GetRecommendedFollows 获取推荐关注的用户
// func (r *relationshipRepository) GetRecommendedFollows(ctx context.Context, userID uint64, limit int) ([]*model.User, error) {
// 	var users []*model.User

// 	// 子查询：获取用户已关注的人的ID列表
// 	followingSubQuery := r.db.Model(&model.UserRelationship{}).
// 		Select("following_id").
// 		Where("follower_id = ? AND status = ?", userID, "accepted")

// 	// 子查询：获取已关注的用户的关注者
// 	recommendedQuery := r.db.Model(&model.UserRelationship{}).
// 		Select("follower_id").
// 		Where("following_id IN (?) AND follower_id != ? AND status = ?", followingSubQuery, userID, "accepted").
// 		Group("follower_id").
// 		Order("count(*) DESC")

// 	// 获取推荐用户列表
// 	err := r.db.Where("id IN (?) AND id NOT IN (?)", recommendedQuery, followingSubQuery).
// 		Limit(limit).
// 		Find(&users).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	return users, nil
// }

// // BlockUser 拉黑用户
// func (r *relationshipRepository) BlockUser(ctx context.Context, blockerID, blockedID uint64) error {
// 	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		// 创建或更新拉黑关系
// 		relationship := &model.UserRelationship{
// 			FollowerID:  blockerID,
// 			FollowingID: blockedID,
// 			Status:      "blocked",
// 		}

// 		// 删除现有的关注关系（如果存在）
// 		if err := tx.Where("(follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)",
// 			blockerID, blockedID, blockedID, blockerID).
// 			Delete(&model.UserRelationship{}).Error; err != nil {
// 			return err
// 		}

// 		// 创建拉黑关系
// 		return tx.Create(relationship).Error
// 	})
// }

// // UnblockUser 取消拉黑用户
// func (r *relationshipRepository) UnblockUser(ctx context.Context, blockerID, blockedID uint64) error {
// 	return r.db.WithContext(ctx).
// 		Where("follower_id = ? AND following_id = ? AND status = ?",
// 			blockerID, blockedID, "blocked").
// 		Delete(&model.UserRelationship{}).Error
// }

// // GetBlockedUsers 获取拉黑的用户列表
// func (r *relationshipRepository) GetBlockedUsers(ctx context.Context, userID uint64, offset, limit int) ([]*model.UserRelationship, int64, error) {
// 	var relationships []*model.UserRelationship
// 	var total int64

// 	db := r.db.WithContext(ctx).
// 		Where("follower_id = ? AND status = ?", userID, "blocked")

// 	// 获取总数
// 	if err := db.Model(&model.UserRelationship{}).Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	// 获取拉黑列表
// 	err := db.Preload("Following").
// 		Order("created_at DESC").
// 		Offset(offset).
// 		Limit(limit).
// 		Find(&relationships).Error

// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	return relationships, total, nil
// }
