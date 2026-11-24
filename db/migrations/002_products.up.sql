-- 002_products.up.sql
CREATE TABLE IF NOT EXISTS products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sku VARCHAR(64) NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT NULL,
    category VARCHAR(80) NOT NULL,
    cost_price NUMERIC(15,2) NOT NULL DEFAULT 0.00 CHECK (cost_price >= 0),
    selling_price NUMERIC(15,2) NOT NULL DEFAULT 0.00 CHECK (selling_price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    unit VARCHAR(30) NOT NULL DEFAULT 'PCS',
    image_url VARCHAR(500) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'OUT_OF_STOCK')),
    created_by BIGINT NULL REFERENCES admins(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_products_sku_active ON products (sku) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_products_status_id ON products (status, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products (deleted_at);
