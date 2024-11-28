package service

import (
	"context"
	"fmt"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/errors"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/chiliososada/distance-back/pkg/storage"
)

// 服务层参数结构定义
type (
	// GroupCreateOptions 创建群聊参数
	GroupCreateOptions struct {
		Name           string
		Announcement   string
		InitialMembers []uint64
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
func (s *ChatService) CreatePrivateRoom(ctx context.Context, userID1, userID2 uint64) (*model.ChatRoom, error) {
	// 检查用户是否存在
	user1, err := s.userRepo.GetByID(ctx, userID1)
	if err != nil || user1 == nil {
		return nil, errors.ErrUserNotFound
	}
	user2, err := s.userRepo.GetByID(ctx, userID2)
	if err != nil || user2 == nil {
		return nil, errors.ErrUserNotFound
	}

	// 检查是否已经存在私聊房间
	existingRoom, err := s.findPrivateRoom(ctx, userID1, userID2)
	if err != nil {
		return nil, err
	}
	if existingRoom != nil {
		return existingRoom, nil
	}

	room := &model.ChatRoom{
		Name: fmt.Sprintf("%s & %s", user1.Nickname, user2.Nickname),
		Type: "individual",
	}

	if err := s.chatRepo.CreateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	members := []*model.ChatRoomMember{
		{
			ChatRoomID: room.ID,
			UserID:     userID1,
			Role:       "member",
			Nickname:   user1.Nickname,
		},
		{
			ChatRoomID: room.ID,
			UserID:     userID2,
			Role:       "member",
			Nickname:   user2.Nickname,
		},
	}

	for _, member := range members {
		if err := s.chatRepo.AddMember(ctx, member); err != nil {
			return nil, fmt.Errorf("failed to add member: %w", err)
		}
	}

	return room, nil
}

// CreateGroupRoom 创建群聊房间
func (s *ChatService) CreateGroupRoom(ctx context.Context, creatorID uint64, opts GroupCreateOptions) (*model.ChatRoom, error) {
	creator, err := s.userRepo.GetByID(ctx, creatorID)
	if err != nil || creator == nil {
		return nil, errors.ErrUserNotFound
	}

	if len(opts.InitialMembers) > s.maxRoomMembers {
		return nil, fmt.Errorf("number of members exceeds maximum limit of %d", s.maxRoomMembers)
	}

	room := &model.ChatRoom{
		Name:         opts.Name,
		Type:         "group",
		Announcement: opts.Announcement,
	}

	if err := s.chatRepo.CreateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 添加创建者为群主
	creatorMember := &model.ChatRoomMember{
		ChatRoomID: room.ID,
		UserID:     creatorID,
		Role:       "owner",
		Nickname:   creator.Nickname,
	}
	if err := s.chatRepo.AddMember(ctx, creatorMember); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	// 添加初始成员
	for _, memberID := range opts.InitialMembers {
		if memberID == creatorID {
			continue
		}

		member, err := s.userRepo.GetByID(ctx, memberID)
		if err != nil || member == nil {
			continue
		}

		roomMember := &model.ChatRoomMember{
			ChatRoomID: room.ID,
			UserID:     memberID,
			Role:       "member",
			Nickname:   member.Nickname,
		}
		if err := s.chatRepo.AddMember(ctx, roomMember); err != nil {
			logger.Error("failed to add member to group",
				logger.Uint64("room_id", room.ID),
				logger.Uint64("user_id", memberID),
				logger.Any("error", err))
		}
	}

	return room, nil
}

// SendMessage 发送消息
func (s *ChatService) SendMessage(ctx context.Context, userID uint64, roomID uint64, opts MessageCreateOptions) (*model.Message, error) {
	if !s.isRoomMember(ctx, roomID, userID) {
		return nil, errors.ErrNotChatMember
	}

	msg := &model.Message{
		ChatRoomID:  roomID,
		SenderID:    userID,
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
					logger.Uint64("message_id", msg.ID))
				continue
			}

			media := &model.MessageMedia{
				MessageID: msg.ID,
				MediaType: file.Type,
				MediaURL:  fileURL,
				FileName:  file.Name,
				FileSize:  file.Size,
			}

			if err := s.chatRepo.AddMessageMedia(ctx, media); err != nil {
				logger.Error("failed to save message media",
					logger.Any("error", err),
					logger.Uint64("message_id", msg.ID))
			}
		}
	}

	go s.updateMembersUnreadStatus(ctx, roomID, msg.ID)

	return msg, nil
}

// GetMessages 获取消息历史
func (s *ChatService) GetMessages(ctx context.Context, userID uint64, roomID uint64, beforeID uint64, limit int) ([]*model.Message, error) {
	if !s.isRoomMember(ctx, roomID, userID) {
		return nil, errors.ErrNotChatMember
	}

	if limit <= 0 || limit > DefaultMessageLimit {
		limit = DefaultMessageLimit
	}

	return s.chatRepo.GetMessagesByRoom(ctx, roomID, beforeID, limit)
}

// MarkMessagesAsRead 标记消息已读
func (s *ChatService) MarkMessagesAsRead(ctx context.Context, userID uint64, roomID uint64, messageID uint64) error {
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.ErrNotChatMember
	}

	// 验证消息是否存在且属于该聊天室
	messages, err := s.chatRepo.GetMessagesByRoom(ctx, roomID, messageID+1, 1)
	if err != nil {
		return err
	}
	if len(messages) == 0 || messages[0].ID > messageID {
		return errors.New(errors.CodeMessageNotFound, "Message not found")
	}

	member.LastReadMessageID = messageID
	return s.chatRepo.UpdateMember(ctx, member)
}

// UpdateRoom 更新聊天室信息
func (s *ChatService) UpdateRoom(ctx context.Context, operatorID uint64, roomID uint64, opts RoomUpdateOptions) error {
	member, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if member == nil || member.Role == "member" {
		return errors.ErrForbidden
	}

	room, err := s.chatRepo.GetRoomByID(ctx, roomID)
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
func (s *ChatService) UpdateRoomAvatar(ctx context.Context, operatorID uint64, roomID uint64, avatar *model.File) (*model.ChatRoom, error) {
	member, err := s.getMemberInfo(ctx, roomID, operatorID)
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

	room, err := s.chatRepo.GetRoomByID(ctx, roomID)
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
func (s *ChatService) AddMember(ctx context.Context, operatorID uint64, roomID uint64, userID uint64) error {
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return errors.ErrForbidden
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return errors.ErrUserNotFound
	}

	existingMember, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if existingMember != nil {
		return errors.New(errors.CodeDuplicate, "User is already a member")
	}

	members, err := s.chatRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return err
	}
	if len(members) >= s.maxRoomMembers {
		return fmt.Errorf("room member limit reached")
	}

	member := &model.ChatRoomMember{
		ChatRoomID: roomID,
		UserID:     userID,
		Role:       "member",
		Nickname:   user.Nickname,
	}

	return s.chatRepo.AddMember(ctx, member)
}

// RemoveMember 从群聊中移除成员
func (s *ChatService) RemoveMember(ctx context.Context, operatorID uint64, roomID uint64, userID uint64) error {
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return errors.ErrForbidden
	}

	targetMember, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if targetMember == nil {
		return errors.ErrNotChatMember
	}
	if targetMember.Role == "owner" {
		return errors.ErrForbidden
	}

	return s.chatRepo.RemoveMember(ctx, roomID, userID)
}

// UpdateMember 更新成员信息
func (s *ChatService) UpdateMember(ctx context.Context, operatorID uint64, roomID uint64, memberID uint64, opts MemberUpdateOptions) error {
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return errors.ErrForbidden
	}

	member, err := s.getMemberInfo(ctx, roomID, memberID)
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

// GetRoomInfo 获取聊天室信息 (续)
func (s *ChatService) GetRoomInfo(ctx context.Context, roomID uint64) (*model.ChatRoom, error) {
	room, err := s.chatRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.ErrChatRoomNotFound
	}
	return room, nil
}

// ListUserRooms 获取用户的聊天室列表
func (s *ChatService) ListUserRooms(ctx context.Context, userID uint64, page, pageSize int) ([]*model.ChatRoom, int64, error) {
	offset := (page - 1) * pageSize
	return s.chatRepo.ListUserRooms(ctx, userID, offset, pageSize)
}

// GetUnreadCount 获取未读消息数
func (s *ChatService) GetUnreadCount(ctx context.Context, userID uint64, roomID uint64) (uint64, error) {
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return 0, err
	}
	if member == nil {
		return 0, errors.ErrNotChatMember
	}

	// 获取最新消息ID
	messages, err := s.chatRepo.GetLatestMessages(ctx, roomID, 1)
	if err != nil {
		return 0, err
	}
	if len(messages) == 0 {
		return 0, nil
	}

	latestMessageID := messages[0].ID
	if latestMessageID <= member.LastReadMessageID {
		return 0, nil
	}

	return latestMessageID - member.LastReadMessageID, nil
}

// LeaveRoom 退出聊天室
func (s *ChatService) LeaveRoom(ctx context.Context, userID uint64, roomID uint64) error {
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return errors.ErrNotChatMember
	}

	// 群主不能退出群聊
	if member.Role == "owner" {
		return errors.New(errors.CodeForbidden, "Owner cannot leave the room")
	}

	return s.chatRepo.RemoveMember(ctx, roomID, userID)
}

// GetRoomMembers 获取成员列表
func (s *ChatService) GetRoomMembers(ctx context.Context, roomID uint64, page, size int) ([]*model.ChatRoomMember, int64, error) {
	// 获取总数
	members, err := s.chatRepo.GetRoomMembers(ctx, roomID)
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
func (s *ChatService) PinRoom(ctx context.Context, userID uint64, roomID uint64) error {
	if !s.isRoomMember(ctx, roomID, userID) {
		return errors.ErrNotChatMember
	}

	return s.chatRepo.PinRoom(ctx, userID, roomID)
}

// UnpinRoom 取消置顶聊天室
func (s *ChatService) UnpinRoom(ctx context.Context, userID uint64, roomID uint64) error {
	return s.chatRepo.UnpinRoom(ctx, userID, roomID)
}

// updateMembersUnreadStatus 更新成员未读状态
func (s *ChatService) updateMembersUnreadStatus(ctx context.Context, roomID uint64, messageID uint64) {
	members, err := s.chatRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		logger.Error("failed to get room members",
			logger.Uint64("room_id", roomID),
			logger.Any("error", err))
		return
	}

	for _, member := range members {
		if member.LastReadMessageID < messageID {
			logger.Info("new message notification",
				logger.Uint64("user_id", member.UserID),
				logger.Uint64("room_id", roomID),
				logger.Uint64("message_id", messageID))
		}
	}
}

// Helper methods

// isRoomMember 检查用户是否是房间成员
func (s *ChatService) isRoomMember(ctx context.Context, roomID, userID uint64) bool {
	member, _ := s.getMemberInfo(ctx, roomID, userID)
	return member != nil
}

// getMemberInfo 获取成员信息
func (s *ChatService) getMemberInfo(ctx context.Context, roomID, userID uint64) (*model.ChatRoomMember, error) {
	members, err := s.chatRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if member.UserID == userID {
			return member, nil
		}
	}

	return nil, nil
}

// findPrivateRoom 查找两个用户之间的私聊房间
func (s *ChatService) findPrivateRoom(ctx context.Context, userID1, userID2 uint64) (*model.ChatRoom, error) {
	rooms, _, err := s.chatRepo.ListUserRooms(ctx, userID1, 0, 1000)
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		if room.Type != "individual" {
			continue
		}

		members, err := s.chatRepo.GetRoomMembers(ctx, room.ID)
		if err != nil {
			continue
		}

		if len(members) == 2 {
			memberIDs := map[uint64]bool{
				members[0].UserID: true,
				members[1].UserID: true,
			}
			if memberIDs[userID1] && memberIDs[userID2] {
				return room, nil
			}
		}
	}

	return nil, nil
}
