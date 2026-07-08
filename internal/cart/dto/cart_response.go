package dto

type CartDetailResponse struct {
	Items      []CartItemResponse `json:"items"`
	GrandTotal float64            `json:"grand_total"`
}

type CartItemResponse struct {
	ProductID      uint    `json:"product_id"`
	Name           string  `json:"name"`
	Price          float64 `json:"price"`
	Quantity       int     `json:"quantity"`
	Subtotal       float64 `json:"subtotal"`
	StockAvailable int     `json:"stock_available"`
}
