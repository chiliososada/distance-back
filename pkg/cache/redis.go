package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"DistanceBack_v1/config"
	"DistanceBack_v1/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  4 * time.Second,
	})

	// 测试连接
	if err := RedisClient.Ping(Ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %v", err)
	}

	logger.Info("Redis connected successfully")
	return nil
}

// Close 关闭Redis连接
func Close() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// Set 设置缓存
func Set(key string, value interface{}, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %v", err)
	}

	err = RedisClient.Set(Ctx, key, bytes, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %v", err)
	}

	return nil
}

// Get 获取缓存
func Get(key string, value interface{}) error {
	bytes, err := RedisClient.Get(Ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to get cache: %v", err)
	}

	err = json.Unmarshal(bytes, value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %v", err)
	}

	return nil
}

// Delete 删除缓存
func Delete(key string) error {
	return RedisClient.Del(Ctx, key).Err()
}

// SetNX 设置缓存（如果key不存在）
func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cache value: %v", err)
	}

	return RedisClient.SetNX(Ctx, key, bytes, expiration).Result()
}

// Exists 检查key是否存在
func Exists(key string) (bool, error) {
	n, err := RedisClient.Exists(Ctx, key).Result()
	return n > 0, err
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return RedisClient.Expire(Ctx, key, expiration).Err()
}

// Keys 获取所有匹配的key
func Keys(pattern string) ([]string, error) {
	return RedisClient.Keys(Ctx, pattern).Result()
}

// HSet 设置哈希表字段
func HSet(key string, field string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %v", err)
	}

	return RedisClient.HSet(Ctx, key, field, bytes).Err()
}

// HGet 获取哈希表字段
func HGet(key string, field string, value interface{}) error {
	bytes, err := RedisClient.HGet(Ctx, key, field).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to get hash field: %v", err)
	}

	return json.Unmarshal(bytes, value)
}

// HDel 删除哈希表字段
func HDel(key string, fields ...string) error {
	return RedisClient.HDel(Ctx, key, fields...).Err()
}

// Lock 分布式锁
type Lock struct {
	key        string
	value      string
	expiration time.Duration
}

// NewLock 创建一个新的分布式锁
func NewLock(key string, expiration time.Duration) *Lock {
	return &Lock{
		key:        fmt.Sprintf("lock:%s", key),
		value:      fmt.Sprintf("%d", time.Now().UnixNano()),
		expiration: expiration,
	}
}

// Lock 获取锁
func (l *Lock) Lock() (bool, error) {
	return RedisClient.SetNX(Ctx, l.key, l.value, l.expiration).Result()
}

// Unlock 释放锁
func (l *Lock) Unlock() error {
	// 使用Lua脚本确保原子性
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := RedisClient.Eval(Ctx, script, []string{l.key}, l.value).Result()
	return err
}
