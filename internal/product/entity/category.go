package entity

import "time"

type Category struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(100);not null"`
	Slug      string `gorm:"type:varchar(120);uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c Category) TableName() string {
	return "categories"
}
