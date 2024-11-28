package mysql

import (
	"context"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/internal/repository"

	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

// NewChatRepository 创建聊天仓储实例
func NewChatRepository(db *gorm.DB) repository.ChatRepository {
	return &chatRepository{db: db}
}

// CreateRoom 创建聊天室
func (r *chatRepository) CreateRoom(ctx context.Context, room *model.ChatRoom) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建聊天室
		if err := tx.Create(room).Error; err != nil {
			return err
		}
		return nil
	})
}

// UpdateRoom 更新聊天室信息
func (r *chatRepository) UpdateRoom(ctx context.Context, room *model.ChatRoom) error {
	return r.db.WithContext(ctx).Save(room).Error
}

// GetRoomByID 获取聊天室信息
func (r *chatRepository) GetRoomByID(ctx context.Context, id uint64) (*model.ChatRoom, error) {
	var room model.ChatRoom
	err := r.db.WithContext(ctx).
		First(&room, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

// ListUserRooms 获取用户的聊天室列表
func (r *chatRepository) ListUserRooms(ctx context.Context, userID uint64, offset, limit int) ([]*model.ChatRoom, int64, error) {
	var rooms []*model.ChatRoom
	var total int64

	subQuery := r.db.Model(&model.ChatRoomMember{}).
		Select("chat_room_id").
		Where("user_id = ?", userID)

	db := r.db.WithContext(ctx).
		Where("id IN (?)", subQuery)

	// 获取总数
	if err := db.Model(&model.ChatRoom{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取聊天室列表，包括成员信息和最后一条消息
	err := db.Preload("ChatRoomMembers", func(db *gorm.DB) *gorm.DB {
		return db.Order("joined_at DESC")
	}).
		Preload("ChatRoomMembers.User").
		Order("updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rooms).Error

	if err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

// AddMember 添加聊天室成员
func (r *chatRepository) AddMember(ctx context.Context, member *model.ChatRoomMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查是否已经是成员
		var count int64
		if err := tx.Model(&model.ChatRoomMember{}).
			Where("chat_room_id = ? AND user_id = ?", member.ChatRoomID, member.UserID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil // 已经是成员，直接返回
		}

		// 添加成员
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		return nil
	})
}

// RemoveMember 移除聊天室成员
func (r *chatRepository) RemoveMember(ctx context.Context, roomID, userID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 移除成员
		if err := tx.Where("chat_room_id = ? AND user_id = ?", roomID, userID).
			Delete(&model.ChatRoomMember{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdateMember 更新成员信息
func (r *chatRepository) UpdateMember(ctx context.Context, member *model.ChatRoomMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// GetRoomMembers 获取聊天室成员列表
func (r *chatRepository) GetRoomMembers(ctx context.Context, roomID uint64) ([]*model.ChatRoomMember, error) {
	var members []*model.ChatRoomMember
	err := r.db.WithContext(ctx).
		Where("chat_room_id = ?", roomID).
		Preload("User").
		Order("role DESC, joined_at ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// CreateMessage 创建消息
func (r *chatRepository) CreateMessage(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建消息
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		// 更新聊天室最后更新时间
		if err := tx.Model(&model.ChatRoom{}).
			Where("id = ?", message.ChatRoomID).
			UpdateColumn("updated_at", time.Now()).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetMessagesByRoom 获取聊天室消息列表（向前加载）
func (r *chatRepository) GetMessagesByRoom(ctx context.Context, roomID uint64, beforeID uint64, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	query := r.db.WithContext(ctx).
		Where("chat_room_id = ?", roomID).
		Preload("Sender").
		Preload("MessageMedia").
		Order("id DESC").
		Limit(limit)

	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}

	// 反转消息列表，使其按时间正序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetLatestMessages 获取聊天室最新消息
func (r *chatRepository) GetLatestMessages(ctx context.Context, roomID uint64, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("chat_room_id = ?", roomID).
		Preload("Sender").
		Preload("MessageMedia").
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 反转消息列表，使其按时间正序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// AddMessageMedia 添加消息媒体
func (r *chatRepository) AddMessageMedia(ctx context.Context, media *model.MessageMedia) error {
	return r.db.WithContext(ctx).Create(media).Error
}

// GetMessageMedia 获取消息媒体列表
func (r *chatRepository) GetMessageMedia(ctx context.Context, messageID uint64) ([]*model.MessageMedia, error) {
	var media []*model.MessageMedia
	err := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Find(&media).Error
	if err != nil {
		return nil, err
	}
	return media, nil
}

// PinRoom 置顶聊天室
func (r *chatRepository) PinRoom(ctx context.Context, userID, roomID uint64) error {
	pinned := &model.PinnedChatRoom{
		UserID:     userID,
		ChatRoomID: roomID,
		PinnedAt:   time.Now(),
	}
	return r.db.WithContext(ctx).Create(pinned).Error
}

// UnpinRoom 取消置顶聊天室
func (r *chatRepository) UnpinRoom(ctx context.Context, userID, roomID uint64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND chat_room_id = ?", userID, roomID).
		Delete(&model.PinnedChatRoom{}).Error
}

// GetPinnedRooms 获取用户置顶的聊天室列表
func (r *chatRepository) GetPinnedRooms(ctx context.Context, userID uint64) ([]*model.ChatRoom, error) {
	var rooms []*model.ChatRoom
	err := r.db.WithContext(ctx).
		Joins("JOIN pinned_chat_rooms ON pinned_chat_rooms.chat_room_id = chat_rooms.id").
		Where("pinned_chat_rooms.user_id = ?", userID).
		Preload("ChatRoomMembers", func(db *gorm.DB) *gorm.DB {
			return db.Order("joined_at DESC")
		}).
		Preload("ChatRoomMembers.User").
		Order("pinned_chat_rooms.pinned_at DESC").
		Find(&rooms).Error
	if err != nil {
		return nil, err
	}
	return rooms, nil
}
