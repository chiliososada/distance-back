package cache

import (
	"fmt"
	"time"
)

const (
	// 过期时间相关常量
	DefaultExpiration = time.Hour * 24
	ShortExpiration   = time.Minute * 30
	LongExpiration    = time.Hour * 24 * 7

	// 用户相关前缀
	UserKeyPrefix     = "user:"
	UserTokenPrefix   = "user:token:"
	UserProfilePrefix = "user:profile:"
	UserOnlinePrefix  = "user:online:"

	// 话题相关前缀
	TopicKeyPrefix  = "topic:"
	TopicLikePrefix = "topic:like:"
	TopicViewPrefix = "topic:view:"

	// 聊天相关前缀
	ChatRoomPrefix     = "chat:room:"
	ChatMembersPrefix  = "chat:members:"
	ChatMessagesPrefix = "chat:messages:"

	// 位置相关前缀
	LocationKeyPrefix = "location:"
	NearbyKeyPrefix   = "nearby:"

	// 标签相关前缀
	TagKeyPrefix    = "tag:"
	TopicTagsPrefix = "topic:tags:"
	PopularTagsName = "popular:tags" // 改名以避免与函数冲突
)

// 用户相关键生成函数
func UserKey(userID uint64) string {
	return fmt.Sprintf("%s%d", UserKeyPrefix, userID)
}

func UserTokenKey(userID uint64) string {
	return fmt.Sprintf("%s%d", UserTokenPrefix, userID)
}

func UserProfileKey(userID uint64) string {
	return fmt.Sprintf("%s%d", UserProfilePrefix, userID)
}

func UserOnlineKey(userID uint64) string {
	return fmt.Sprintf("%s%d", UserOnlinePrefix, userID)
}

// 话题相关键生成函数
func TopicKey(topicID uint64) string {
	return fmt.Sprintf("%s%d", TopicKeyPrefix, topicID)
}

func TopicLikeKey(topicID uint64) string {
	return fmt.Sprintf("%s%d", TopicLikePrefix, topicID)
}

func TopicViewKey(topicID uint64) string {
	return fmt.Sprintf("%s%d", TopicViewPrefix, topicID)
}

// 聊天相关键生成函数
func ChatRoomKey(roomID uint64) string {
	return fmt.Sprintf("%s%d", ChatRoomPrefix, roomID)
}

func ChatMembersKey(roomID uint64) string {
	return fmt.Sprintf("%s%d", ChatMembersPrefix, roomID)
}

func ChatMessagesKey(roomID uint64) string {
	return fmt.Sprintf("%s%d", ChatMessagesPrefix, roomID)
}

// 位置相关键生成函数
func LocationKey(userID uint64) string {
	return fmt.Sprintf("%s%d", LocationKeyPrefix, userID)
}

func NearbyKey(latitude, longitude float64) string {
	return fmt.Sprintf("%s%.6f:%.6f", NearbyKeyPrefix, latitude, longitude)
}

// 标签相关键生成函数
func TagKey(tagID uint64) string {
	return fmt.Sprintf("%s%d", TagKeyPrefix, tagID)
}

func TopicTagsKey(topicID uint64) string {
	return fmt.Sprintf("%s%d", TopicTagsPrefix, topicID)
}

func PopularTagsKey() string {
	return PopularTagsName
}

// 缓存删除函数
func RemoveUserCache(userID uint64) error {
	keys := []string{
		UserKey(userID),
		UserTokenKey(userID),
		UserProfileKey(userID),
		UserOnlineKey(userID),
	}

	for _, key := range keys {
		if err := Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func RemoveTopicCache(topicID uint64) error {
	keys := []string{
		TopicKey(topicID),
		TopicLikeKey(topicID),
		TopicViewKey(topicID),
	}

	for _, key := range keys {
		if err := Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func RemoveChatRoomCache(roomID uint64) error {
	keys := []string{
		ChatRoomKey(roomID),
		ChatMembersKey(roomID),
		ChatMessagesKey(roomID),
	}

	for _, key := range keys {
		if err := Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func RemoveLocationCache(userID uint64) error {
	return Delete(LocationKey(userID))
}
