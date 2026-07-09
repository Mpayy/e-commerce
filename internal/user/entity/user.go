package entity

import "time"

const (
	RoleCustomer string = "customer"
	RoleAdmin    string = "admin"
)

type User struct {
	ID        uint   `gorm:"column:id;primaryKey"`
	Name      string `gorm:"column:name;type:varchar(100);not null"`
	Email     string `gorm:"column:email;type:varchar(150);not null;uniqueIndex"`
	Password  string `gorm:"column:password;type:varchar(255);not null"`
	Role      string `gorm:"column:role;type:varchar(20);not null;default:'customer'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}
