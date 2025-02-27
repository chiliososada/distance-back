package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/chiliososada/distance-back/internal/api/request"
	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/logger"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户信息
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, uid string) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, "uid = ?", uid).Error
}

// GetByUID 根据UID获取用户
func (r *userRepository) GetByUID(ctx context.Context, uid string) (*model.User, error) {
	logger.Info("Getting user by UID", logger.String("uid", uid))

	var user model.User
	result := r.db.WithContext(ctx).Preload("Chat", func(db *gorm.DB) *gorm.DB {
		return db.Select("chat_room_uid,user_uid")
	}).
		// Debug(). // 添加这行来看SQL日志
		Where("uid = ?", uid).
		First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Info("User not found", logger.String("uid", uid))
			return nil, nil
		}
		logger.Error("Failed to get user",
			logger.String("uid", uid),
			logger.Any("error", result.Error))
		return nil, result.Error
	}

	logger.Info("Found user", logger.Any("user", user))
	return &user, nil
}

// GetByFirebaseUID 根据Firebase UID获取用户
func (r *userRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	var auth model.UserAuthentication
	if err := r.db.WithContext(ctx).Where("firebase_uid = ?", firebaseUID).First(&auth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var user model.User
	if err := r.db.WithContext(ctx).Where("uid = ?", auth.UserUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	db := r.db.WithContext(ctx)
	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Search 搜索用户
func (r *userRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	db := r.db.WithContext(ctx).Where(
		"nickname LIKE ? OR bio LIKE ?",
		fmt.Sprintf("%%%s%%", keyword),
		fmt.Sprintf("%%%s%%", keyword),
	)

	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetNearbyUsers 获取附近的用户
func (r *userRepository) GetNearbyUsers(ctx context.Context, lat, lng float64, radius float64, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// 使用 MySQL 空间函数计算距离
	distanceSQL := "ST_Distance_Sphere(POINT(location_longitude, location_latitude), POINT(?, ?))"
	db := r.db.WithContext(ctx).
		Where(fmt.Sprintf("%s <= ?", distanceSQL), lng, lat, radius).
		Where("location_sharing = ?", true)

	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Select("*, "+distanceSQL+" as distance", lng, lat).
		Order("distance").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateStatus 更新用户状态
func (r *userRepository) UpdateStatus(ctx context.Context, userUID string, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uid = ?", userUID).
		Update("status", status).
		Error
}

// UpdateLastActive 更新用户最后活跃时间
func (r *userRepository) UpdateLastActive(ctx context.Context, userUID string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("uid = ?", userUID).
		Update("last_active_at", time.Now()).
		Error
}

// CreateAuthentication 创建用户认证信息
func (r *userRepository) CreateAuthentication(ctx context.Context, auth *model.UserAuthentication) error {
	return r.db.WithContext(ctx).Create(auth).Error
}

// UpdateAuthentication 更新用户认证信息
func (r *userRepository) UpdateAuthentication(ctx context.Context, auth *model.UserAuthentication) error {
	return r.db.WithContext(ctx).Save(auth).Error
}

// CreateDevice 创建用户设备
func (r *userRepository) CreateDevice(ctx context.Context, device *model.UserDevice) error {
	return r.db.WithContext(ctx).Create(device).Error
}

// UpdateDevice 更新用户设备
func (r *userRepository) UpdateDevice(ctx context.Context, device *model.UserDevice) error {
	return r.db.WithContext(ctx).Save(device).Error
}

// GetDeviceByToken 根据设备令牌获取设备信息
func (r *userRepository) GetDeviceByToken(ctx context.Context, token string) (*model.UserDevice, error) {
	var device model.UserDevice
	if err := r.db.WithContext(ctx).Where("device_token = ?", token).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

// GetUserDevices 获取用户的所有设备
func (r *userRepository) GetUserDevices(ctx context.Context, userUID string) ([]*model.UserDevice, error) {
	var devices []*model.UserDevice
	if err := r.db.WithContext(ctx).Where("user_uid = ?", userUID).Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}
func (r *userRepository) CreateWithAuth(ctx context.Context, user *model.User, auth *model.UserAuthentication) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		auth.UserUID = user.UID
		if err := tx.Create(auth).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *userRepository) UpdateWithAuth(ctx context.Context, user *model.User, auth *model.UserAuthentication) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		if err := tx.Where("user_uid = ?", user.UID).Updates(auth).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *userRepository) RegisterOrUpdateUser(ctx context.Context, uid string, req *request.UpdateProfileRequest) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//first ensure the user exists
		if result := tx.Where("uid = ?", uid).FirstOrCreate(&user,
			model.User{BaseModel: model.BaseModel{UID: uid}}); result.Error != nil {
			return result.Error
		} else {
			user.UpdateFromRequest(req)
			fmt.Printf("update user: %+v\n", user)
			if err := tx.Model(&user).Updates(user).Error; err != nil {
				return err
			}

			return nil
		}
	})
	if err != nil {
		return nil, err
	} else {
		return &user, nil
	}
}
