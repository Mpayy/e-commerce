package entity

type OrderItem struct {
	ID          uint    `gorm:"primaryKey"`
	OrderID     uint    `gorm:"index;not null"`
	ProductID   uint    `gorm:"index;not null"`
	ProductName string  `gorm:"type:varchar(150)"`
	Price       float64 `gorm:"type:decimal(12,2)"`
	Quantity    int     `gorm:"not null"`
	Subtotal    float64 `gorm:"type:decimal(12,2)"`
}

func (OrderItem) TableName() string {
	return "order_items"
}