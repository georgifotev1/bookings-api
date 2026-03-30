package repository

import (
	"context"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/internal/cache"
	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	Create(ctx context.Context, tenant *domain.Tenant) error
	Update(ctx context.Context, tenant *domain.Tenant) error
	GetWorkingHours(ctx context.Context, tenantID string) ([]domain.WorkingHours, error)
	UpsertWorkingHours(ctx context.Context, tenantID string, hours []domain.WorkingHours) error
}

type tenantRepository struct {
	pool  *pgxpool.Pool
	cache *cache.TenantStore
}

func NewTenantRepository(pool *pgxpool.Pool, cache *cache.TenantStore) TenantRepository {
	return &tenantRepository{
		pool:  pool,
		cache: cache,
	}
}

func (r *tenantRepository) dbFromContext(ctx context.Context) txmanager.DBTX {
	if tx, ok := txmanager.TxFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *tenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	tenant, err := r.cache.Get(ctx, id)
	if err != nil {
		fmt.Printf("cache get failed for tenant:%s: %v\n", id, err)
	}
	if tenant != nil {
		return tenant, nil
	}

	query := `
		SELECT id, name, slug, phone, email, tier, timezone, address_id, created_at, updated_at
		FROM tenants WHERE id = $1`

	var t domain.Tenant
	row := r.dbFromContext(ctx).QueryRow(ctx, query, id)
	err = row.Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&t.Phone,
		&t.Email,
		&t.Tier,
		&t.Timezone,
		&t.AddressID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("tenantRepository.GetByID: %w", apperr.NotFound("tenant not found", err))
		}
		return nil, fmt.Errorf("tenantRepository.GetByID: %w", apperr.Internal(err))
	}

	if err := r.cache.Set(ctx, id, &t); err != nil {
		fmt.Printf("cache set failed for tenant:%s: %v\n", id, err)
	}

	return &t, nil
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, phone, email, tier, timezone, address_id, created_at, updated_at
		FROM tenants WHERE slug = $1`

	var t domain.Tenant
	row := r.dbFromContext(ctx).QueryRow(ctx, query, slug)
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&t.Phone,
		&t.Email,
		&t.Tier,
		&t.Timezone,
		&t.AddressID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("tenantRepository.GetBySlug: %w", apperr.NotFound("tenant not found", err))
		}
		return nil, fmt.Errorf("tenantRepository.GetBySlug: %w", apperr.Internal(err))
	}

	return &t, nil
}

func (r *tenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, phone, email, tier, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.dbFromContext(ctx).Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Phone,
		tenant.Email,
		tenant.Tier,
		tenant.Timezone,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("tenantRepository.Create: %w", apperr.Conflict("tenant already exists"))
		}
		return fmt.Errorf("tenantRepository.Create: %w", apperr.Internal(err))
	}

	return nil
}

func (r *tenantRepository) GetWorkingHours(ctx context.Context, tenantID string) ([]domain.WorkingHours, error) {
	hours, err := r.cache.WorkingHours.Get(ctx, tenantID)
	if err != nil {
		fmt.Printf("cache get failed for working_hours:%s: %v\n", tenantID, err)
	}
	if hours != nil {
		return *hours, nil
	}

	query := `
		SELECT id, tenant_id, day_of_week, opens_at, closes_at, is_closed
		FROM working_hours WHERE tenant_id = $1
		ORDER BY day_of_week`

	rows, err := r.dbFromContext(ctx).Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenantRepository.GetWorkingHours: %w", apperr.Internal(err))
	}
	defer rows.Close()

	var workingHours []domain.WorkingHours
	for rows.Next() {
		var h domain.WorkingHours
		if err := rows.Scan(&h.ID, &h.TenantID, &h.DayOfWeek, &h.OpensAt, &h.ClosesAt, &h.IsClosed); err != nil {
			return nil, fmt.Errorf("tenantRepository.GetWorkingHours scan: %w", apperr.Internal(err))
		}
		workingHours = append(workingHours, h)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("tenantRepository.GetWorkingHours: %w", apperr.Internal(rows.Err()))
	}

	if err := r.cache.WorkingHours.Set(ctx, tenantID, &workingHours); err != nil {
		fmt.Printf("cache set failed for working_hours:%s: %v\n", tenantID, err)
	}

	return workingHours, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		UPDATE tenants 
		SET name = $2, slug = $3, phone = $4, email = $5, tier = $6, timezone = $7, address_id = $8, updated_at = $9
		WHERE id = $1`

	result, err := r.dbFromContext(ctx).Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Phone,
		tenant.Email,
		tenant.Tier,
		tenant.Timezone,
		tenant.AddressID,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("tenantRepository.Update: %w", apperr.Internal(err))
	}

	if result.RowsAffected() == 0 {
		return apperr.NotFound("tenant not found", nil)
	}

	if err := r.cache.Delete(ctx, tenant.ID); err != nil {
		fmt.Printf("cache delete failed for tenant:%s: %v\n", tenant.ID, err)
	}

	if err := r.cache.WorkingHours.Delete(ctx, tenant.ID); err != nil {
		fmt.Printf("cache delete failed for working_hours:%s: %v\n", tenant.ID, err)
	}

	return nil
}

func (r *tenantRepository) UpsertWorkingHours(ctx context.Context, tenantID string, hours []domain.WorkingHours) error {
	if len(hours) == 0 {
		return nil
	}

	query := `
        INSERT INTO working_hours (id, tenant_id, day_of_week, opens_at, closes_at, is_closed)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (tenant_id, day_of_week) DO UPDATE SET
            opens_at  = EXCLUDED.opens_at,
            closes_at = EXCLUDED.closes_at,
            is_closed = EXCLUDED.is_closed`

	for _, h := range hours {
		_, err := r.dbFromContext(ctx).Exec(ctx, query,
			h.ID, tenantID, h.DayOfWeek, h.OpensAt, h.ClosesAt, h.IsClosed,
		)
		if err != nil {
			return fmt.Errorf("tenantRepository.UpsertWorkingHours: %w", apperr.Internal(err))
		}
	}

	if err := r.cache.WorkingHours.Delete(ctx, tenantID); err != nil {
		fmt.Printf("cache delete failed for working_hours:%s: %v\n", tenantID, err)
	}

	return nil
}
