package repository

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/user/entity"
)

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../mocks/mock_user_repository.go
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uint) (*entity.User, error)
}
