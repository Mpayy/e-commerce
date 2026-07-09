package dto

type CheckoutResponse struct {
	OrderID       uint                `json:"order_id"`
	InvoiceNumber string              `json:"invoice_number"`
	TotalAmount   float64             `json:"total_amount"`
	Status        string              `json:"status"`
	Items         []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}
