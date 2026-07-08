package entity

import "time"

type Order struct {
	ID            uint    `gorm:"primaryKey"`
	UserID        uint    `gorm:"index;not null"`
	InvoiceNumber string  `gorm:"type:varchar(50);uniqueIndex;not null"`
	TotalAmount   float64 `gorm:"type:decimal(12,2);not null"`
	Status        string  `gorm:"type:varchar(20);default:'PAID'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Order) TableName() string {
	return "orders"
}
