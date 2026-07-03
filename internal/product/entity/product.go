package entity

import "time"

type Product struct {
	ID          uint    `gorm:"primaryKey"`
	CategoryID  uint    `gorm:"index;not null"`
	Name        string  `gorm:"type:varchar(150);not null"`
	Slug        string  `gorm:"type:varchar(180);uniqueIndex"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"type:decimal(12,2);not null"`
	Stock       int     `gorm:"not null;default:0"`
	SKU         string  `gorm:"type:varchar(50);uniqueIndex"`
	IsActive    bool    `gorm:"default:true"`
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
