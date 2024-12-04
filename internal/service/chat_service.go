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
		InitialMembers []string // 改为使用 UID
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
	// 检查用户是否存在
	user1, err := s.userRepo.GetByUID(ctx, userUID1)
	if err != nil || user1 == nil {
		return nil, errors.ErrUserNotFound
	}
	user2, err := s.userRepo.GetByUID(ctx, userUID2)
	if err != nil || user2 == nil {
		return nil, errors.ErrUserNotFound
	}

	// 检查是否已经存在私聊房间
	existingRoom, err := s.findPrivateRoom(ctx, userUID1, userUID2)
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
			ChatRoomUID: room.UID,
			UserUID:     userUID1,
			Role:        "member",
			Nickname:    user1.Nickname,
		},
		{
			ChatRoomUID: room.UID,
			UserUID:     userUID2,
			Role:        "member",
			Nickname:    user2.Nickname,
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
func (s *ChatService) CreateGroupRoom(ctx context.Context, creatorUID string, opts GroupCreateOptions) (*model.ChatRoom, error) {
	creator, err := s.userRepo.GetByUID(ctx, creatorUID)
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
		ChatRoomUID: room.UID,
		UserUID:     creatorUID,
		Role:        "owner",
		Nickname:    creator.Nickname,
	}
	if err := s.chatRepo.AddMember(ctx, creatorMember); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	// 添加初始成员
	for _, memberUID := range opts.InitialMembers {
		if memberUID == creatorUID {
			continue
		}

		member, err := s.userRepo.GetByUID(ctx, memberUID)
		if err != nil || member == nil {
			continue
		}

		roomMember := &model.ChatRoomMember{
			ChatRoomUID: room.UID,
			UserUID:     memberUID,
			Role:        "member",
			Nickname:    member.Nickname,
		}
		if err := s.chatRepo.AddMember(ctx, roomMember); err != nil {
			logger.Error("failed to add member to group",
				logger.String("room_uid", room.UID),
				logger.String("user_uid", memberUID),
				logger.Any("error", err))
		}
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
	member, err := s.getMemberInfo(ctx, roomUID, userUID)
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
	rooms, _, err := s.chatRepo.ListUserRooms(ctx, userUID1, 0, 1000)
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		if room.Type != "individual" {
			continue
		}

		members, err := s.chatRepo.GetRoomMembers(ctx, room.UID)
		if err != nil {
			continue
		}

		if len(members) == 2 {
			memberUIDs := map[string]bool{
				members[0].UserUID: true,
				members[1].UserUID: true,
			}
			if memberUIDs[userUID1] && memberUIDs[userUID2] {
				return room, nil
			}
		}
	}

	return nil, nil
}
