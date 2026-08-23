package entity

type OrderItem struct {
	ID          uint
	OrderID     uint
	ProductID   uint
	ProductName string
	Price       float64
	Quantity    int
	Subtotal    float64
}
