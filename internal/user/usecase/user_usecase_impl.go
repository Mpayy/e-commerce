package userusecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Mpayy/e-commerce/dependency"
	"github.com/Mpayy/e-commerce/internal/user/dto"
	"github.com/Mpayy/e-commerce/internal/user/entity"
	userrepository "github.com/Mpayy/e-commerce/internal/user/repository"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecaseImpl struct {
	UserRepository userrepository.UserRepository
	Redis          dependency.Redis
	Transaction    transaction.Transaction
	Log            *logrus.Logger
	JwtToken       jwt.JwtToken
}

func NewUserUsecase(userRepo userrepository.UserRepository, redis dependency.Redis, tx transaction.Transaction, log *logrus.Logger, jwt jwt.JwtToken) UserUsecase {
	return &UserUsecaseImpl{
		UserRepository: userRepo,
		Redis:          redis,
		Transaction:    tx,
		Log:            log,
		JwtToken:       jwt,
	}
}

func (u *UserUsecaseImpl) Register(ctx context.Context, request *dto.UserRegisterRequest) (*dto.UserResponse, error) {
	u.Log.WithField("email", request.Email).Debug("Attempting to register user")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"email": request.Email,
			"error": err,
		}).Error("Failed to generate password")
		return nil, apperror.ErrInternalServer
	}

	user := &entity.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: string(hashedPassword),
		Role:     entity.RoleCustomer,
	}

	err = u.Transaction.WithTransaction(ctx, func(ctx context.Context) error {
		errCreate := u.UserRepository.Create(ctx, user)
		if errCreate != nil {
			u.Log.WithField("email", request.Email).
				Warn("Create user failed: duplicate email")
			if errors.Is(errCreate, apperror.ErrDuplicatedKey) {
				return apperror.ErrDuplicatedEmail
			}
			u.Log.WithFields(logrus.Fields{
				"email": request.Email,
				"error": errCreate,
			}).Error("Create user failed: unexpected DB error")
			return apperror.ErrInternalServer
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	u.Log.WithField("user_id", user.ID).Info("User registered successfully")
	return response, nil
}

func (u *UserUsecaseImpl) Login(ctx context.Context, request *dto.UserLoginRequest) (*dto.TokenResponse, error) {
	u.Log.WithField("email", request.Email).Debug("Attempting to login user")

	user, err := u.UserRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			u.Log.WithField("email", request.Email).Warn("Login failed: user not found")
			return nil, apperror.ErrWrongEmailOrPassword
		}
		u.Log.WithFields(logrus.Fields{
			"email": request.Email,
			"error": err,
		}).Error("Login failed: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"email": request.Email,
			"error": err,
		}).Error("Failed to compare password")
		return nil, apperror.ErrWrongEmailOrPassword
	}

	auth := &jwt.Auth{
		UserID: user.ID,
		Role:   user.Role,
	}

	token, err := u.JwtToken.CreateToken(auth)
	if err != nil {
		u.Log.WithError(err).Error("Failed to create token")
		return nil, apperror.ErrInternalServer
	}

	authData, err := json.Marshal(auth)
	if err != nil {
		u.Log.WithError(err).Error("Failed to marshal auth data")
		return nil, apperror.ErrInternalServer
	}

	err = u.Redis.SetToRedis(ctx, dependency.AuthPrefix+token, authData, jwt.TokenDuration)
	if err != nil {
		u.Log.WithError(err).Error("Failed to save token to redis")
		return nil, apperror.ErrInternalServer
	}

	tokenResponse := &dto.TokenResponse{
		Token: token,
	}

	u.Log.WithField("user_id", user.ID).Info("User logged in successfully")
	return tokenResponse, nil
}

func (u *UserUsecaseImpl) GetProfile(ctx context.Context, userId uint) (*dto.UserResponse, error) {
	u.Log.WithField("user_id", userId).Debug("Attempting to get user profile")

	user, err := u.UserRepository.FindByID(ctx, userId)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			u.Log.WithField("user_id", userId).Warn("Get profile failed: user not found")
			return nil, apperror.ErrNotFound
		}
		u.Log.WithFields(logrus.Fields{
			"user_id": userId,
			"error":   err,
		}).Error("Get profile failed: unexpected DB error")
		return nil, apperror.ErrInternalServer
	}

	response := &dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	u.Log.WithField("user_id", userId).Info("User profile retrieved successfully")
	return response, nil
}

func (u *UserUsecaseImpl) Logout(ctx context.Context, token string) error {
	u.Log.WithField("token", token).Debug("Logout attempt")

	err := u.Redis.DeleteFromRedis(ctx, dependency.AuthPrefix+token)
	if err != nil {
		u.Log.WithFields(logrus.Fields{
			"token": token,
			"error": err,
		}).Error("Failed to delete token from redis")
		return apperror.ErrInternalServer
	}

	u.Log.WithField("token", token).Info("User logged out successfully")
	return nil

}
