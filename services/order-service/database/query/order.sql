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
SELECT COUNT(*)::BIGINT
FROM orders
WHERE user_id = $1;

-- name: GetOrderByID :one
SELECT *
FROM orders
WHERE id = $1;

-- name: GetDailyRevenueReport :many
WITH daily_summary AS (
    SELECT
    DATE(created_at) AS date,
    COUNT(*)::BIGINT AS order_count,
    SUM(total_amount)::NUMERIC AS daily_revenue
    FROM orders
    WHERE status = 'PAID'
      AND created_at >= sqlc.arg(from_date)::TIMESTAMP 
      AND created_at <= sqlc.arg(to_date)::TIMESTAMP
    GROUP BY DATE(created_at)
)
SELECT
date, 
order_count,
daily_revenue,
SUM(daily_revenue) OVER (ORDER BY date ASC)::NUMERIC AS running_total
FROM daily_summary
ORDER BY date ASC;

-- name: CancelOrder :one
UPDATE orders
SET status = 'CANCELLED', updated_at = NOW()
WHERE id = sqlc.arg(order_id) AND status = 'PAID'
RETURNING id, invoice_number, status;