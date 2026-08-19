package model

import "time"

type StockLedgerItem struct {
	ProductID int64 `bson:"product_id"`
	Quantity  int   `bson:"quantity"`
}

type StockLedgerModel struct {
	ID         string            `bson:"_id"`
	CheckoutID string            `bson:"checkout_id"`
	Operation  string            `bson:"operation"`
	Items      []StockLedgerItem `bson:"items"`
	CreatedAt  time.Time         `bson:"created_at"`
}

func (StockLedgerModel) CollectionName() string {
	return "stock_ledgers"
}