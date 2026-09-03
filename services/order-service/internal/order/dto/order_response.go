package dto

type OrderResponse struct {
	OrderID       uint                `json:"order_id"`
	InvoiceNumber string              `json:"invoice_number"`
	TotalAmount   float64             `json:"total_amount"`
	Status        string              `json:"status"`
	Items         []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type OrderHistoryResponse struct {
	Orders []OrderResponse `json:"orders"`
	Meta   MetaPagination  `json:"meta"`
}

type MetaPagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type SalesAnalyticsResponse struct {
	Period       PeriodResponse          `json:"period"`
	Summary      SummaryResponse         `json:"summary"`
	DailyRevenue []DailyRevenueResponse  `json:"daily_revenue"`
	TopProducts  []TopProductResponse    `json:"top_products"`
}

type PeriodResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type SummaryResponse struct {
	TotalRevenue float64 `json:"total_revenue"`
	TotalOrders  int64   `json:"total_orders"`
}

type DailyRevenueResponse struct {
	Date         string  `json:"date"`
	OrderCount   int64   `json:"order_count"`
	DailyRevenue float64 `json:"daily_revenue"`
	RunningTotal float64 `json:"running_total"`
}

type TopProductResponse struct {
	Rank              int64   `json:"rank"`
	ProductID         uint    `json:"product_id"`
	ProductName       string  `json:"product_name"`
	TotalQuantitySold int64   `json:"total_quantity_sold"`
	TotalRevenue      float64 `json:"total_revenue"`
}

type AdminOrderListResponse struct {
	Orders []AdminOrderSummaryResponse `json:"orders"`
	Meta   MetaPagination               `json:"meta"`
}

type AdminOrderSummaryResponse struct {
	OrderID       uint    `json:"order_id"`
	UserID        uint    `json:"user_id"`
	InvoiceNumber string  `json:"invoice_number"`
	TotalAmount   float64 `json:"total_amount"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}
