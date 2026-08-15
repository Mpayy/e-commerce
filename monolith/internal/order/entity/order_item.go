package entity

type OrderItem struct {
	ID          uint    `gorm:"column:id;primaryKey"`
	OrderID     uint    `gorm:"column:order_id;index;not null"`
	ProductID   uint    `gorm:"column:product_id;index;not null"`
	ProductName string  `gorm:"column:product_name;type:varchar(150)"`
	Price       float64 `gorm:"column:price;type:decimal(12,2)"`
	Quantity    int     `gorm:"column:quantity;not null"`
	Subtotal    float64 `gorm:"column:subtotal;type:decimal(12,2)"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
