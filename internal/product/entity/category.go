package entity

import "time"

type Category struct {
	ID        uint      `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name;type:varchar(100);not null"`
	Slug      string    `gorm:"column:slug;type:varchar(120);unique"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (c Category) TableName() string {
	return "categories"
}
