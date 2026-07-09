package entity

import "time"

type Product struct {
	ID          uint    `gorm:"column:id;primaryKey"`
	CategoryID  uint    `gorm:"column:category_id;index;not null"`
	Name        string  `gorm:"column:name;type:varchar(150);not null"`
	Slug        string  `gorm:"column:slug;type:varchar(180);uniqueIndex"`
	Description string  `gorm:"column:description;type:text"`
	Price       float64 `gorm:"column:price;type:decimal(12,2);not null"`
	Stock       int     `gorm:"column:stock;not null;default:0"`
	SKU         string  `gorm:"column:sku;type:varchar(50);uniqueIndex"`
	IsActive    bool    `gorm:"column:is_active;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Product) TableName() string { return "products" }

type ProductFilter struct {
	Search     string
	CategoryID uint
	Page       int
	Limit      int
}
