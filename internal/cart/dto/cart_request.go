package dto

type CartItemCreateRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gte=1"`
}

type CartItemUpdateRequest struct {
	Quantity *int `json:"quantity" validate:"required,gte=0"`
}
