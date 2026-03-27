-- +goose Up
ALTER TABLE tenants ADD COLUMN timezone VARCHAR(50) NOT NULL DEFAULT 'UTC';

-- +goose Down
ALTER TABLE tenants DROP COLUMN IF EXISTS timezone;
