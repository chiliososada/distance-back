package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"DistanceBack_v1/internal/model"
	"DistanceBack_v1/internal/repository"
	"DistanceBack_v1/pkg/auth"
	"DistanceBack_v1/pkg/cache"
	"DistanceBack_v1/pkg/logger"
	"DistanceBack_v1/pkg/storage"
	"DistanceBack_v1/pkg/utils"
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
			Nickname:            firebaseUser.DisplayName,
			AvatarURL:           firebaseUser.PhotoURL,
			Status:              model.UserStatusActive,
			PrivacyLevel:        model.PrivacyPublic,
			NotificationEnabled: true,
			LocationSharing:     true,
			PhotoEnabled:        true,
			UserType:            "individual",
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// 创建认证信息
		auth := &model.UserAuthentication{
			UserID:       user.ID,
			FirebaseUID:  firebaseUser.UID,
			Email:        utils.NewNullString(firebaseUser.Email),
			PhoneNumber:  utils.NewNullString(firebaseUser.PhoneNumber),
			LastSignInAt: utils.TimePtr(time.Now()), // 转换为 *time.Time
		}

		if err := s.userRepo.CreateAuthentication(ctx, auth); err != nil {
			return nil, fmt.Errorf("failed to create user authentication: %w", err)
		}
	} else {
		// 更新现有用户信息
		if firebaseUser.DisplayName != "" {
			user.Nickname = firebaseUser.DisplayName
		}
		if firebaseUser.PhotoURL != "" {
			user.AvatarURL = firebaseUser.PhotoURL
		}

		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		// 更新认证信息
		auth := &model.UserAuthentication{
			UserID:       user.ID,
			FirebaseUID:  firebaseUser.UID,
			Email:        utils.NewNullString(firebaseUser.Email),
			PhoneNumber:  utils.NewNullString(firebaseUser.PhoneNumber),
			LastSignInAt: utils.TimePtr(time.Now()), // 转换为 *time.Time
		}

		if err := s.userRepo.UpdateAuthentication(ctx, auth); err != nil {
			return nil, fmt.Errorf("failed to update user authentication: %w", err)
		}
	}

	// 缓存用户信息
	cacheKey := cache.UserKey(user.ID)
	if err := cache.Set(cacheKey, user, cache.DefaultExpiration); err != nil {
		logger.Warn("failed to cache user info", logger.Any("error", err))
	}

	return user, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID uint64, profile *model.User) error {
	// 获取现有用户信息
	user, err := s.GetUserByID(ctx, userID)
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
	cacheKey := cache.UserKey(userID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete user cache", logger.Any("error", err))
	}

	return nil
}

// UpdateAvatar 更新用户头像
func (s *UserService) UpdateAvatar(ctx context.Context, userID uint64, avatar *model.File) error {
	// 获取现有用户信息
	user, err := s.GetUserByID(ctx, userID)
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
	cacheKey := cache.UserKey(userID)
	if err := cache.Delete(cacheKey); err != nil {
		logger.Warn("failed to delete user cache", logger.Any("error", err))
	}

	return nil
}

// UpdateLocation 更新用户位置
func (s *UserService) UpdateLocation(ctx context.Context, userID uint64, lat, lng float64) error {
	// 获取现有用户信息
	user, err := s.GetUserByID(ctx, userID)
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
	locationKey := cache.LocationKey(userID)
	location := &utils.Location{
		Latitude:  lat,
		Longitude: lng,
	}
	if err := cache.Set(locationKey, location, cache.ShortExpiration); err != nil {
		logger.Warn("failed to cache user location", logger.Any("error", err))
	}

	return nil
}

// GetUserByID 获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID uint64) (*model.User, error) {
	// 尝试从缓存获取
	cacheKey := cache.UserKey(userID)
	var cachedUser model.User
	err := cache.Get(cacheKey, &cachedUser)
	if err == nil {
		return &cachedUser, nil
	}

	// 从数据库获取
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	// 缓存用户信息
	if err := cache.Set(cacheKey, user, cache.DefaultExpiration); err != nil {
		logger.Warn("failed to cache user info", logger.Any("error", err))
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
func (s *UserService) RegisterDevice(ctx context.Context, userID uint64, device *model.UserDevice) error {
	// 检查设备是否已存在
	existingDevice, err := s.userRepo.GetDeviceByToken(ctx, device.DeviceToken)
	if err != nil {
		return fmt.Errorf("failed to check device: %w", err)
	}

	if existingDevice != nil {
		// 更新现有设备
		existingDevice.UserID = userID
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
		device.UserID = userID
		device.LastActiveAt = time.Now()
		device.IsActive = true

		if err := s.userRepo.CreateDevice(ctx, device); err != nil {
			return fmt.Errorf("failed to create device: %w", err)
		}
	}

	return nil
}

// UpdateLastActive 更新用户最后活跃时间
func (s *UserService) UpdateLastActive(ctx context.Context, userID uint64) error {
	return s.userRepo.UpdateLastActive(ctx, userID)
}
