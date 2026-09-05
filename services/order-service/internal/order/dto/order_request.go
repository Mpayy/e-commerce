package dto

type OrderFilter struct {
	Page  int `form:"page" validate:"omitempty,gte=1"`
	Limit int `form:"limit" validate:"omitempty,gte=1,lte=100"`
}

type SalesAnalyticsRequest struct {
	From  string `form:"from" validate:"omitempty,datetime=2006-01-02"`
	To    string `form:"to" validate:"omitempty,datetime=2006-01-02,gtefield=From"`
	Limit int    `form:"limit" validate:"omitempty,min=1,max=50"`
}

type AdminOrderListRequest struct {
	Status    string  `form:"status" validate:"omitempty"`
	UserID    uint    `form:"user_id" validate:"omitempty"`
	MinAmount float64 `form:"min_amount" validate:"omitempty,gte=0"`
	MaxAmount float64 `form:"max_amount" validate:"omitempty,gte=0,gtefield=MinAmount"`
	From      string  `form:"from" validate:"omitempty,datetime=2006-01-02"`
	To        string  `form:"to" validate:"omitempty,datetime=2006-01-02,gtefield=From"`
	Page      int     `form:"page" validate:"omitempty,min=1"`
	Limit     int     `form:"limit" validate:"omitempty,min=1,max=100"`
}

type AdminCancelOrderRequest struct {
	Status string `json:"status" validate:"required,oneof=CANCELLED"`
}
