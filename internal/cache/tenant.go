package cache

import (
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type TenantStore struct {
	*BaseStore[domain.Tenant]
	WorkingHours *WorkingHoursStore
}

type WorkingHoursStore struct {
	*BaseStore[[]domain.WorkingHours]
}

func NewWorkingHoursStore(rdb *redis.Client) *WorkingHoursStore {
	return &WorkingHoursStore{
		BaseStore: NewBaseStore[[]domain.WorkingHours](rdb, "working_hours", 5*time.Minute),
	}
}

func NewTenantStore(rdb *redis.Client) *TenantStore {
	workingHoursStore := NewWorkingHoursStore(rdb)
	return &TenantStore{
		BaseStore:    NewBaseStore[domain.Tenant](rdb, "tenant", 5*time.Minute),
		WorkingHours: workingHoursStore,
	}
}
