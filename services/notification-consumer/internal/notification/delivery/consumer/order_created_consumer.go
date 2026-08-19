package consumer

import "context"

type OrderCreatedConsumer interface {
	Start(ctx context.Context) error
}
	