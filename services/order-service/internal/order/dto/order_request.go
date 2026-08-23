package dto

type OrderFilter struct {
	Page  int `form:"page" validate:"omitempty,gte=1"`
	Limit int `form:"limit" validate:"omitempty,gte=1,lte=100"`
}
