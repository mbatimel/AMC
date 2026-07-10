package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	customerrors "github.com/mbatimel/AMC/auth/internal/errors"
)

type ICacheRepository interface {
	SaveVerificationCode(ctx context.Context, key, code string, ttl time.Duration) error
	GetVerificationCode(ctx context.Context, key string) (string, error)
	DeleteVerificationCode(ctx context.Context, key string) error
}

type CacheRepo struct {
	client *redis.Client
	logger zerolog.Logger
}

func NewCacheRepo(addr, password string, db int, logger zerolog.Logger) *CacheRepo {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &CacheRepo{client: rdb, logger: logger}
}

func (r *CacheRepo) SaveVerificationCode(ctx context.Context, key, code string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, code, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (r *CacheRepo) GetVerificationCode(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", customerrors.ErrCodeExpiredOrNotFound()
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (r *CacheRepo) DeleteVerificationCode(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}
