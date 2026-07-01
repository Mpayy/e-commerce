package userusecase

import (
	"context"

	"github.com/Mpayy/e-commerce/internal/user/dto"
)

type UserUsecase interface {
	Register(ctx context.Context, request *dto.UserRegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, request *dto.UserLoginRequest) (*dto.TokenResponse, error)
	GetProfile(ctx context.Context, userId uint) (*dto.UserResponse, error)
	Logout(ctx context.Context, token string) error
}
