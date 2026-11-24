-- name: CashflowCreate
-- params: type text, category text, title text, description text, amount numeric, product_id bigint nullable, transaction_date date, created_by bigint
INSERT INTO cashflows (
    type, category, title, description, amount, product_id, transaction_date, created_by
) VALUES (
    $1, $2, $3, $4, $5, NULLIF($6, 0), $7::date, $8
) RETURNING jsonb_build_object(
    'id', id, 'type', type, 'category', category, 'title', title,
    'description', description, 'amount', amount, 'product_id', product_id,
    'transaction_date', transaction_date, 'created_by', created_by,
    'created_at', created_at, 'updated_at', updated_at
);

-- name: CashflowGetByID
-- params: id bigint
SELECT jsonb_build_object(
    'id', c.id, 'type', c.type, 'category', c.category, 'title', c.title,
    'description', c.description, 'amount', c.amount, 'product_id', c.product_id,
    'product_name', p.name,
    'transaction_date', c.transaction_date, 'created_by', c.created_by,
    'created_at', c.created_at, 'updated_at', c.updated_at
) FROM cashflows c
LEFT JOIN products p ON p.id = c.product_id
WHERE c.id = $1 AND c.deleted_at IS NULL;

-- name: CashflowList
-- params: cursor_id bigint, limit int, type text, category text, product_id bigint, start_date text, end_date text
SELECT jsonb_build_object(
    'id', c.id, 'type', c.type, 'category', c.category, 'title', c.title,
    'description', c.description, 'amount', c.amount, 'product_id', c.product_id,
    'product_name', p.name,
    'transaction_date', c.transaction_date, 'created_by', c.created_by,
    'created_at', c.created_at, 'updated_at', c.updated_at
) FROM cashflows c
LEFT JOIN products p ON p.id = c.product_id
WHERE c.deleted_at IS NULL
  AND ($1 = 0 OR c.id < $1)
  AND ($3 = '' OR c.type = $3)
  AND ($4 = '' OR c.category = $4)
  AND ($5 = 0 OR c.product_id = $5)
  AND ($6 = '' OR c.transaction_date >= $6::date)
  AND ($7 = '' OR c.transaction_date <= $7::date)
ORDER BY c.id DESC
LIMIT $2;

-- name: CashflowUpdate
-- params: id bigint, type text, category text, title text, description text, amount numeric, product_id bigint nullable, transaction_date date
UPDATE cashflows SET
    type = $2,
    category = $3,
    title = $4,
    description = $5,
    amount = $6,
    product_id = NULLIF($7, 0),
    transaction_date = $8::date,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING jsonb_build_object(
    'id', id, 'type', type, 'category', category, 'title', title,
    'description', description, 'amount', amount, 'product_id', product_id,
    'transaction_date', transaction_date, 'created_by', created_by,
    'created_at', created_at, 'updated_at', updated_at
);

-- name: CashflowDelete
-- params: id bigint
UPDATE cashflows SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;

-- name: CashflowSummary
-- params: start_date text, end_date text
SELECT jsonb_build_object(
    'total_profit', COALESCE(SUM(amount) FILTER (WHERE type = 'PROFIT'), 0),
    'total_loss', COALESCE(SUM(amount) FILTER (WHERE type = 'LOSS'), 0),
    'total_saving', COALESCE(SUM(amount) FILTER (WHERE type = 'SAVING'), 0),
    'net_profit', COALESCE(SUM(amount) FILTER (WHERE type = 'PROFIT'), 0) - COALESCE(SUM(amount) FILTER (WHERE type = 'LOSS'), 0)
) FROM cashflows
WHERE deleted_at IS NULL
  AND ($1 = '' OR transaction_date >= $1::date)
  AND ($2 = '' OR transaction_date <= $2::date);
