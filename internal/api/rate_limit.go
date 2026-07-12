package api

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Wait(ctx context.Context) error
}

type RateLimiterImpl struct {
	redisClient *redis.Client
	key         string
	limit       int
	period      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, key string, limit int, period time.Duration) *RateLimiterImpl {
	return &RateLimiterImpl{redisClient: redisClient, key: key, limit: limit, period: period}
}

func (r *RateLimiterImpl) Wait(ctx context.Context) error {
	for {

		rdb := r.redisClient

		val, err := rdb.Incr(ctx, r.key).Result()
		if err != nil {
			return err
		}

		if val == 1 {
			_, err = rdb.Expire(ctx, r.key, r.period).Result()
			if err != nil {
				return err
			}
		}

		if val > int64(r.limit) {
			ttl, err := rdb.TTL(ctx, r.key).Result()
			if err != nil {
				return err
			}
			select {
			case <-time.After(ttl):
				// окно истекло, пробуем заново
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	}

}
