// repository.go

package repository

import (
	"context"

	"github.com/chiliososada/distance-back/internal/model"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// 基础操作
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, uid string) error
	GetByUID(ctx context.Context, uid string) (*model.User, error)

	// 认证相关
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	CreateAuthentication(ctx context.Context, auth *model.UserAuthentication) error
	UpdateAuthentication(ctx context.Context, auth *model.UserAuthentication) error
	CreateWithAuth(ctx context.Context, user *model.User, auth *model.UserAuthentication) error
	UpdateWithAuth(ctx context.Context, user *model.User, auth *model.UserAuthentication) error

	// 设备相关
	CreateDevice(ctx context.Context, device *model.UserDevice) error
	UpdateDevice(ctx context.Context, device *model.UserDevice) error
	GetDeviceByToken(ctx context.Context, token string) (*model.UserDevice, error)
	GetUserDevices(ctx context.Context, userUID string) ([]*model.UserDevice, error)

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]*model.User, int64, error)
	Search(ctx context.Context, keyword string, offset, limit int) ([]*model.User, int64, error)
	GetNearbyUsers(ctx context.Context, lat, lng float64, radius float64, offset, limit int) ([]*model.User, int64, error)

	// 状态操作
	UpdateStatus(ctx context.Context, userUID string, status string) error
	UpdateLastActive(ctx context.Context, userUID string) error
}

// TopicRepository 话题仓储接口
type TopicRepository interface {
	// 基础操作
	Create(ctx context.Context, topic *model.Topic) error
	Update(ctx context.Context, topic *model.Topic) error
	Delete(ctx context.Context, uid string) error
	GetByUID(ctx context.Context, uid string) (*model.Topic, error)

	// 图片相关
	AddImages(ctx context.Context, topicUID string, images []*model.TopicImage) error
	GetImages(ctx context.Context, topicUID string) ([]*model.TopicImage, error)
	DeleteTopicImages(ctx context.Context, topicUID string, imageUIDs []string) error
	// 标签相关
	AddTags(ctx context.Context, topicUID string, tagUIDs []string) error
	RemoveTags(ctx context.Context, topicUID string, tagUIDs []string) error
	GetTags(ctx context.Context, topicUID string) ([]*model.Tag, error)
	BatchCreate(ctx context.Context, tags []string) ([]string, error)
	ListPopular(ctx context.Context, limit int) ([]*model.Tag, error)

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]*model.Topic, int64, error)
	ListByUser(ctx context.Context, userUID string, offset, limit int) ([]*model.Topic, int64, error)
	ListByTag(ctx context.Context, tagUID string, offset, limit int) ([]*model.Topic, int64, error)
	GetNearbyTopics(ctx context.Context, lat, lng float64, radius float64, offset, limit int) ([]*model.Topic, int64, error)

	// 互动操作
	AddInteraction(ctx context.Context, interaction *model.TopicInteraction) error
	RemoveInteraction(ctx context.Context, topicUID, userUID string, interactionType string) error
	GetInteractions(ctx context.Context, topicUID string, interactionType string) ([]*model.TopicInteraction, error)

	// 计数操作
	IncrementViewCount(ctx context.Context, topicUID string) error
	UpdateCounts(ctx context.Context, topicUID string) error
}

// ChatRepository 聊天仓储接口
type ChatRepository interface {

	// 聊天室操作
	CreateRoom(ctx context.Context, room *model.ChatRoom) error
	UpdateRoom(ctx context.Context, room *model.ChatRoom) error
	GetRoomByUID(ctx context.Context, uid string) (*model.ChatRoom, error)
	ListUserRooms(ctx context.Context, userUID string, offset, limit int) ([]*model.ChatRoom, int64, error)
	FindPrivateRoom(ctx context.Context, userUID1, userUID2 string) (*model.ChatRoom, error)
	// 成员操作
	AddMember(ctx context.Context, member *model.ChatRoomMember) error
	RemoveMember(ctx context.Context, roomUID, userUID string) error
	UpdateMember(ctx context.Context, member *model.ChatRoomMember) error
	GetRoomMembers(ctx context.Context, roomUID string) ([]*model.ChatRoomMember, error)

	// 消息操作
	CreateMessage(ctx context.Context, message *model.Message) error
	GetMessagesByRoom(ctx context.Context, roomUID string, beforeUID string, limit int) ([]*model.Message, error)
	GetLatestMessages(ctx context.Context, roomUID string, limit int) ([]*model.Message, error)

	// 媒体操作
	AddMessageMedia(ctx context.Context, media *model.MessageMedia) error
	GetMessageMedia(ctx context.Context, messageUID string) ([]*model.MessageMedia, error)

	// 置顶操作
	PinRoom(ctx context.Context, userUID, roomUID string) error
	UnpinRoom(ctx context.Context, userUID, roomUID string) error
	GetPinnedRooms(ctx context.Context, userUID string) ([]*model.ChatRoom, error)
}

// RelationshipRepository 关系仓储接口
type RelationshipRepository interface {
	// 基础操作
	Create(ctx context.Context, relationship *model.UserRelationship) error
	Update(ctx context.Context, relationship *model.UserRelationship) error
	Delete(ctx context.Context, followerUID, followingUID string) error
	AcceptFollow(ctx context.Context, relationship *model.UserRelationship) error
	// 查询操作
	GetRelationship(ctx context.Context, followerUID, followingUID string) (*model.UserRelationship, error)
	GetFollowers(ctx context.Context, userUID string, status string, offset, limit int) ([]*model.UserRelationship, int64, error)
	GetFollowings(ctx context.Context, userUID string, status string, offset, limit int) ([]*model.UserRelationship, int64, error)

	// 状态操作
	UpdateStatus(ctx context.Context, followerUID, followingUID string, status string) error
	ExistsRelationship(ctx context.Context, followerUID, followingUID string) (bool, error)
}

// TagRepository 标签仓储接口
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	Update(ctx context.Context, tag *model.Tag) error
	GetByUID(ctx context.Context, uid string) (*model.Tag, error)
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	List(ctx context.Context, offset, limit int) ([]*model.Tag, int64, error)
	ListPopular(ctx context.Context, limit int) ([]*model.Tag, error)
	IncrementUseCount(ctx context.Context, uid string) error
	DecrementUseCount(ctx context.Context, uid string) error
	Delete(ctx context.Context, uid string) error
	BatchCreate(ctx context.Context, tags []string) ([]string, error)
}
