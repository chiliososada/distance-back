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
func (s *RelationshipService) Follow(ctx context.Context, followerUID, followingUID string) error {
	// 检查是否自关注
	if followerUID == followingUID {
		return ErrSelfRelation
	}

	// 验证用户是否存在
	follower, err := s.userRepo.GetByUID(ctx, followerUID)
	if err != nil || follower == nil {
		return ErrUserNotFound
	}
	following, err := s.userRepo.GetByUID(ctx, followingUID)
	if err != nil || following == nil {
		return ErrUserNotFound
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
		FollowerUID:  followerUID,
		FollowingUID: followingUID,
		Status:       status,
	}

	if err := s.relationRepo.Create(ctx, relationship); err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	// 如果是直接接受的关注，需要处理互相关注（好友）的情况
	if status == "accepted" {
		s.handleMutualFollow(ctx, followerUID, followingUID)
	}

	return nil
}

// Unfollow 取消关注(accepted_at 为空)
func (s *RelationshipService) Unfollow(ctx context.Context, followerUID, followingUID string) error {
	if followerUID == followingUID {
		return ErrSelfRelation
	}

	// 直接更新状态为rejected
	return s.relationRepo.UpdateStatus(ctx, followerUID, followingUID, "rejected")
}

// AcceptFollow 接受关注请求
func (s *RelationshipService) AcceptFollow(ctx context.Context, userUID, followerUID string) error {
	// 获取关注请求
	relationship, err := s.relationRepo.GetRelationship(ctx, followerUID, userUID)
	if err != nil {
		return err
	}
	if relationship == nil {
		return ErrNotFound
	}
	if relationship.Status != "pending" {
		return ErrInvalidRelationType
	}

	now := time.Now()
	relationship.Status = "accepted"
	relationship.AcceptedAt = &now

	// 调用repository层的AcceptFollow方法
	if err := s.relationRepo.AcceptFollow(ctx, relationship); err != nil {
		return fmt.Errorf("failed to accept follow: %w", err)
	}

	// 成功后处理互相关注的情况
	s.handleMutualFollow(ctx, followerUID, userUID)

	return nil
}

// RejectFollow 拒绝关注请求(accepted_at 非空)
func (s *RelationshipService) RejectFollow(ctx context.Context, userUID, followerUID string) error {
	return s.relationRepo.UpdateStatus(ctx, followerUID, userUID, "rejected")
}

// GetFollowers 获取粉丝列表
func (s *RelationshipService) GetFollowers(ctx context.Context, userUID string, status string, page, pageSize int) ([]*model.UserRelationship, int64, error) {
	return s.relationRepo.GetFollowers(ctx, userUID, status, (page-1)*pageSize, pageSize)
}

// GetFollowings 获取关注列表
func (s *RelationshipService) GetFollowings(ctx context.Context, userUID string, status string, page, pageSize int) ([]*model.UserRelationship, int64, error) {
	return s.relationRepo.GetFollowings(ctx, userUID, status, (page-1)*pageSize, pageSize)
}

// GetFriends 获取好友列表（互相关注）
func (s *RelationshipService) GetFriends(ctx context.Context, userUID string, page, pageSize int) ([]*model.User, int64, error) {
	followings, total, err := s.relationRepo.GetFollowings(ctx, userUID, "accepted", (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var friends []*model.User
	for _, following := range followings {
		// 检查是否互相关注
		isFollowed, err := s.IsFriend(ctx, following.FollowingUID, userUID)
		if err != nil {
			continue
		}
		if isFollowed {
			user, err := s.userRepo.GetByUID(ctx, following.FollowingUID)
			if err != nil || user == nil {
				continue
			}
			friends = append(friends, user)
		}
	}

	return friends, total, nil
}

// IsFollowing 检查是否正在关注
func (s *RelationshipService) IsFollowing(ctx context.Context, followerUID, followingUID string) (bool, error) {
	relationship, err := s.relationRepo.GetRelationship(ctx, followerUID, followingUID)
	if err != nil {
		return false, err
	}
	return relationship != nil && relationship.Status == "pending", nil
}

// IsFollowed 检查是否被关注
func (s *RelationshipService) IsFollowed(ctx context.Context, userUID, followerUID string) (bool, error) {
	relationship, err := s.relationRepo.GetRelationship(ctx, followerUID, userUID)
	if err != nil {
		return false, err
	}
	return relationship != nil && relationship.Status == "accepted", nil
}

// IsBlocked 检查是否被拉黑
func (s *RelationshipService) IsRejected(ctx context.Context, blockerUID, userUID string) (bool, error) {
	relationship, err := s.relationRepo.GetRelationship(ctx, blockerUID, userUID)
	if err != nil {
		return false, err
	}
	return relationship != nil && relationship.Status == "rejected", nil
}

// IsFriend 检查是否是好友（互相关注）
func (s *RelationshipService) IsFriend(ctx context.Context, userUID1, userUID2 string) (bool, error) {
	// 获取双向关系
	rel1, err := s.relationRepo.GetRelationship(ctx, userUID1, userUID2)
	if err != nil || rel1 == nil || rel1.Status != "accepted" {
		return false, err
	}

	rel2, err := s.relationRepo.GetRelationship(ctx, userUID2, userUID1)
	if err != nil || rel2 == nil || rel2.Status != "accepted" {
		return false, err
	}

	return true, nil
}

// 处理互相关注（好友）情况
func (s *RelationshipService) handleMutualFollow(ctx context.Context, userUID1, userUID2 string) {
	isFriend, err := s.IsFriend(ctx, userUID1, userUID2)
	if err != nil {
		logger.Error("failed to check friend status",
			logger.Any("error", err),
			logger.String("user1", userUID1),
			logger.String("user2", userUID2))
		return
	}

	if isFriend {
		// 创建私聊房间
		_, err := s.chatService.CreatePrivateRoom(ctx, userUID1, userUID2)
		if err != nil {
			logger.Error("failed to create private room",
				logger.Any("error", err),
				logger.String("user1", userUID1),
				logger.String("user2", userUID2))
		}
	}
}
