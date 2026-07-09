package entity

import "time"

type Category struct {
	ID        uint   `gorm:"column:id;primaryKey"`
	Name      string `gorm:"column:name;type:varchar(100);not null"`
	Slug      string `gorm:"column:slug;type:varchar(120);uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c Category) TableName() string {
	return "categories"
}
