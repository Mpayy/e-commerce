package entity

import "time"

type Product struct {
	ID          uint      `gorm:"column:id;primaryKey"`
	CategoryID  uint      `gorm:"column:category_id;not null;index"`
	Name        string    `gorm:"column:name;type:varchar(150);not null"`
	Slug        string    `gorm:"column:slug;type:varchar(180);unique"`
	Description string    `gorm:"column:description;type:text"`
	Price       float64   `gorm:"column:price;type:decimal(12,2);not null"`
	Stock       int       `gorm:"column:stock;not null;default:0"`
	SKU         string    `gorm:"column:sku;type:varchar(50);unique"`
	IsActive    bool      `gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Product) TableName() string { return "products" }

type ProductFilter struct {
	Search     string
	CategoryID uint
	Page       int
	Limit      int
}
