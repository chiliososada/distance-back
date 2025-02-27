package repository

import (
	"context"
	"strconv"

	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chiliososada/distance-back/internal/model"
	"github.com/chiliososada/distance-back/pkg/cache"
	"github.com/chiliososada/distance-back/pkg/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
)

const (
	CheckDBError = "check db"
)

type TopicCache interface {
	ReLoad(ctx context.Context) error
	Get(c *gin.Context, args ...interface{}) ([]*model.CachedTopic, int, error)
}

const (
	zsetKey   = "recent-topic-zset"
	loadLimit = 1000
)

type RedisRecentTopic struct {
	mu     sync.RWMutex
	client *redis.Client
	db     *gorm.DB
}

func NewRedisRecentTopic(ctx context.Context) (*RedisRecentTopic, error) {
	r := &RedisRecentTopic{
		client: cache.RedisClient,
		db:     database.GetDB(),
	}

	if err := r.ReLoad(ctx); err != nil {
		return nil, err
	} else {
		return r, nil
	}
}
func (r *RedisRecentTopic) recentTopics(ctx context.Context) ([]*model.Topic, error) {
	var topics []*model.Topic

	result := r.db.WithContext(ctx).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("users.uid, users.nickname, users.avatar_url, users.gender, users.location_latitude, users.location_longitude")
	}).Preload("TopicImages").Preload("Tags").Preload("ChatRoom", func(db *gorm.DB) *gorm.DB {
		return db.Select("chat_rooms.uid, chat_rooms.topic_uid")
	}).
		Where("topics.expires_at > ? AND topics.status = ?", time.Now(), "active").
		Order("topics.updated_at DESC").
		Limit(loadLimit).Find(&topics)

	if result.Error != nil {
		return nil, result.Error
	} else {
		return topics, nil
	}
}

func (r *RedisRecentTopic) ReLoad(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	topics, err := r.recentTopics(ctx)
	if err != nil {
		return nil
	}

	var members []redis.Z
	for _, t := range topics {
		value, err := t.MarshalToCached()
		if err != nil {
			continue
		} else if err := r.client.Set(ctx, fmt.Sprintf("topic:%s", t.UID), value, time.Until(t.ExpiresAt)).Err(); err != nil {
			continue
		} else {
			members = append(members, redis.Z{
				Score:  float64(t.UpdatedAt.Unix()),
				Member: t.UID,
			})
		}
	}
	if err := r.client.Del(ctx, zsetKey).Err(); err != nil {
		return err
	}
	if len(members) > 0 {
		return r.client.ZAdd(ctx, zsetKey, members...).Err()
	} else {
		return nil
	}

}

func (r *RedisRecentTopic) Get(c *gin.Context, args ...interface{}) ([]*model.CachedTopic, int, error) {
	if len(args) != 2 {
		return nil, 0, errors.New("invalid arguments")
	}

	count, ok := args[0].(int)
	if !ok {
		return nil, 0, errors.New("invalid arguments")
	}

	recencyScore, ok := args[1].(int)
	if !ok {
		return nil, 0, errors.New("invalid arguments")
	}

	fmt.Printf("count: %v, recencyScore: %v\n", count, recencyScore)

	if count == 0 {
		return nil, recencyScore, nil
	}
	ctx := c.Request.Context()
	if r.mu.TryRLock() {
		defer r.mu.RUnlock()
		//fastpath

		//fmt.Printf("recency: %v, count: %v\n", recencyScore, count)
		// we try to read 1 more to get the next cursor
		max := "+inf"
		if recencyScore != 0 {
			max = "(" + strconv.Itoa(recencyScore)
		}
		result, err := r.client.ZRevRangeByScoreWithScores(ctx, zsetKey, &redis.ZRangeBy{
			Min:   "0",
			Max:   max,
			Count: int64(count)}).Result()
		if err != nil {
			return nil, recencyScore, err
		}

		var cachedTopics []*model.CachedTopic
		var topicKeys []string
		for _, member := range result {
			topicKeys = append(topicKeys, fmt.Sprintf("topic:%s", member.Member))
		}

		if len(result) == 0 {
			goto SLOWPATH
		} else {
			values, err := r.client.MGet(ctx, topicKeys...).Result()
			if err != nil {
				return nil, recencyScore, err
			}
			for _, raw := range values {
				if raw == nil {
					continue
				}
				var ct model.CachedTopic
				if json.Unmarshal([]byte(raw.(string)), &ct) == nil {
					cachedTopics = append(cachedTopics, &ct)
				}
			}
			updatedScore := result[len(result)-1].Score

			return cachedTopics, int(updatedScore), nil

		}

	}

SLOWPATH:

	// slowpath
	// let the caller check db
	return nil, recencyScore, errors.New(CheckDBError)

}
