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

type OrderFilter struct {
	UserID uint
	Status string
	MinAmount float64
	MaxAmount float64
	From *time.Time
	To *time.Time
	Page int
	Limit int
}