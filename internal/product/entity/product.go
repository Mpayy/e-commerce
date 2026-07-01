package entity

import "time"

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CategoryID  uint      `gorm:"index;not null" json:"category_id"`
	Name        string    `gorm:"type:varchar(150);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(180);uniqueIndex" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(12,2);not null" json:"price"`
	Stock       int       `gorm:"not null;default:0" json:"stock"`
	SKU         string    `gorm:"type:varchar(50);uniqueIndex" json:"sku"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Product) TableName() string { return "products" }