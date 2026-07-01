CREATE TABLE IF NOT EXISTS orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    total_amount DECIMAL(12, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PAID',
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_user_id (user_id)
);
    