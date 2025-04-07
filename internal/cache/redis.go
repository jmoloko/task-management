package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoloko/taskmange/internal/domain/repository"
	"github.com/redis/go-redis/v9"
)

const (
	// Формат ключа: analytics:{userID}:{period}
	analyticsKeyFormat = "analytics:%s:%s"
	analyticsTTL       = 6 * time.Hour
)

type RedisCache struct {
	client *redis.Client
}

// создание нового экземпляра кэша Redis
func NewRedisCache(client *redis.Client) repository.AnalyticsCache {
	return &RedisCache{client: client}
}

// извлечение аналитических данных для определенного пользователя и периода из Redis
func (c *RedisCache) GetUserAnalytics(ctx context.Context, userID, period string) (*repository.CachedAnalytics, error) {
	key := fmt.Sprintf(analyticsKeyFormat, userID, period)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get analytics from cache: %w", err)
	}

	var analytics repository.CachedAnalytics
	if err := json.Unmarshal(data, &analytics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analytics data: %w", err)
	}

	return &analytics, nil
}

// хранение аналитических данных для определенного пользователя и периода в Redis.
func (c *RedisCache) SetUserAnalytics(ctx context.Context, analytics repository.CachedAnalytics) error {
	key := fmt.Sprintf(analyticsKeyFormat, analytics.UserID, analytics.Period)

	data, err := json.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics data: %w", err)
	}

	if err := c.client.Set(ctx, key, data, analyticsTTL).Err(); err != nil {
		return fmt.Errorf("failed to set analytics in cache: %w", err)
	}

	return nil
}

// удаление аналитическич данных для определенного пользователя из Redis
func (c *RedisCache) InvalidateUserAnalytics(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf(analyticsKeyFormat, userID, "*")

	// Находим все ключи для данного пользователя
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete analytics key %s: %w", iter.Val(), err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan analytics keys: %w", err)
	}

	return nil
}
