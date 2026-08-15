package model

import "time"

type ProductModel struct {
	ID          int64     `bson:"_id"`
	CategoryID  int64     `bson:"category_id"`
	Name        string    `bson:"name"`
	Slug        string    `bson:"slug"`
	Description string    `bson:"description"`
	Price       float64   `bson:"price"`
	Stock       int       `bson:"stock"`
	SKU         string    `bson:"sku"`
	IsActive    bool      `bson:"is_active"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func (p *ProductModel) CollectionName() string {
	return "products"
}
