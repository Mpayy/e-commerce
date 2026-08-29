-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, product_name, price, quantity, subtotal)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetItemsByOrderID :many
SELECT *
FROM order_items
WHERE order_id = $1;

-- name: ListItemsByOrderIDs :many
SELECT *
FROM order_items
WHERE order_id = ANY(sqlc.arg(order_ids)::BIGINT[]);

-- name: GetTopProducts :many
SELECT 
    oi.product_id,
    oi.product_name,
    SUM(oi.quantity)::BIGINT AS total_quantity_sold,
    SUM(oi.subtotal)::NUMERIC AS total_revenue,
    DENSE_RANK() OVER (
        ORDER BY SUM(oi.quantity) DESC, SUM(oi.subtotal) DESC
        )::BIGINT AS rank
FROM orders AS o
JOIN order_items AS oi ON o.id = oi.order_id
WHERE o.status = 'PAID'
  AND o.created_at >= sqlc.arg(from_date)::TIMESTAMP
  AND o.created_at <= sqlc.arg(to_date)::TIMESTAMP
GROUP BY oi.product_id, oi.product_name
ORDER BY rank ASC
LIMIT sqlc.arg(limit_count);