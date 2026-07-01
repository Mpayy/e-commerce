package dto

type ProductResponse struct {
	ID          uint    `json:"id"`
	CategoryID  uint    `json:"category_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	SKU         string  `json:"sku"`
	IsActive    bool    `json:"is_active"`
}
