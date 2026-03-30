-- +goose Up
ALTER TABLE refresh_tokens ADD COLUMN customer_id VARCHAR(27) REFERENCES customers(id) ON DELETE CASCADE;
ALTER TABLE refresh_tokens ADD COLUMN tenant_id VARCHAR(27) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX idx_refresh_tokens_customer_id ON refresh_tokens(customer_id);
CREATE INDEX idx_refresh_tokens_tenant_id ON refresh_tokens(tenant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_refresh_tokens_customer_id;
DROP INDEX IF EXISTS idx_refresh_tokens_tenant_id;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS customer_id;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS tenant_id;
