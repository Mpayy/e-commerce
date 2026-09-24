package entity

import "time"

const MaxImageSize = 2 * 1024 * 1024

type Product struct {
	ID          uint
	CategoryID  uint
	Name        string
	Slug        string
	ImagePath   string
	Description string
	Price       float64
	Stock       int
	SKU         string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductFilter struct {
	Search     string
	CategoryID uint
	Page       int
	Limit      int
}

type StockItem struct {
	ProductID uint
	Quantity  int
}
