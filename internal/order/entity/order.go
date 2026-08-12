package entity

import "time"

type Order struct {
	ID            uint      `gorm:"column:id;primaryKey"`
	UserID        uint      `gorm:"column:user_id;not null;index"`
	InvoiceNumber string    `gorm:"column:invoice_number;type:varchar(50);unique"`
	TotalAmount   float64   `gorm:"column:total_amount;type:decimal(12,2);not null"`
	Status        string    `gorm:"column:status;type:varchar(20);default:'PAID'"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Order) TableName() string {
	return "orders"
}
