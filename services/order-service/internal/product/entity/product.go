package entity

import "time"

type Product struct {
	ID          uint
	CategoryID  uint
	Name        string
	Slug        string
	Description string
	Price       float64
	Stock       int
	SKU         string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BulkDecreaseStock struct {
	ProductID uint
	Quantity  int
}
