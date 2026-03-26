-- +goose Up
ALTER TABLE tenants ADD COLUMN tier VARCHAR(20) NOT NULL DEFAULT 'base' CHECK(tier IN ('base', 'pro'));

ALTER TABLE user_invitations ADD COLUMN name VARCHAR(100);
ALTER TABLE user_invitations ADD COLUMN phone VARCHAR(20);

CREATE INDEX idx_tenants_tier ON tenants(tier);
CREATE INDEX idx_invitations_tenant_id ON user_invitations(tenant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_invitations_tenant_id;
DROP INDEX IF EXISTS idx_tenants_tier;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS phone;
ALTER TABLE user_invitations DROP COLUMN IF EXISTS name;
ALTER TABLE tenants DROP COLUMN IF EXISTS tier;
