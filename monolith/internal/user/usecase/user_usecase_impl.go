package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Mpayy/e-commerce/monolith/internal/user/dto"
	"github.com/Mpayy/e-commerce/monolith/internal/user/entity"
	"github.com/Mpayy/e-commerce/monolith/internal/user/repository"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecaseImpl struct {
	userRepository      repository.UserRepository
	userRedisRepository repository.UserRedisRepository
	transaction         transaction.Transaction
	log                 *logger.Logger
	jwtToken            jwt.JwtToken
}

func NewUserUsecase(userRepo repository.UserRepository, userRedisRepo repository.UserRedisRepository, tx transaction.Transaction, log *logger.Logger, jwt jwt.JwtToken) UserUsecase {
	return &UserUsecaseImpl{
		userRepository:      userRepo,
		userRedisRepository: userRedisRepo,
		transaction:         tx,
		log:                 log,
		jwtToken:            jwt,
	}
}

func (u *UserUsecaseImpl) Register(ctx context.Context, request *dto.UserRegisterRequest) (*dto.UserResponse, error) {
	log := u.log.WithField("email", request.Email)
	log.Debug("Attempting to register user")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	user := &entity.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: string(hashedPassword),
		Role:     entity.RoleCustomer,
	}

	if err := u.userRepository.Create(ctx, user); err != nil {
		if errors.Is(err, apperror.ErrDuplicatedKey) {
			return nil, apperror.ErrDuplicatedEmail
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	response := &dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	log.WithField("user_id", user.ID).Info("User registered successfully")
	return response, nil
}

func (u *UserUsecaseImpl) Login(ctx context.Context, request *dto.UserLoginRequest) (*dto.TokenResponse, error) {
	log := u.log.WithField("email", request.Email)
	log.Debug("Attempting to login user")

	user, err := u.userRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.ErrWrongEmailOrPassword
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return nil, apperror.ErrWrongEmailOrPassword
	}

	auth := &jwt.Auth{
		ID:   user.ID,
		Role: user.Role,
	}

	token, err := u.jwtToken.Create(auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	authData, err := json.Marshal(auth)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth data: %w", err)
	}

	err = u.userRedisRepository.SaveSession(ctx, token, authData, jwt.TokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to save token to redis: %w", err)
	}

	tokenResponse := &dto.TokenResponse{
		Token: token,
	}

	log.WithField("user_id", user.ID).Info("User logged in successfully")
	return tokenResponse, nil
}

func (u *UserUsecaseImpl) GetProfile(ctx context.Context, userId uint) (*dto.UserResponse, error) {
	log := u.log.WithField("user_id", userId)
	log.Debug("Attempting to get user profile")

	user, err := u.userRepository.FindByID(ctx, userId)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	response := &dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	log.Info("User profile retrieved successfully")
	return response, nil
}

func (u *UserUsecaseImpl) Logout(ctx context.Context, token string) error {
	log := u.log.WithField("token", token)
	log.Debug("Logout attempt")

	err := u.userRedisRepository.DeleteSession(ctx, token)
	if err != nil {
		log.WithError(err).Error("Failed to delete token from redis")
		return apperror.ErrInternalServer
	}

	log.Info("User logged out successfully")
	return nil
}
