package redis

import (
	"context"
	"im-chat/internal/infra/config"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Cache *redisCache

func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     100,
		MinIdleConns: 20,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return err
	}

	Cache = newRedisCache(RDB)
	log.Println("Redis connected successfully")
	return nil
}

func GetRedis() *redis.Client {
	return RDB
}

func GetCache() *redisCache {
	return Cache
}

func Set(key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

func Del(key string) error {
	return RDB.Del(ctx, key).Err()
}
