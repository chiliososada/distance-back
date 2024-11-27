package repository

import (
	"DistanceBack_v1/internal/model"
	"context"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// 基础操作
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)

	// 认证相关
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	CreateAuthentication(ctx context.Context, auth *model.UserAuthentication) error
	UpdateAuthentication(ctx context.Context, auth *model.UserAuthentication) error

	// 设备相关
	CreateDevice(ctx context.Context, device *model.UserDevice) error
	UpdateDevice(ctx context.Context, device *model.UserDevice) error
	GetDeviceByToken(ctx context.Context, token string) (*model.UserDevice, error)
	GetUserDevices(ctx context.Context, userID uint64) ([]*model.UserDevice, error)

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
	Search(ctx context.Context, keyword string, offset, limit int) ([]*model.User, int64, error)
	GetNearbyUsers(ctx context.Context, lat, lng float64, radius float64, offset, limit int) ([]*model.User, int64, error)

	// 状态操作
	UpdateStatus(ctx context.Context, userID uint64, status string) error
	UpdateLastActive(ctx context.Context, userID uint64) error
}

// TopicRepository 话题仓储接口
type TopicRepository interface {
	// 基础操作
	Create(ctx context.Context, topic *model.Topic) error
	Update(ctx context.Context, topic *model.Topic) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*model.Topic, error)

	// 图片相关
	AddImages(ctx context.Context, topicID uint64, images []*model.TopicImage) error
	GetImages(ctx context.Context, topicID uint64) ([]*model.TopicImage, error)

	// 标签相关
	AddTags(ctx context.Context, topicID uint64, tagIDs []uint64) error
	RemoveTags(ctx context.Context, topicID uint64, tagIDs []uint64) error
	GetTags(ctx context.Context, topicID uint64) ([]*model.Tag, error)
	BatchCreate(ctx context.Context, tags []string) ([]uint64, error)
	ListPopular(ctx context.Context, limit int) ([]*model.Tag, error)

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]*model.Topic, int64, error)
	ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*model.Topic, int64, error)
	ListByTag(ctx context.Context, tagID uint64, offset, limit int) ([]*model.Topic, int64, error)
	GetNearbyTopics(ctx context.Context, lat, lng float64, radius float64, offset, limit int) ([]*model.Topic, int64, error)

	// 互动操作
	AddInteraction(ctx context.Context, interaction *model.TopicInteraction) error
	RemoveInteraction(ctx context.Context, topicID, userID uint64, interactionType string) error
	GetInteractions(ctx context.Context, topicID uint64, interactionType string) ([]*model.TopicInteraction, error)

	// 计数操作
	IncrementViewCount(ctx context.Context, topicID uint64) error
	UpdateCounts(ctx context.Context, topicID uint64) error
}

// ChatRepository 聊天仓储接口
type ChatRepository interface {
	// 聊天室操作
	CreateRoom(ctx context.Context, room *model.ChatRoom) error
	UpdateRoom(ctx context.Context, room *model.ChatRoom) error
	GetRoomByID(ctx context.Context, id uint64) (*model.ChatRoom, error)
	ListUserRooms(ctx context.Context, userID uint64, offset, limit int) ([]*model.ChatRoom, int64, error)

	// 成员操作
	AddMember(ctx context.Context, member *model.ChatRoomMember) error
	RemoveMember(ctx context.Context, roomID, userID uint64) error
	UpdateMember(ctx context.Context, member *model.ChatRoomMember) error
	GetRoomMembers(ctx context.Context, roomID uint64) ([]*model.ChatRoomMember, error)

	// 消息操作
	CreateMessage(ctx context.Context, message *model.Message) error
	GetMessagesByRoom(ctx context.Context, roomID uint64, beforeID uint64, limit int) ([]*model.Message, error)
	GetLatestMessages(ctx context.Context, roomID uint64, limit int) ([]*model.Message, error)

	// 媒体操作
	AddMessageMedia(ctx context.Context, media *model.MessageMedia) error
	GetMessageMedia(ctx context.Context, messageID uint64) ([]*model.MessageMedia, error)

	// 置顶操作
	PinRoom(ctx context.Context, userID, roomID uint64) error
	UnpinRoom(ctx context.Context, userID, roomID uint64) error
	GetPinnedRooms(ctx context.Context, userID uint64) ([]*model.ChatRoom, error)
}

// RelationshipRepository 关系仓储接口
type RelationshipRepository interface {
	// 基础操作
	Create(ctx context.Context, relationship *model.UserRelationship) error
	Update(ctx context.Context, relationship *model.UserRelationship) error
	Delete(ctx context.Context, followerID, followingID uint64) error

	// 查询操作
	GetRelationship(ctx context.Context, followerID, followingID uint64) (*model.UserRelationship, error)
	GetFollowers(ctx context.Context, userID uint64, status string, offset, limit int) ([]*model.UserRelationship, int64, error)
	GetFollowings(ctx context.Context, userID uint64, status string, offset, limit int) ([]*model.UserRelationship, int64, error)

	// 状态操作
	UpdateStatus(ctx context.Context, followerID, followingID uint64, status string) error
	ExistsRelationship(ctx context.Context, followerID, followingID uint64) (bool, error)
}

// TagRepository 标签仓储接口
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	Update(ctx context.Context, tag *model.Tag) error
	GetByID(ctx context.Context, id uint64) (*model.Tag, error)
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	List(ctx context.Context, offset, limit int) ([]*model.Tag, int64, error)
	ListPopular(ctx context.Context, limit int) ([]*model.Tag, error)
	IncrementUseCount(ctx context.Context, id uint64) error
	DecrementUseCount(ctx context.Context, id uint64) error
	Delete(ctx context.Context, id uint64) error
	BatchCreate(ctx context.Context, tags []string) ([]uint64, error)
}
