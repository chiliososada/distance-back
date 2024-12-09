package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/auth"
	"github.com/chiliososada/distance-back/pkg/cache"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/chiliososada/distance-back/pkg/storage"
	"github.com/chiliososada/distance-back/pkg/utils"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo repository.UserRepository
	storage  storage.Storage
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, storage storage.Storage) *UserService {
	return &UserService{
		userRepo: userRepo,
		storage:  storage,
	}
}

// RegisterOrUpdateUser 注册或更新用户信息（Firebase认证后）
func (s *UserService) RegisterOrUpdateUser(ctx context.Context, firebaseUser *auth.AuthUser) (*model.User, error) {
	// 查找现有用户
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUser.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by firebase uid: %w", err)
	}

	if user == nil {
		// 创建新用户
		user = &model.User{
			BaseModel: model.BaseModel{
				UID: uuid.NewString(), // 添加 uuid 导入
			},
			Nickname:            firebaseUser.DisplayName,
			AvatarURL:           firebaseUser.PhotoURL,
			Gender:              "other",
			Status:              model.UserStatusActive,
			PrivacyLevel:        model.PrivacyPublic,
			NotificationEnabled: true,
			LocationSharing:     true,
			PhotoEnabled:        true,
			UserType:            "individual",
		}

		// 准备认证信息
		auth := &model.UserAuthentication{
			BaseModel: model.BaseModel{
				UID: uuid.NewString(),
			},
			UserUID:      user.UID,
			FirebaseUID:  firebaseUser.UID,
			Email:        utils.NewNullString(firebaseUser.Email),
			PhoneNumber:  utils.NewNullString(firebaseUser.PhoneNumber),
			LastSignInAt: utils.TimePtr(time.Now()),
			AuthProvider: "password",
		}

		// 使用事务创建用户和认证信息
		if err := s.userRepo.CreateWithAuth(ctx, user, auth); err != nil {
			return nil, fmt.Errorf("failed to create user and authentication: %w", err)
		}
	} else {
		// 更新现有用户信息
		if firebaseUser.DisplayName != "" {
			user.Nickname = firebaseUser.DisplayName
		}
		if firebaseUser.PhotoURL != "" {
			user.AvatarURL = firebaseUser.PhotoURL
		}

		// 准备认证信息更新
		auth := &model.UserAuthentication{
			UserUID:      user.UID,
			FirebaseUID:  firebaseUser.UID,
			Email:        utils.NewNullString(firebaseUser.Email),
			PhoneNumber:  utils.NewNullString(firebaseUser.PhoneNumber),
			LastSignInAt: utils.TimePtr(time.Now()),
			AuthProvider: "password",
		}

		// 使用事务更新用户和认证信息
		if err := s.userRepo.UpdateWithAuth(ctx, user, auth); err != nil {
			return nil, fmt.Errorf("failed to update user and authentication: %w", err)
		}
	}

	// 缓存用户信息
	cacheKey := cache.UserKey(user.UID)
	if err := cache.Set(cacheKey, user, cache.DefaultExpiration); err != nil {
		logger.Warn("failed to cache user info", logger.Any("error", err))
	}

	return user, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userUID string, profile *model.User) error {
	// 获取现有用户信息
	user, err := s.GetUserByUID(ctx, userUID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 更新可修改的字段
	user.Nickname = profile.Nickname
	user.BirthDate = profile.BirthDate
	user.Gender = profile.Gender
	user.Bio = profile.Bio
	user.Language = profile.Language
	user.PrivacyLevel = profile.PrivacyLevel
	user.NotificationEnabled = profile.NotificationEnabled
	user.LocationSharing = profile.LocationSharing
	user.PhotoEnabled = profile.PhotoEnabled

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	// 清除缓存
	cacheKey := cache.UserKey(userUID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete user cache", logger.Any("error", err))
	}

	return nil
}

// UpdateAvatar 更新用户头像
func (s *UserService) UpdateAvatar(ctx context.Context, userUID string, avatar *model.File) error {
	// 获取现有用户信息
	user, err := s.GetUserByUID(ctx, userUID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 上传新头像
	fileURL, err := s.storage.UploadFile(ctx, avatar.File, storage.AvatarDirectory)
	if err != nil {
		return fmt.Errorf("failed to upload avatar: %w", err)
	}

	// 更新用户头像URL
	user.AvatarURL = fileURL
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
	}

	// 清除缓存
	cacheKey := cache.UserKey(userUID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete user cache", logger.Any("error", err))
	}

	return nil
}

// UpdateLocation 更新用户位置
func (s *UserService) UpdateLocation(ctx context.Context, userUID string, lat, lng float64) error {
	// 获取现有用户信息
	user, err := s.GetUserByUID(ctx, userUID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 更新位置信息
	user.LocationLatitude = lat
	user.LocationLongitude = lng
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user location: %w", err)
	}

	// 更新位置缓存
	locationKey := cache.LocationKey(userUID)
	location := &utils.Location{
		Latitude:  lat,
		Longitude: lng,
	}
	if err := cache.Set(locationKey, location, cache.ShortExpiration); err != nil {
		logger.Warn("failed to cache user location", logger.Any("error", err))
	}

	return nil
}

// 在 UserService 中添加这个方法
func (s *UserService) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	// 通过 Firebase UID 查找用户
	user, err := s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by firebase uid: %w", err)
	}

	// 用户不存在时返回 nil
	if user == nil {
		return nil, nil
	}

	return user, nil
}

// GetUserByUID 获取用户信息
// 在 service/user_service.go 中
func (s *UserService) GetUserByUID(ctx context.Context, userUID string) (*model.User, error) {
	logger.Info("Getting user by UID in service", logger.String("uid", userUID))

	// 先从缓存获取
	cacheKey := cache.UserKey(userUID)
	var cachedUser model.User
	err := cache.Get(cacheKey, &cachedUser)
	if err == nil && cachedUser.UID != "" {
		logger.Info("Got user from cache", logger.Any("user", cachedUser))
		return &cachedUser, nil
	}

	// 从数据库获取
	user, err := s.userRepo.GetByUID(ctx, userUID)
	if err != nil {
		logger.Error("Failed to get user from db",
			logger.String("uid", userUID),
			logger.Any("error", err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil || user.UID == "" {
		logger.Error("User not found in db", logger.String("uid", userUID))
		return nil, errors.New("user not found")
	}

	logger.Info("Got user from db", logger.Any("user", user))

	// 缓存用户信息
	if err := cache.Set(cacheKey, user, cache.DefaultExpiration); err != nil {
		logger.Warn("Failed to cache user info", logger.Any("error", err))
	}

	return user, nil
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(ctx context.Context, keyword string, page, pageSize int) ([]*model.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.Search(ctx, keyword, offset, pageSize)
}

// GetNearbyUsers 获取附近的用户
func (s *UserService) GetNearbyUsers(ctx context.Context, lat, lng float64, radius float64, page, pageSize int) ([]*model.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.GetNearbyUsers(ctx, lat, lng, radius, offset, pageSize)
}

// RegisterDevice 注册用户设备
func (s *UserService) RegisterDevice(ctx context.Context, userUID string, device *model.UserDevice) error {
	// 检查设备是否已存在
	existingDevice, err := s.userRepo.GetDeviceByToken(ctx, device.DeviceToken)
	if err != nil {
		return fmt.Errorf("failed to check device: %w", err)
	}

	if existingDevice != nil {
		// 更新现有设备
		existingDevice.UserUID = userUID
		existingDevice.DeviceName = device.DeviceName
		existingDevice.DeviceModel = device.DeviceModel
		existingDevice.OSVersion = device.OSVersion
		existingDevice.AppVersion = device.AppVersion
		existingDevice.PushEnabled = device.PushEnabled
		existingDevice.LastActiveAt = time.Now()
		existingDevice.IsActive = true

		if err := s.userRepo.UpdateDevice(ctx, existingDevice); err != nil {
			return fmt.Errorf("failed to update device: %w", err)
		}
	} else {
		// 创建新设备
		device.UserUID = userUID
		device.LastActiveAt = time.Now()
		device.IsActive = true

		if err := s.userRepo.CreateDevice(ctx, device); err != nil {
			return fmt.Errorf("failed to create device: %w", err)
		}
	}

	return nil
}

// UpdateLastActive 更新用户最后活跃时间
func (s *UserService) UpdateLastActive(ctx context.Context, userUID string) error {
	return s.userRepo.UpdateLastActive(ctx, userUID)
}

// VerifyDevice 验证设备信息
func (s *UserService) VerifyDevice(ctx context.Context, userUID string, deviceToken string) (bool, error) {
	// 获取设备信息
	device, err := s.userRepo.GetDeviceByToken(ctx, deviceToken)
	if err != nil {
		return false, fmt.Errorf("failed to get device: %w", err)
	}

	// 如果设备不存在则返回false
	if device == nil {
		return false, nil
	}

	// 验证设备是否属于该用户且处于活跃状态
	return device.UserUID == userUID && device.IsActive, nil
}

// ListUsers 获取用户列表（管理员接口）
func (s *UserService) ListUsers(ctx context.Context, page, size int) ([]*model.User, int64, error) {
	offset := (page - 1) * size

	// 从数据库获取用户列表
	users, total, err := s.userRepo.List(ctx, offset, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// 批量更新缓存
	for _, user := range users {
		cacheKey := cache.UserKey(user.UID)
		if err := cache.Set(cacheKey, user, cache.DefaultExpiration); err != nil {
			logger.Warn("failed to cache user info",
				logger.String("user_uid", user.UID),
				logger.Any("error", err))
		}
	}

	return users, total, nil
}
