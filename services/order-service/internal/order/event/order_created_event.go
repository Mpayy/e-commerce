package event

type OrderCreatedEvent struct {
	OrderID       uint    `json:"order_id"`
	UserID        uint    `json:"user_id"`
	InvoiceNumber string  `json:"invoice_number"`
	TotalAmount   float64 `json:"total_amount"`
}