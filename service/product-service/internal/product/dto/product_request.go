package dto

type ProductCreateRequest struct {
	CategoryID  uint    `json:"category_id" validate:"required"`
	Name        string  `json:"name" validate:"required,max=150"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"min=0"`
	SKU         string  `json:"sku" validate:"omitempty,max=50"`
}

type ProductUpdateRequest struct {
	CategoryID  uint    `json:"category_id" validate:"required"`
	Name        string  `json:"name"        validate:"required,max=150"`
	Description string  `json:"description"`
	Price       float64 `json:"price"       validate:"required,gt=0"`
	Stock       int     `json:"stock"       validate:"min=0"`
	SKU         string  `json:"sku"         validate:"omitempty,max=50"`
	IsActive    *bool   `json:"is_active"   validate:"omitempty"`
}

type ProductSearchRequest struct {
	Search     string `form:"search" validate:"omitempty,max=150"`
	CategoryID uint   `form:"category_id" validate:"omitempty"`
	Page       int    `form:"page" validate:"omitempty,min=1"`
	Limit      int    `form:"limit" validate:"omitempty,min=1,max=100"`
}

type ProductStockAdjustmentRequest struct {
	Stock *int `json:"stock" validate:"required,gte=0"`
}
