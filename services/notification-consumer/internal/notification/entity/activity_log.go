package entity

import "time"

type ActivityLog struct {
	ID        int64
	OrderID   uint
	UserID    uint
	Message   string
	CreatedAt time.Time
}
