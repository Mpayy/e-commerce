package dto

type OrderFilter struct {
	Page  int `form:"page" validate:"omitempty,gte=1"`
	Limit int `form:"limit" validate:"omitempty,gte=1,lte=100"`
}

type SalesAnalyticsRequest struct {
	From  string `form:"from" validate:"omitempty,datetime=2006-01-02"`
	To    string `form:"to" validate:"omitempty,datetime=2006-01-02"`
	Limit int    `form:"limit" validate:"omitempty,min=1,max=50"`
}