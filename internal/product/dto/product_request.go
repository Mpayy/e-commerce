package dto

type ProductRequest struct {
	CategoryID  uint    `json:"category_id" validate:"required"`
	Name        string  `json:"name" validate:"required,max=150"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"min=0"`
	SKU         string  `json:"sku" validate:"omitempty,max=50"`
}
