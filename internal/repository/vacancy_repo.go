package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type VacancyRepo interface {
	IsDuplicate(ctx context.Context, vacancyID string) (bool, error)
}

type VacancyRepoImpl struct {
	redisClient *redis.Client
	key         string
}

func NewVacancyRepo(redisClient *redis.Client, key string) *VacancyRepoImpl {
	return &VacancyRepoImpl{redisClient: redisClient, key: key}
}

func (v *VacancyRepoImpl) IsDuplicate(ctx context.Context, vacancyID string) (bool, error) {
	rdb := v.redisClient
	//Если такая вакансия уже есть, не добавим
	sAdd, err := rdb.SAdd(ctx, v.key, vacancyID).Result()
	if err != nil {
		return false, err
	}

	if sAdd == 0 {
		return true, nil
	}

	return false, nil
}
