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
	PopularTagsName = "popular:tags"
)

// 用户相关键生成函数
func UserKey(uid string) string {
	return fmt.Sprintf("%s%s", UserKeyPrefix, uid)
}

func UserTokenKey(uid string) string {
	return fmt.Sprintf("%s%s", UserTokenPrefix, uid)
}

func UserProfileKey(uid string) string {
	return fmt.Sprintf("%s%s", UserProfilePrefix, uid)
}

func UserOnlineKey(uid string) string {
	return fmt.Sprintf("%s%s", UserOnlinePrefix, uid)
}

// 话题相关键生成函数
func TopicKey(uid string) string {
	return fmt.Sprintf("%s%s", TopicKeyPrefix, uid)
}

func TopicLikeKey(uid string) string {
	return fmt.Sprintf("%s%s", TopicLikePrefix, uid)
}

func TopicViewKey(uid string) string {
	return fmt.Sprintf("%s%s", TopicViewPrefix, uid)
}

// 聊天相关键生成函数
func ChatRoomKey(uid string) string {
	return fmt.Sprintf("%s%s", ChatRoomPrefix, uid)
}

func ChatMembersKey(uid string) string {
	return fmt.Sprintf("%s%s", ChatMembersPrefix, uid)
}

func ChatMessagesKey(uid string) string {
	return fmt.Sprintf("%s%s", ChatMessagesPrefix, uid)
}

// 位置相关键生成函数
func LocationKey(uid string) string {
	return fmt.Sprintf("%s%s", LocationKeyPrefix, uid)
}

func NearbyKey(latitude, longitude float64) string {
	return fmt.Sprintf("%s%.6f:%.6f", NearbyKeyPrefix, latitude, longitude)
}

// 标签相关键生成函数
func TagKey(uid string) string {
	return fmt.Sprintf("%s%s", TagKeyPrefix, uid)
}

func TopicTagsKey(uid string) string {
	return fmt.Sprintf("%s%s", TopicTagsPrefix, uid)
}

func PopularTagsKey() string {
	return PopularTagsName
}

// 缓存删除函数
func RemoveUserCache(uid string) error {
	keys := []string{
		UserKey(uid),
		UserTokenKey(uid),
		UserProfileKey(uid),
		UserOnlineKey(uid),
	}

	for _, key := range keys {
		if err := Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func RemoveTopicCache(uid string) error {
	keys := []string{
		TopicKey(uid),
		TopicLikeKey(uid),
		TopicViewKey(uid),
	}

	for _, key := range keys {
		if err := Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func RemoveChatRoomCache(uid string) error {
	keys := []string{
		ChatRoomKey(uid),
		ChatMembersKey(uid),
		ChatMessagesKey(uid),
	}

	for _, key := range keys {
		if err := Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func RemoveLocationCache(uid string) error {
	return Delete(LocationKey(uid))
}
