package controller

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Rdb       *redis.Client
	ctx       = context.Background()
	redisAddr = os.Getenv("REDIS_ADDR")
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func SetCache(key string, value string, ttl time.Duration) error {
	log.Printf("Setting cache for key: %s", key)
	return Rdb.Set(ctx, key, value, ttl).Err()
}

func GetCache(key string) (string, error) {
	res, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Cache miss for key %s: %v", key, err)
		return "", err
	}
	return res, nil
}

func DeleteCache(key string) error {
	log.Printf("Deleting cache for key: %s", key)
	return Rdb.Del(ctx, key).Err()
}

func RunCommand(cmd string) (string, error) {
	result, err := Rdb.Do(ctx, cmd).Result()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", result), nil
}
