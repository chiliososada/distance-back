package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/logger"
)

type RelationshipService struct {
	relationRepo repository.RelationshipRepository
	userRepo     repository.UserRepository
	chatService  *ChatService
}

// NewRelationshipService 创建关系服务实例
func NewRelationshipService(
	relationRepo repository.RelationshipRepository,
	userRepo repository.UserRepository,
	chatService *ChatService,
) *RelationshipService {
	return &RelationshipService{
		relationRepo: relationRepo,
		userRepo:     userRepo,
		chatService:  chatService,
	}
}

// Follow 关注用户
func (s *RelationshipService) Follow(ctx context.Context, followerID, followingID uint64) error {
	// 检查是否自关注
	if followerID == followingID {
		return ErrSelfRelation
	}

	// 验证用户是否存在
	follower, err := s.userRepo.GetByID(ctx, followerID)
	if err != nil || follower == nil {
		return ErrUserNotFound
	}
	following, err := s.userRepo.GetByID(ctx, followingID)
	if err != nil || following == nil {
		return ErrUserNotFound
	}

	// 检查是否被对方拉黑
	isBlocked, err := s.IsBlocked(ctx, followingID, followerID)
	if err != nil {
		return err
	}
	if isBlocked {
		return ErrBlockedUser
	}

	// 检查目标用户的隐私设置
	var status string
	if following.PrivacyLevel == "public" {
		status = "accepted"
	} else {
		status = "pending"
	}

	// 创建关注关系
	relationship := &model.UserRelationship{
		FollowerID:  followerID,
		FollowingID: followingID,
		Status:      status,
	}

	if err := s.relationRepo.Create(ctx, relationship); err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	// 如果是直接接受的关注，需要处理互相关注（好友）的情况
	if status == "accepted" {
		s.handleMutualFollow(ctx, followerID, followingID)
	}

	return nil
}

// Unfollow 取消关注
func (s *RelationshipService) Unfollow(ctx context.Context, followerID, followingID uint64) error {
	if followerID == followingID {
		return ErrSelfRelation
	}

	if err := s.relationRepo.Delete(ctx, followerID, followingID); err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	return nil
}

// AcceptFollow 接受关注请求
func (s *RelationshipService) AcceptFollow(ctx context.Context, userID, followerID uint64) error {
	// 获取关注请求
	relationship, err := s.relationRepo.GetRelationship(ctx, followerID, userID)
	if err != nil {
		return err
	}
	if relationship == nil {
		return ErrNotFound
	}
	if relationship.Status != "pending" {
		return ErrInvalidRelationType
	}

	// 更新关系状态
	now := time.Now()
	relationship.Status = "accepted"
	relationship.AcceptedAt = &now

	if err := s.relationRepo.Update(ctx, relationship); err != nil {
		return fmt.Errorf("failed to update relationship: %w", err)
	}

	// 处理互相关注的情况
	s.handleMutualFollow(ctx, followerID, userID)

	return nil
}

// RejectFollow 拒绝关注请求
func (s *RelationshipService) RejectFollow(ctx context.Context, userID, followerID uint64) error {
	return s.relationRepo.Delete(ctx, followerID, userID)
}

// GetFollowers 获取粉丝列表
func (s *RelationshipService) GetFollowers(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*model.UserRelationship, int64, error) {
	return s.relationRepo.GetFollowers(ctx, userID, status, (page-1)*pageSize, pageSize)
}

// GetFollowings 获取关注列表
func (s *RelationshipService) GetFollowings(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*model.UserRelationship, int64, error) {
	return s.relationRepo.GetFollowings(ctx, userID, status, (page-1)*pageSize, pageSize)
}

// GetFriends 获取好友列表（互相关注）
func (s *RelationshipService) GetFriends(ctx context.Context, userID uint64, page, pageSize int) ([]*model.User, int64, error) {
	followings, _, err := s.relationRepo.GetFollowings(ctx, userID, "accepted", 0, 1000)
	if err != nil {
		return nil, 0, err
	}

	var friends []*model.User
	var total int64

	for _, following := range followings {
		// 检查是否互相关注
		isFollowed, err := s.relationRepo.ExistsRelationship(ctx, following.FollowingID, userID)
		if err != nil {
			continue
		}
		if isFollowed {
			user, err := s.userRepo.GetByID(ctx, following.FollowingID)
			if err != nil || user == nil {
				continue
			}
			friends = append(friends, user)
			total++
		}
	}

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(friends) {
		return []*model.User{}, total, nil
	}
	if end > len(friends) {
		end = len(friends)
	}

	return friends[start:end], total, nil
}

// IsFollowing 检查是否正在关注
func (s *RelationshipService) IsFollowing(ctx context.Context, followerID, followingID uint64) (bool, error) {
	relationship, err := s.relationRepo.GetRelationship(ctx, followerID, followingID)
	if err != nil {
		return false, err
	}
	return relationship != nil && relationship.Status == "accepted", nil
}

// IsFollowed 检查是否被关注
func (s *RelationshipService) IsFollowed(ctx context.Context, userID, followerID uint64) (bool, error) {
	relationship, err := s.relationRepo.GetRelationship(ctx, followerID, userID)
	if err != nil {
		return false, err
	}
	return relationship != nil && relationship.Status == "accepted", nil
}

// IsBlocked 检查是否被拉黑
func (s *RelationshipService) IsBlocked(ctx context.Context, blockerID, userID uint64) (bool, error) {
	relationship, err := s.relationRepo.GetRelationship(ctx, blockerID, userID)
	if err != nil {
		return false, err
	}
	return relationship != nil && relationship.Status == "blocked", nil
}

// IsFriend 检查是否是好友（互相关注）
func (s *RelationshipService) IsFriend(ctx context.Context, userID1, userID2 uint64) (bool, error) {
	isFollowing, err := s.IsFollowing(ctx, userID1, userID2)
	if err != nil {
		return false, err
	}
	if !isFollowing {
		return false, nil
	}

	return s.IsFollowing(ctx, userID2, userID1)
}

// 处理互相关注（好友）情况
func (s *RelationshipService) handleMutualFollow(ctx context.Context, userID1, userID2 uint64) {
	isFriend, err := s.IsFriend(ctx, userID1, userID2)
	if err != nil {
		logger.Error("failed to check friend status",
			logger.Any("error", err),
			logger.Uint64("user1", userID1),
			logger.Uint64("user2", userID2))
		return
	}

	if isFriend {
		// 创建私聊房间
		_, err := s.chatService.CreatePrivateRoom(ctx, userID1, userID2)
		if err != nil {
			logger.Error("failed to create private room",
				logger.Any("error", err),
				logger.Uint64("user1", userID1),
				logger.Uint64("user2", userID2))
		}
	}
}
