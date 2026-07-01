package userusecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Mpayy/e-commerce/dependency"
	configMock "github.com/Mpayy/e-commerce/internal/mocks"
	"github.com/Mpayy/e-commerce/internal/user/dto"
	"github.com/Mpayy/e-commerce/internal/user/entity"
	repoMock "github.com/Mpayy/e-commerce/internal/user/mocks"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func hashPassword(t *testing.T, plain string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	assert.NoError(t, err)
	return string(hashed)
}

func setupUserUsecase(t *testing.T) (UserUsecase, *repoMock.MockUserRepository, *configMock.MockRedis, *repoMock.MockJwtToken, *configMock.MockTransaction) {
	userRepository := repoMock.NewMockUserRepository(t)
	jwtTokenMock := repoMock.NewMockJwtToken(t)
	redisClientMock := configMock.NewMockRedis(t)
	transactionMock := configMock.NewMockTransaction(t)
	log := newTestLogger()

	usecase := NewUserUsecase(userRepository, redisClientMock, transactionMock, log, jwtTokenMock)
	return usecase, userRepository, redisClientMock, jwtTokenMock, transactionMock
}

// go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Login"
func TestUserUsecaseImpl_Login(t *testing.T) {
	ctx := context.Background()
	plainPassword := "rahasia123"
	hashedPassword := hashPassword(t, plainPassword)

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Login/successful_login"
	t.Run("successful_login", func(t *testing.T) {
		// ARRANGE
		usecase, userRepo, redisClient, jwtToken, _ := setupUserUsecase(t)

		request := &dto.UserLoginRequest{
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		existingUser := &entity.User{
			ID:       1,
			Email:    "test@mail.com",
			Password: hashedPassword,
			Role:     "customer",
		}

		userRepo.On("FindByEmail", ctx, request.Email).
			Return(existingUser, nil)

		jwtToken.On("CreateToken", &jwt.Auth{UserID: 1, Role: "customer"}).
			Return("dummy.jwt.token", nil)

		redisClient.On("SetToRedis", ctx, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, dependency.AuthPrefix)
		}), mock.Anything, jwt.TokenDuration).
			Return(nil)

		// ACT
		result, err := usecase.Login(ctx, request)

		// ASSERT
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "dummy.jwt.token", result.Token)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Login/failed_email_not_found"
	t.Run("failed_email_not_found", func(t *testing.T) {
		usecase, userRepo, _, _, _ := setupUserUsecase(t)

		request := &dto.UserLoginRequest{Email: "notfound@mail.com", Password: plainPassword}

		userRepo.On("FindByEmail", ctx, request.Email).
			Return(nil, apperror.ErrNotFound)

		result, err := usecase.Login(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrWrongEmailOrPassword)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Login/failed_wrong_password"
	t.Run("failed_wrong_password", func(t *testing.T) {
		usecase, userRepo, _, _, _ := setupUserUsecase(t)

		request := &dto.UserLoginRequest{Email: "test@mail.com", Password: "passwordSalah"}

		existingUser := &entity.User{
			ID:       1,
			Email:    "test@mail.com",
			Password: hashedPassword, // hash dari "rahasia123", bukan "passwordSalah"
			Role:     "customer",
		}

		userRepo.On("FindByEmail", ctx, request.Email).
			Return(existingUser, nil)

		result, err := usecase.Login(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrWrongEmailOrPassword)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Login/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, userRepo, _, _, _ := setupUserUsecase(t)

		request := &dto.UserLoginRequest{Email: "test@mail.com", Password: plainPassword}

		userRepo.On("FindByEmail", ctx, request.Email).
			Return(nil, errors.New("connection refused"))

		result, err := usecase.Login(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Login/failed_unexpected_error_from_redis"
	t.Run("failed_unexpected_error_from_redis", func(t *testing.T) {
		usecase, userRepo, redisClient, jwtToken, _ := setupUserUsecase(t)

		request := &dto.UserLoginRequest{Email: "test@mail.com", Password: plainPassword}

		existingUser := &entity.User{
			ID:       1,
			Email:    "test@mail.com",
			Password: hashedPassword,
			Role:     "customer",
		}

		userRepo.On("FindByEmail", ctx, request.Email).
			Return(existingUser, nil)

		jwtToken.On("CreateToken", &jwt.Auth{UserID: 1, Role: "customer"}).
			Return("dummy.jwt.token", nil)

		redisClient.On("SetToRedis", ctx, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, dependency.AuthPrefix)
		}), mock.Anything, jwt.TokenDuration).
			Return(errors.New("redis down"))

		result, err := usecase.Login(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Register"
func TestUserUsecaseImpl_Register(t *testing.T) {
	ctx := context.Background()
	plainPassword := "rahasia123"

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Register/successful_register"
	t.Run("successful_register", func(t *testing.T) {
		usecase, userRepo, _, _, transactionMock := setupUserUsecase(t)

		request := &dto.UserRegisterRequest{
			Name:     "test",
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		var errFromRepo error
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				errFromRepo = fn(ctx)
			}).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return errFromRepo
			})

		userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			if u.Name != request.Name || u.Email != request.Email || u.Role != entity.RoleCustomer {
				return false
			}
			return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(request.Password)) == nil
		})).Return(nil)

		result, err := usecase.Register(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.Name, result.Name)
		assert.Equal(t, request.Email, result.Email)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Register/failed_email_already_exists"
	t.Run("failed_email_already_exists", func(t *testing.T) {
		usecase, userRepo, _, _, transactionMock := setupUserUsecase(t)

		request := &dto.UserRegisterRequest{
			Name:     "test",
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		var errFromRepo error
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				errFromRepo = fn(ctx)
			}).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return errFromRepo
			})

		userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			if u.Name != request.Name || u.Email != request.Email || u.Role != entity.RoleCustomer {
				return false
			}
			return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(request.Password)) == nil
		})).
			Return(apperror.ErrDuplicatedKey)

		result, err := usecase.Register(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedEmail)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Register/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, userRepo, _, _, transactionMock := setupUserUsecase(t)

		request := &dto.UserRegisterRequest{
			Name:     "test",
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		var errFromRepo error
		transactionMock.On("WithTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				errFromRepo = fn(ctx)
			}).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return errFromRepo
			})

		userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			if u.Name != request.Name || u.Email != request.Email || u.Role != entity.RoleCustomer {
				return false
			}
			return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(request.Password)) == nil
		})).
			Return(errors.New("connection refused"))

		result, err := usecase.Register(ctx, request)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_GetProfile"
func TestUserUsecaseImpl_GetProfile(t *testing.T) {
	ctx := context.Background()

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_GetProfile/successful_get_profile"
	t.Run("successful_get_profile", func(t *testing.T) {
		usecase, userRepo, _, _, _ := setupUserUsecase(t)

		userId := uint(1)

		existingUser := &entity.User{
			ID:       1,
			Name:     "test",
			Email:    "test@mail.com",
			Password: "rahasia123",
			Role:     "customer",
		}

		userRepo.On("FindByID", mock.Anything, uint(1)).Return(existingUser, nil)

		result, err := usecase.GetProfile(ctx, userId)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, existingUser.ID, result.ID)
		assert.Equal(t, existingUser.Name, result.Name)
		assert.Equal(t, existingUser.Email, result.Email)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_GetProfile/failed_user_not_found"
	t.Run("failed_user_not_found", func(t *testing.T) {
		usecase, userRepo, _, _, _ := setupUserUsecase(t)

		userId := uint(1)

		userRepo.On("FindByID", mock.Anything, uint(1)).Return(nil, apperror.ErrNotFound)

		result, err := usecase.GetProfile(ctx, userId)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrNotFound)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_GetProfile/failed_unexpected_error_from_repository"
	t.Run("failed_unexpected_error_from_repository", func(t *testing.T) {
		usecase, userRepo, _, _, _ := setupUserUsecase(t)

		userId := uint(1)

		userRepo.On("FindByID", mock.Anything, uint(1)).Return(nil, errors.New("connection refused"))

		result, err := usecase.GetProfile(ctx, userId)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}

// go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Logout"
func TestUserUsecaseImpl_Logout(t *testing.T) {
	ctx := context.Background()

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Logout/successful_logout"
	t.Run("successful_logout", func(t *testing.T) {
		usecase, _, redisClient, _, _ := setupUserUsecase(t)

		token := "dummy.jwt.token"

		redisClient.On("DeleteFromRedis", ctx, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, dependency.AuthPrefix+token)
		})).
			Return(nil)

		err := usecase.Logout(ctx, token)

		assert.NoError(t, err)
	})

	//go test -v ./internal/user/usecase -run "TestUserUsecaseImpl_Logout/failed_unexpected_error_from_redis"
	t.Run("failed_unexpected_error_from_redis", func(t *testing.T) {
		usecase, _, redisClient, _, _ := setupUserUsecase(t)

		token := "dummy.jwt.token"

		redisClient.On("DeleteFromRedis", ctx, mock.MatchedBy(func(key string) bool {
			return strings.HasPrefix(key, dependency.AuthPrefix+token)
		})).
			Return(errors.New("connection refused"))

		err := usecase.Logout(ctx, token)

		assert.ErrorIs(t, err, apperror.ErrInternalServer)
	})
}
