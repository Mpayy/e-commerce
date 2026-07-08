package cartrepository

import (
	"context"
)

type CartRedisRepository interface {
	AddItem(ctx context.Context, userID uint, productID uint, quantity int) error
	UpdateItem(ctx context.Context, userID uint, productID uint, quantity int) error
	RemoveItem(ctx context.Context, userID uint, productID uint) error
}
