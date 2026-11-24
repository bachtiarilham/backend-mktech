-- 003_cashflows.up.sql
CREATE TABLE IF NOT EXISTS cashflows (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('PROFIT', 'LOSS', 'SAVING')),
    category VARCHAR(80) NOT NULL,
    title VARCHAR(150) NOT NULL,
    description TEXT NULL,
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    product_id BIGINT NULL REFERENCES products(id) ON DELETE SET NULL,
    transaction_date DATE NOT NULL,
    created_by BIGINT NULL REFERENCES admins(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_cashflows_type_date ON cashflows (type, transaction_date DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_date_id ON cashflows (transaction_date DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_product_id ON cashflows (product_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_deleted_at ON cashflows (deleted_at);
