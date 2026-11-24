-- name: ProductCreate
-- params: sku text, name text, description text, category text, cost_price numeric, selling_price numeric, stock int, unit text, image_url text, status text, created_by bigint
INSERT INTO products (
    sku, name, description, category, cost_price, selling_price, stock, unit, image_url, status, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING jsonb_build_object(
    'id', id, 'sku', sku, 'name', name, 'description', description,
    'category', category, 'cost_price', cost_price, 'selling_price', selling_price,
    'stock', stock, 'unit', unit, 'image_url', image_url, 'status', status,
    'created_by', created_by, 'created_at', created_at, 'updated_at', updated_at
);

-- name: ProductGetByID
-- params: id bigint
SELECT jsonb_build_object(
    'id', id, 'sku', sku, 'name', name, 'description', description,
    'category', category, 'cost_price', cost_price, 'selling_price', selling_price,
    'stock', stock, 'unit', unit, 'image_url', image_url, 'status', status,
    'created_by', created_by, 'created_at', created_at, 'updated_at', updated_at
) FROM products WHERE id = $1 AND deleted_at IS NULL;

-- name: ProductGetBySKU
-- params: sku text
SELECT jsonb_build_object(
    'id', id, 'sku', sku, 'name', name, 'status', status
) FROM products WHERE sku = $1 AND deleted_at IS NULL;

-- name: ProductList
-- params: cursor_id bigint, limit int, category text, status text, search text
SELECT jsonb_build_object(
    'id', id, 'sku', sku, 'name', name, 'description', description,
    'category', category, 'cost_price', cost_price, 'selling_price', selling_price,
    'stock', stock, 'unit', unit, 'image_url', image_url, 'status', status,
    'created_by', created_by, 'created_at', created_at, 'updated_at', updated_at
) FROM products
WHERE deleted_at IS NULL
  AND ($1 = 0 OR id < $1)
  AND ($3 = '' OR category = $3)
  AND ($4 = '' OR status = $4)
  AND ($5 = '' OR name ILIKE '%' || $5 || '%' OR sku ILIKE '%' || $5 || '%')
ORDER BY id DESC
LIMIT $2;

-- name: ProductUpdate
-- params: id bigint, name text, description text, category text, cost_price numeric, selling_price numeric, stock int, unit text, image_url text, status text
UPDATE products SET
    name = $2,
    description = $3,
    category = $4,
    cost_price = $5,
    selling_price = $6,
    stock = $7,
    unit = $8,
    image_url = $9,
    status = $10,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING jsonb_build_object(
    'id', id, 'sku', sku, 'name', name, 'description', description,
    'category', category, 'cost_price', cost_price, 'selling_price', selling_price,
    'stock', stock, 'unit', unit, 'image_url', image_url, 'status', status,
    'created_by', created_by, 'created_at', created_at, 'updated_at', updated_at
);

-- name: ProductUpdateStock
-- params: id bigint, stock int, status text
UPDATE products SET
    stock = $2,
    status = $3,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING jsonb_build_object('id', id, 'stock', stock, 'status', status);

-- name: ProductDelete
-- params: id bigint
UPDATE products SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;
