-- name: CreateOrder :one
INSERT INTO orders (user_id, total_amount, status, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: UpdateInvoiceNumber :exec
UPDATE orders
SET invoice_number = $2
WHERE id = $1;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: CountOrdersByUser :one
SELECT COUNT(*)
FROM orders
WHERE user_id = $1;

-- name: GetOrderByID :one
SELECT *
FROM orders
WHERE id = $1;