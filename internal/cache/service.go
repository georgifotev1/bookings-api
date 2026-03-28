package cache

import (
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ServiceStore struct {
	*BaseStore[domain.Service]
}

func NewServiceStore(rdb *redis.Client) *ServiceStore {
	return &ServiceStore{
		BaseStore: NewBaseStore[domain.Service](rdb, "service", 5*time.Minute),
	}
}
