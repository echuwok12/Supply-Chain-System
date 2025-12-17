package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}

	log.Println("✅ Redis connection established")
	return client
}
