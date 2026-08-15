-- CREATE TABLE IF NOT EXISTS orders (
--     id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
--     user_id BIGINT UNSIGNED NOT NULL,
--     invoice_number VARCHAR(50) UNIQUE NULL,
--     total_amount DECIMAL(12, 2) NOT NULL,
--     status VARCHAR(20) NOT NULL DEFAULT 'PAID',
--     created_at DATETIME,
--     updated_at DATETIME,
--     INDEX idx_user_id (user_id)
-- );
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    invoice_number VARCHAR(50) UNIQUE NULL,
    total_amount DECIMAL(12, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PAID',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
    