package model

import "time"

type CategoryModel struct {
	ID        int64     `bson:"_id"`
	Name      string    `bson:"name"`
	Slug      string    `bson:"slug"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func (c *CategoryModel) CollectionName() string {
	return "categories"
}
