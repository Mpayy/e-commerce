package entity

import "time"

type Order struct {
	ID            uint
	UserID        uint
	InvoiceNumber string
	TotalAmount   float64
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Items         []OrderItem
}

type DailyRevenueRow struct {
	Date         time.Time
	OrderCount   int64
	DailyRevenue float64
	RunningTotal float64
}
