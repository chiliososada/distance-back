package service

import (
	"context"
	"fmt"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"
	"github.com/chiliososada/distance-back/pkg/logger"
	"github.com/chiliososada/distance-back/pkg/storage"
)

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
		return nil, ErrUserNotFound
	}
	user2, err := s.userRepo.GetByID(ctx, userID2)
	if err != nil || user2 == nil {
		return nil, ErrUserNotFound
	}

	// 检查是否已经存在私聊房间
	existingRoom, err := s.findPrivateRoom(ctx, userID1, userID2)
	if err != nil {
		return nil, err
	}
	if existingRoom != nil {
		return existingRoom, nil
	}

	// 创建新的私聊房间
	room := &model.ChatRoom{
		Name: fmt.Sprintf("%s & %s", user1.Nickname, user2.Nickname),
		Type: "individual",
	}

	// 创建房间
	if err := s.chatRepo.CreateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 添加成员
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
func (s *ChatService) CreateGroupRoom(ctx context.Context, creatorID uint64, name string, initialMembers []uint64) (*model.ChatRoom, error) {
	// 验证创建者
	creator, err := s.userRepo.GetByID(ctx, creatorID)
	if err != nil || creator == nil {
		return nil, ErrUserNotFound
	}

	// 验证初始成员数量
	if len(initialMembers) > s.maxRoomMembers {
		return nil, fmt.Errorf("number of members exceeds maximum limit of %d", s.maxRoomMembers)
	}

	// 创建群聊房间
	room := &model.ChatRoom{
		Name: name,
		Type: "group",
	}

	// 创建房间
	if err := s.chatRepo.CreateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 添加创建者为管理员
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
	for _, memberID := range initialMembers {
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
func (s *ChatService) SendMessage(ctx context.Context, userID uint64, roomID uint64, msgType string, content string, files []*model.File) (*model.Message, error) {
	// 检查发送者是否是房间成员
	if !s.isRoomMember(ctx, roomID, userID) {
		return nil, ErrNotRoomMember
	}

	// 创建消息
	msg := &model.Message{
		ChatRoomID:  roomID,
		SenderID:    userID,
		ContentType: msgType,
		Content:     content,
	}

	// 发送消息
	if err := s.chatRepo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// 处理媒体文件
	if len(files) > 0 {
		for _, file := range files {
			// 上传文件
			fileURL, err := s.storage.UploadFile(ctx, file.File, storage.ChatDirectory)
			if err != nil {
				logger.Error("failed to upload message media",
					logger.Any("error", err),
					logger.Uint64("message_id", msg.ID))
				continue
			}

			// 创建媒体记录
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

	// 更新房间成员的未读消息状态
	// 实际项目中，这里应该通过消息队列异步处理
	go s.updateMembersUnreadStatus(ctx, roomID, msg.ID)

	return msg, nil
}

// GetMessages 获取消息历史
func (s *ChatService) GetMessages(ctx context.Context, userID, roomID uint64, beforeID uint64, limit int) ([]*model.Message, error) {
	// 检查用户是否是房间成员
	if !s.isRoomMember(ctx, roomID, userID) {
		return nil, ErrNotRoomMember
	}

	if limit <= 0 || limit > DefaultMessageLimit {
		limit = DefaultMessageLimit
	}

	return s.chatRepo.GetMessagesByRoom(ctx, roomID, beforeID, limit)
}

// MarkMessagesAsRead 标记消息为已读
func (s *ChatService) MarkMessagesAsRead(ctx context.Context, userID, roomID uint64, messageID uint64) error {
	// 更新成员的最后读取消息ID
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}

	member.LastReadMessageID = messageID
	return s.chatRepo.UpdateMember(ctx, member)
}

// AddMember 添加成员到群聊
func (s *ChatService) AddMember(ctx context.Context, operatorID, roomID, userID uint64) error {
	// 检查操作者权限
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return ErrForbidden
	}

	// 检查用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	// 检查是否已是成员
	existingMember, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if existingMember != nil {
		return ErrConflict
	}

	// 检查成员数量限制
	members, err := s.chatRepo.GetRoomMembers(ctx, roomID)
	if err != nil {
		return err
	}
	if len(members) >= s.maxRoomMembers {
		return fmt.Errorf("room member limit reached")
	}

	// 添加新成员
	member := &model.ChatRoomMember{
		ChatRoomID: roomID,
		UserID:     userID,
		Role:       "member",
		Nickname:   user.Nickname,
	}

	return s.chatRepo.AddMember(ctx, member)
}

// RemoveMember 从群聊中移除成员
func (s *ChatService) RemoveMember(ctx context.Context, operatorID, roomID, userID uint64) error {
	// 检查操作者权限
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return ErrForbidden
	}

	// 不能移除群主
	targetMember, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if targetMember == nil {
		return ErrNotRoomMember
	}
	if targetMember.Role == "owner" {
		return ErrForbidden
	}

	return s.chatRepo.RemoveMember(ctx, roomID, userID)
}

// UpdateMemberRole 更新成员角色
func (s *ChatService) UpdateMemberRole(ctx context.Context, operatorID, roomID, userID uint64, newRole string) error {
	// 检查操作者权限
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role != "owner" {
		return ErrForbidden
	}

	// 获取目标成员
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}

	// 更新角色
	member.Role = newRole
	return s.chatRepo.UpdateMember(ctx, member)
}

// GetRoomInfo 获取聊天室信息
func (s *ChatService) GetRoomInfo(ctx context.Context, roomID uint64) (*model.ChatRoom, error) {
	return s.chatRepo.GetRoomByID(ctx, roomID)
}

// ListUserRooms 获取用户的聊天室列表
func (s *ChatService) ListUserRooms(ctx context.Context, userID uint64, page, pageSize int) ([]*model.ChatRoom, int64, error) {
	offset := (page - 1) * pageSize
	return s.chatRepo.ListUserRooms(ctx, userID, offset, pageSize)
}

// 辅助方法

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
		// 更新未读消息数
		if member.LastReadMessageID < messageID {
			// 这里应该通过WebSocket或推送通知用户有新消息
			// 实际项目中应该通过消息队列处理
			logger.Info("new message notification",
				logger.Uint64("user_id", member.UserID),
				logger.Uint64("room_id", roomID),
				logger.Uint64("message_id", messageID))
		}
	}
}

// PinRoom 置顶聊天室
func (s *ChatService) PinRoom(ctx context.Context, userID, roomID uint64) error {
	// 检查用户是否是房间成员
	if !s.isRoomMember(ctx, roomID, userID) {
		return ErrNotRoomMember
	}

	return s.chatRepo.PinRoom(ctx, userID, roomID)
}

// UnpinRoom 取消置顶聊天室
func (s *ChatService) UnpinRoom(ctx context.Context, userID, roomID uint64) error {
	return s.chatRepo.UnpinRoom(ctx, userID, roomID)
}

// GetPinnedRooms 获取用户置顶的聊天室列表
func (s *ChatService) GetPinnedRooms(ctx context.Context, userID uint64) ([]*model.ChatRoom, error) {
	return s.chatRepo.GetPinnedRooms(ctx, userID)
}

// UpdateRoomInfo 更新聊天室信息
func (s *ChatService) UpdateRoomInfo(ctx context.Context, operatorID uint64, room *model.ChatRoom) error {
	// 检查操作者权限
	member, err := s.getMemberInfo(ctx, room.ID, operatorID)
	if err != nil {
		return err
	}
	if member == nil || member.Role == "member" {
		return ErrForbidden
	}

	// 获取现有房间信息
	existingRoom, err := s.chatRepo.GetRoomByID(ctx, room.ID)
	if err != nil {
		return err
	}
	if existingRoom == nil {
		return ErrChatRoomNotFound
	}

	// 只更新允许的字段
	existingRoom.Name = room.Name
	existingRoom.Announcement = room.Announcement
	if room.AvatarURL != "" {
		existingRoom.AvatarURL = room.AvatarURL
	}

	return s.chatRepo.UpdateRoom(ctx, existingRoom)
}

// UpdateRoomAvatar 更新聊天室头像
func (s *ChatService) UpdateRoomAvatar(ctx context.Context, operatorID uint64, roomID uint64, avatar *model.File) error {
	// 检查操作者权限
	member, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if member == nil || member.Role == "member" {
		return ErrForbidden
	}

	// 上传新头像
	fileURL, err := s.storage.UploadFile(ctx, avatar.File, storage.ChatDirectory)
	if err != nil {
		return fmt.Errorf("failed to upload avatar: %w", err)
	}

	// 更新房间头像
	room, err := s.chatRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrChatRoomNotFound
	}

	room.AvatarURL = fileURL
	return s.chatRepo.UpdateRoom(ctx, room)
}

// MuteMember 将成员禁言
func (s *ChatService) MuteMember(ctx context.Context, operatorID, roomID, userID uint64) error {
	// 检查操作者权限
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return ErrForbidden
	}

	// 获取目标成员
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}

	// 不能禁言权限更高的成员
	if member.Role == "owner" || (member.Role == "admin" && operatorMember.Role != "owner") {
		return ErrForbidden
	}

	member.IsMuted = true
	return s.chatRepo.UpdateMember(ctx, member)
}

// UnmuteMember 解除成员禁言
func (s *ChatService) UnmuteMember(ctx context.Context, operatorID, roomID, userID uint64) error {
	// 检查操作者权限
	operatorMember, err := s.getMemberInfo(ctx, roomID, operatorID)
	if err != nil {
		return err
	}
	if operatorMember == nil || operatorMember.Role == "member" {
		return ErrForbidden
	}

	// 获取目标成员
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}

	member.IsMuted = false
	return s.chatRepo.UpdateMember(ctx, member)
}

// UpdateMemberNickname 更新成员在群内的昵称
func (s *ChatService) UpdateMemberNickname(ctx context.Context, roomID, userID uint64, nickname string) error {
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}

	member.Nickname = nickname
	return s.chatRepo.UpdateMember(ctx, member)
}

// GetUnreadCount 获取未读消息数
func (s *ChatService) GetUnreadCount(ctx context.Context, userID, roomID uint64) (uint64, error) {
	member, err := s.getMemberInfo(ctx, roomID, userID)
	if err != nil {
		return 0, err
	}
	if member == nil {
		return 0, ErrNotRoomMember
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

	// 计算未读消息数
	return latestMessageID - member.LastReadMessageID, nil
}

// SearchMessages 搜索消息
func (s *ChatService) SearchMessages(ctx context.Context, userID, roomID uint64, keyword string, page, pageSize int) ([]*model.Message, int64, error) {
	if !s.isRoomMember(ctx, roomID, userID) {
		return nil, 0, ErrNotRoomMember
	}

	// TODO: 实现消息搜索功能
	return nil, 0, nil
}

// GetMemberList 获取聊天室成员列表
func (s *ChatService) GetMemberList(ctx context.Context, roomID uint64) ([]*model.ChatRoomMember, error) {
	return s.chatRepo.GetRoomMembers(ctx, roomID)
}
