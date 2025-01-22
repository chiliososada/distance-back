package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/chiliososada/distance-back/pkg/storage"
	"github.com/google/uuid"
)

// 服务层参数结构定义
type (
	// GroupCreateOptions 创建群聊参数
	GroupCreateOptions struct {
		Name            string
		TopicUID        string // 添加话题ID
		Announcement    string
		CreatorNickname string // 添加创建者昵称
	}
	// MessageCreateOptions 发送消息参数
	MessageCreateOptions struct {
		ContentType string
		Content     string
		Files       []*model.File
	}

	// RoomUpdateOptions 更新聊天室参数
	RoomUpdateOptions struct {
		Name         string
		Announcement string
	}

	// MemberUpdateOptions 更新成员参数
	MemberUpdateOptions struct {
		Role     string
		Nickname string
		IsMuted  *bool
	}
)

// ChatService 聊天服务
type ChatService struct {
	chatRepo       repository.ChatRepository
	userRepo       repository.UserRepository
	relationRepo   repository.RelationshipRepository
	storage        storage.Storage
	maxRoomMembers int
}

const (
	DefaultMaxRoomMembers = 500
	DefaultMessageLimit   = 50
)

// NewChatService 创建聊天服务实例
func NewChatService(
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	relationRepo repository.RelationshipRepository,
	storage storage.Storage,
) *ChatService {
	return &ChatService{
		chatRepo:       chatRepo,
		userRepo:       userRepo,
		relationRepo:   relationRepo,
		storage:        storage,
		maxRoomMembers: DefaultMaxRoomMembers,
	}
}

// CreatePrivateRoom 创建私聊房间
func (s *ChatService) CreatePrivateRoom(ctx context.Context, userUID1, userUID2 string) (*model.ChatRoom, error) {
	// 1. 检查用户是否存在
	user1, err := s.userRepo.GetByUID(ctx, userUID1)
	if err != nil || user1 == nil {
		return nil, errors.ErrUserNotFound
	}

	user2, err := s.userRepo.GetByUID(ctx, userUID2)
	if err != nil || user2 == nil {
		return nil, errors.ErrUserNotFound
	}

	// 2. 检查是否已经存在私聊房间
	existingRoom, err := s.chatRepo.FindPrivateRoom(ctx, userUID1, userUID2)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing room: %w", err)
	}
	if existingRoom != nil {
		return existingRoom, nil
	}

	// 3. 创建新的私聊房间
	now := time.Now()
	room := &model.ChatRoom{
		BaseModel: model.BaseModel{
			UID:       uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name: fmt.Sprintf("%s & %s", user1.Nickname, user2.Nickname),
		Type: "individual",
	}

	// 使用事务创建房间和成员
	err = s.chatRepo.CreateRoom(ctx, room)
	if err != nil {
		logger.Error("Failed to create chat room",
			logger.Any("error", err))
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 4. 创建成员记录
	members := []*model.ChatRoomMember{
		{
			BaseModel: model.BaseModel{
				UID:       uuid.New().String(),
				CreatedAt: now,
				UpdatedAt: now,
			},
			ChatRoomUID: room.UID,
			UserUID:     userUID1,
			Role:        "member",
			Nickname:    user1.Nickname,
		},
		{
			BaseModel: model.BaseModel{
				UID:       uuid.New().String(),
				CreatedAt: now,
				UpdatedAt: now,
			},
			ChatRoomUID: room.UID,
			UserUID:     userUID2,
			Role:        "member",
			Nickname:    user2.Nickname,
		},
	}

	// 添加成员
	for _, member := range members {
		if err := s.chatRepo.AddMember(ctx, member); err != nil {
			logger.Error("Failed to add member",
				logger.String("room_uid", room.UID),
				logger.String("user_uid", member.UserUID),
				logger.Any("error", err))
			return nil, fmt.Errorf("failed to add member: %w", err)
		}
	}

	logger.Info("Successfully created private room",
		logger.String("room_uid", room.UID),
		logger.String("user1", userUID1),
		logger.String("user2", userUID2))

	return room, nil
}

// CreateGroupRoom 创建群聊房间
func (s *ChatService) CreateGroupRoom(ctx context.Context, creatorUID string, opts GroupCreateOptions) (*model.ChatRoom, error) {
	room := &model.ChatRoom{
		BaseModel: model.BaseModel{
			UID: uuid.New().String(),
		},
		Name:         opts.Name,
		Type:         "group",
		TopicUID:     opts.TopicUID,
		Announcement: opts.Announcement,
	}

	if err := s.chatRepo.CreateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 只添加群主
	creatorMember := &model.ChatRoomMember{
		BaseModel: model.BaseModel{
			UID: uuid.New().String(),
		},
		ChatRoomUID: room.UID,
		UserUID:     creatorUID,
		Role:        "owner",
		Nickname:    opts.CreatorNickname,
	}
	if err := s.chatRepo.AddMember(ctx, creatorMember); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	return room, nil
}

// SendMessage 发送消息
func (s *ChatService) SendMessage(ctx context.Context, userUID string, roomUID string, opts MessageCreateOptions) (*model.Message, error) {
	if !s.isRoomMember(ctx, roomUID, userUID) {
		return nil, errors.ErrNotChatMember
	}

	msg := &model.Message{
		ChatRoomUID: roomUID,
		SenderUID:   userUID,
		ContentType: opts.ContentType,
		Content:     opts.Content,
	}

	if err := s.chatRepo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	if len(opts.Files) > 0 {
		for _, file := range opts.Files {
			fileURL, err := s.storage.UploadFile(ctx, file.File, storage.ChatDirectory)
			if err != nil {
				logger.Error("failed to upload message media",
					logger.Any("error", err),
					logger.String("message_uid", msg.UID))
				continue
			}

			media := &model.MessageMedia{
				MessageUID: msg.UID,
				MediaType:  file.Type,
				MediaURL:   fileURL,
				FileName:   file.Name,
				FileSize:   file.Size,
			}

			if err := s.chatRepo.AddMessageMedia(ctx, media); err != nil {
				logger.Error("failed to save message media",
					logger.Any("error", err),
					logger.String("message_uid", msg.UID))
			}
		}
	}

	go s.updateMembersUnreadStatus(ctx, roomUID, msg.UID)

	return msg, nil
}

// GetMessages 获取消息历史
func (s *ChatService) GetMessages(ctx context.Context, userUID string, roomUID string, beforeUID string, limit int) ([]*model.Message, error) {
	if !s.isRoomMember(ctx, roomUID, userUID) {
		return nil, errors.ErrNotChatMember
	}

	if limit <= 0 || limit > DefaultMessageLimit {
		limit = DefaultMessageLimit
	}

	return s.chatRepo.GetMessagesByRoom(ctx, roomUID, beforeUID, limit)
}

// MarkMessagesAsRead 标记消息已读
func (s *ChatService) MarkMessagesAsRead(ctx context.Context, userUID string, roomUID string, messageUID string) error {
	member, err := s.getMemberInfo(ctx, roomUID, userUID)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.ErrNotChatMember
	}

	// 验证消息是否存在且属于该聊天室
	messages, err := s.chatRepo.GetMessagesByRoom(ctx, roomUID, messageUID, 1)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		return errors.New(errors.CodeMessageNotFound, "Message not found")
	}

	member.LastReadMessageID = messageUID
	return s.chatRepo.UpdateMember(ctx, member)
}

// UpdateRoom 更新聊天室信息
func (s *ChatService) UpdateRoom(ctx context.Context, operatorUID string, roomUID string, opts RoomUpdateOptions) error {
	member, err := s.getMemberInfo(ctx, roomUID, operatorUID)
	if err != nil {
		return err
	}
	if member == nil || member.Role == "member" {
		return errors.ErrForbidden
	}

	room, err := s.chatRepo.GetRoomByUID(ctx, roomUID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.ErrChatRoomNotFound
	}

	if opts.Name != "" {
		room.Name = opts.Name
	}
	if opts.Announcement != "" {
		room.Announcement = opts.Announcement
	}

	return s.chatRepo.UpdateRoom(ctx, room)
}

// UpdateRoomAvatar 更新聊天室头像
func (s *ChatService) UpdateRoomAvatar(ctx context.Context, operatorUID string, roomUID string, avatar *model.File) (*model.ChatRoom, error) {
	member, err := s.getMemberInfo(ctx, roomUID, operatorUID)
	if err != nil {
		return nil, err
	}
	if member == nil || member.Role == "member" {
		return nil, errors.ErrForbidden
	}

	fileURL, err := s.storage.UploadFile(ctx, avatar.File, storage.ChatDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	room, err := s.chatRepo.GetRoomByUID(ctx, roomUID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.ErrChatRoomNotFound
	}

	room.AvatarURL = fileURL
	if err := s.chatRepo.UpdateRoom(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// AddMember 添加成员到群聊
func (s *ChatService) AddMember(ctx context.Context, operatorUID string, roomUID string, userUID string) error {
	operatorMember, err := s.getMemberInfo(ctx, roomUID, operatorUID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return errors.ErrForbidden
	}

	user, err := s.userRepo.GetByUID(ctx, userUID)
	if err != nil || user == nil {
		return errors.ErrUserNotFound
	}

	existingMember, err := s.getMemberInfo(ctx, roomUID, userUID)
	if err != nil {
		return err
	}
	if existingMember != nil {
		return errors.New(errors.CodeDuplicate, "User is already a member")
	}

	members, err := s.chatRepo.GetRoomMembers(ctx, roomUID)
	if err != nil {
		return err
	}
	if len(members) >= s.maxRoomMembers {
		return fmt.Errorf("room member limit reached")
	}

	member := &model.ChatRoomMember{
		ChatRoomUID: roomUID,
		UserUID:     userUID,
		Role:        "member",
		Nickname:    user.Nickname,
	}

	return s.chatRepo.AddMember(ctx, member)
}

// RemoveMember 从群聊中移除成员
func (s *ChatService) RemoveMember(ctx context.Context, operatorUID string, roomUID string, userUID string) error {
	operatorMember, err := s.getMemberInfo(ctx, roomUID, operatorUID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return errors.ErrForbidden
	}

	targetMember, err := s.getMemberInfo(ctx, roomUID, userUID)
	if err != nil {
		return err
	}
	if targetMember == nil {
		return errors.ErrNotChatMember
	}
	if targetMember.Role == "owner" {
		return errors.ErrForbidden
	}

	return s.chatRepo.RemoveMember(ctx, roomUID, userUID)
}

// UpdateMember 更新成员信息
func (s *ChatService) UpdateMember(ctx context.Context, operatorUID string, roomUID string, memberUID string, opts MemberUpdateOptions) error {
	operatorMember, err := s.getMemberInfo(ctx, roomUID, operatorUID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return errors.ErrForbidden
	}

	member, err := s.getMemberInfo(ctx, roomUID, memberUID)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.ErrNotChatMember
	}

	if member.Role == "owner" || (member.Role == "admin" && operatorMember.Role != "owner") {
		return errors.ErrForbidden
	}

	if opts.Role != "" {
		member.Role = opts.Role
	}
	if opts.Nickname != "" {
		member.Nickname = opts.Nickname
	}
	if opts.IsMuted != nil {
		member.IsMuted = *opts.IsMuted
	}

	return s.chatRepo.UpdateMember(ctx, member)
}

// GetRoomInfo 获取聊天室信息
func (s *ChatService) GetRoomInfo(ctx context.Context, roomUID string) (*model.ChatRoom, error) {
	room, err := s.chatRepo.GetRoomByUID(ctx, roomUID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.ErrChatRoomNotFound
	}
	return room, nil
}

// ListUserRooms 获取用户的聊天室列表
func (s *ChatService) ListUserRooms(ctx context.Context, userUID string, page, pageSize int) ([]*model.ChatRoom, int64, error) {
	offset := (page - 1) * pageSize
	return s.chatRepo.ListUserRooms(ctx, userUID, offset, pageSize)
}

// GetUnreadCount 获取未读消息数
func (s *ChatService) GetUnreadCount(ctx context.Context, userUID string, roomUID string) (int64, error) {
	member, err := s.getMemberInfo(ctx, roomUID, userUID)
	if err != nil {
		return 0, err
	}
	if member == nil {
		return 0, errors.ErrNotChatMember
	}

	// 获取最新消息
	messages, err := s.chatRepo.GetLatestMessages(ctx, roomUID, 1)
	if err != nil {
		return 0, err
	}
	if len(messages) == 0 {
		return 0, nil
	}

	latestMessage := messages[0]
	if member.LastReadMessageID == latestMessage.UID {
		return 0, nil
	}

	// 获取未读消息数量
	unreadMessages, err := s.chatRepo.GetMessagesByRoom(ctx, roomUID, member.LastReadMessageID, DefaultMessageLimit)
	if err != nil {
		return 0, err
	}

	return int64(len(unreadMessages)), nil
}

// LeaveRoom 退出聊天室
func (s *ChatService) LeaveRoom(ctx context.Context, userUID string, roomUID string) error {
	// 1. 获取聊天室信息
	room, err := s.chatRepo.GetRoomByUID(ctx, roomUID)
	if err != nil {
		return fmt.Errorf("failed to get chat room: %w", err)
	}
	if room == nil {
		return errors.ErrChatRoomNotFound
	}

	// 2. 获取所有成员
	members, err := s.chatRepo.GetRoomMembers(ctx, roomUID)
	if err != nil {
		return fmt.Errorf("failed to get room members: %w", err)
	}

	// 3. 检查当前用户身份
	var isOwner bool
	var currentMember *model.ChatRoomMember
	for _, member := range members {
		if member.UserUID == userUID {
			currentMember = member
			isOwner = member.Role == "owner"
			break
		}
	}

	if currentMember == nil {
		return errors.ErrNotChatMember
	}

	// 4. 如果是群主
	if isOwner {
		// 统计除群主外的成员数
		otherMembers := make([]*model.ChatRoomMember, 0)
		for _, member := range members {
			if member.UserUID != userUID {
				otherMembers = append(otherMembers, member)
			}
		}

		// 如果没有其他成员，软删除聊天室和关联的话题
		if len(otherMembers) == 0 {
			return s.chatRepo.SoftDeleteTopicAndRoom(ctx, room.TopicUID, room.UID)
		}

		// 有其他成员，转移群主身份给最早加入的成员
		newOwner := otherMembers[0] // 按加入时间排序，第一个就是最早的
		newOwner.Role = "owner"

		if err := s.chatRepo.UpdateMember(ctx, newOwner); err != nil {
			return fmt.Errorf("failed to update new owner: %w", err)
		}

		// 移除当前群主
		if err := s.chatRepo.RemoveMember(ctx, roomUID, userUID); err != nil {
			return fmt.Errorf("failed to remove owner: %w", err)
		}

		return nil
	}

	// 5. 如果是普通成员，软删除成员记录
	return s.chatRepo.RemoveMember(ctx, roomUID, userUID)
}

// GetRoomMembers 获取成员列表
func (s *ChatService) GetRoomMembers(ctx context.Context, roomUID string, page, size int) ([]*model.ChatRoomMember, int64, error) {
	// 获取总数
	members, err := s.chatRepo.GetRoomMembers(ctx, roomUID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(members))

	// 计算分页
	start := (page - 1) * size
	end := start + size
	if start >= len(members) {
		return []*model.ChatRoomMember{}, total, nil
	}
	if end > len(members) {
		end = len(members)
	}

	return members[start:end], total, nil
}

// PinRoom 置顶聊天室
func (s *ChatService) PinRoom(ctx context.Context, userUID string, roomUID string) error {
	if !s.isRoomMember(ctx, roomUID, userUID) {
		return errors.ErrNotChatMember
	}

	return s.chatRepo.PinRoom(ctx, userUID, roomUID)
}

// UnpinRoom 取消置顶聊天室
func (s *ChatService) UnpinRoom(ctx context.Context, userUID string, roomUID string) error {
	return s.chatRepo.UnpinRoom(ctx, userUID, roomUID)
}

// GetPinnedRooms 获取置顶的聊天室列表
func (s *ChatService) GetPinnedRooms(ctx context.Context, userUID string) ([]*model.ChatRoom, error) {
	return s.chatRepo.GetPinnedRooms(ctx, userUID)
}

// updateMembersUnreadStatus 更新成员未读状态
func (s *ChatService) updateMembersUnreadStatus(ctx context.Context, roomUID string, messageUID string) {
	members, err := s.chatRepo.GetRoomMembers(ctx, roomUID)
	if err != nil {
		logger.Error("failed to get room members",
			logger.String("room_uid", roomUID),
			logger.Any("error", err))
		return
	}

	for _, member := range members {
		if member.LastReadMessageID < messageUID {
			logger.Info("new message notification",
				logger.String("user_uid", member.UserUID),
				logger.String("room_uid", roomUID),
				logger.String("message_uid", messageUID))
		}
	}
}

// Helper methods

// isRoomMember 检查用户是否是房间成员
func (s *ChatService) isRoomMember(ctx context.Context, roomUID, userUID string) bool {
	member, _ := s.getMemberInfo(ctx, roomUID, userUID)
	return member != nil
}

// getMemberInfo 获取成员信息
func (s *ChatService) getMemberInfo(ctx context.Context, roomUID, userUID string) (*model.ChatRoomMember, error) {
	members, err := s.chatRepo.GetRoomMembers(ctx, roomUID)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if member.UserUID == userUID {
			return member, nil
		}
	}

	return nil, nil
}

// findPrivateRoom 查找两个用户之间的私聊房间

func (s *ChatService) findPrivateRoom(ctx context.Context, userUID1, userUID2 string) (*model.ChatRoom, error) {
	return s.chatRepo.FindPrivateRoom(ctx, userUID1, userUID2)
}

// JoinRoom 加入群聊
func (s *ChatService) JoinRoom(ctx context.Context, userUID string, nickname string, roomUID string) error {
	// 获取聊天室信息
	room, err := s.GetRoomInfo(ctx, roomUID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.ErrChatRoomNotFound
	}

	// 验证是否为群聊
	if room.Type != "group" {
		return errors.ErrChatRoomNotFound
	}

	// 检查是否已经是成员
	if s.isRoomMember(ctx, roomUID, userUID) {
		return errors.ErrUserExists
	}

	// 构造成员对象
	member := &model.ChatRoomMember{
		ChatRoomUID: roomUID,
		UserUID:     userUID,
		Role:        "member",
		Nickname:    nickname,
	}

	// 添加成员
	return s.chatRepo.AddMember(ctx, member)
}
