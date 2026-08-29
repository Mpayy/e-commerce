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

type TopProductRow struct {
	ProductID uint
	ProductName string
	TotalQuantitySold int64
	TotalRevenue float64
	Rank int64
}