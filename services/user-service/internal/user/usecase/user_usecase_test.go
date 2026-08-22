package usecase

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/Mpayy/e-commerce/pkg/jwt"
	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/dto"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/entity"
	repoMock "github.com/Mpayy/e-commerce/services/user-service/internal/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func newTestLogger() *logger.Logger {
	cfg := config.Load()
	log := logger.NewLogger(cfg)
	log.SetOutput(io.Discard)
	return log
}

func hashPassword(t *testing.T, plain string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	assert.NoError(t, err)
	return string(hashed)
}

func setupUserUsecase(t *testing.T) (UserUsecase, *repoMock.MockUserRepository, *repoMock.MockUserRedisRepository, *repoMock.MockJwtToken) {
	userRepository := repoMock.NewMockUserRepository(t)
	UserRedisRepository := repoMock.NewMockUserRedisRepository(t)
	jwtTokenMock := repoMock.NewMockJwtToken(t)
	log := newTestLogger()

	usecase := NewUserUsecase(userRepository, UserRedisRepository, log, jwtTokenMock)
	return usecase, userRepository, UserRedisRepository, jwtTokenMock
}

func TestUserUsecaseImpl_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("success_register", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		req := &dto.UserRegisterRequest{
			Name:     "Test User",
			Email:    "test@mail.com",
			Password: "password123",
		}

		userRepo.EXPECT().
			Create(ctx, mock.MatchedBy(func(u *entity.User) bool {
				return u.Name == req.Name && u.Email == req.Email && u.Role == entity.RoleCustomer
			})).
			Run(func(ctx context.Context, u *entity.User) {
				u.ID = 1 // Simulasi RETURNING id dari DB
			}).
			Return(nil)

		res, err := uc.Register(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, uint(1), res.ID)
		assert.Equal(t, req.Name, res.Name)
		assert.Equal(t, req.Email, res.Email)
	})

	t.Run("err_hash_password_failed", func(t *testing.T) {
		uc, _, _, _ := setupUserUsecase(t)

		// Password lebih dari 72 byte akan memicu error dari bcrypt
		longPassword := make([]byte, 73)
		for i := range longPassword {
			longPassword[i] = 'a'
		}

		req := &dto.UserRegisterRequest{
			Name:     "Test User",
			Email:    "test@mail.com",
			Password: string(longPassword),
		}

		res, err := uc.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("err_duplicated_email", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		req := &dto.UserRegisterRequest{
			Name:     "Test User",
			Email:    "existing@mail.com",
			Password: "password123",
		}

		userRepo.EXPECT().
			Create(ctx, mock.Anything).
			Return(apperror.ErrDuplicatedKey)

		res, err := uc.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrDuplicatedEmail)
	})

	t.Run("err_failed_to_create_user", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		req := &dto.UserRegisterRequest{
			Name:     "Test User",
			Email:    "test@mail.com",
			Password: "password123",
		}

		dbErr := errors.New("database connection lost")
		userRepo.EXPECT().
			Create(ctx, mock.Anything).
			Return(dbErr)

		res, err := uc.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "failed to create user")
	})
}

func TestUserUsecaseImpl_Login(t *testing.T) {
	ctx := context.Background()
	plainPassword := "rahasia123"
	hashedPassword := hashPassword(t, plainPassword)

	t.Run("success_login", func(t *testing.T) {
		uc, userRepo, redisRepo, jwtToken := setupUserUsecase(t)

		req := &dto.UserLoginRequest{
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		existingUser := &entity.User{
			ID:       1,
			Name:     "Test User",
			Email:    "test@mail.com",
			Password: hashedPassword,
			Role:     entity.RoleCustomer,
		}

		expectedAuth := &jwt.Auth{
			ID:   existingUser.ID,
			Role: existingUser.Role,
		}

		userRepo.EXPECT().
			FindByEmail(ctx, req.Email).
			Return(existingUser, nil)

		jwtToken.EXPECT().
			Create(expectedAuth).
			Return("dummy.jwt.token", nil)

		redisRepo.EXPECT().
			SaveSession(ctx, "dummy.jwt.token", mock.Anything, jwt.TokenDuration).
			Return(nil)

		res, err := uc.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "dummy.jwt.token", res.Token)
	})

	t.Run("err_user_not_found", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		req := &dto.UserLoginRequest{
			Email:    "notfound@mail.com",
			Password: plainPassword,
		}

		userRepo.EXPECT().
			FindByEmail(ctx, req.Email).
			Return(nil, apperror.ErrRecordNotFound)

		res, err := uc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrWrongEmailOrPassword)
	})

	t.Run("err_find_by_email_failed", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		req := &dto.UserLoginRequest{
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		userRepo.EXPECT().
			FindByEmail(ctx, req.Email).
			Return(nil, errors.New("db error"))

		res, err := uc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "failed to find user by email")
	})

	t.Run("err_wrong_password", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		req := &dto.UserLoginRequest{
			Email:    "test@mail.com",
			Password: "wrongpassword",
		}

		existingUser := &entity.User{
			ID:       1,
			Email:    "test@mail.com",
			Password: hashedPassword,
			Role:     entity.RoleCustomer,
		}

		userRepo.EXPECT().
			FindByEmail(ctx, req.Email).
			Return(existingUser, nil)

		res, err := uc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrWrongEmailOrPassword)
	})

	t.Run("err_create_jwt_failed", func(t *testing.T) {
		uc, userRepo, _, jwtToken := setupUserUsecase(t)

		req := &dto.UserLoginRequest{
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		existingUser := &entity.User{
			ID:       1,
			Email:    "test@mail.com",
			Password: hashedPassword,
			Role:     entity.RoleCustomer,
		}

		userRepo.EXPECT().
			FindByEmail(ctx, req.Email).
			Return(existingUser, nil)

		jwtToken.EXPECT().
			Create(mock.Anything).
			Return("", errors.New("jwt signing error"))

		res, err := uc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "failed to create token")
	})

	t.Run("err_save_session_redis_failed", func(t *testing.T) {
		uc, userRepo, redisRepo, jwtToken := setupUserUsecase(t)

		req := &dto.UserLoginRequest{
			Email:    "test@mail.com",
			Password: plainPassword,
		}

		existingUser := &entity.User{
			ID:       1,
			Email:    "test@mail.com",
			Password: hashedPassword,
			Role:     entity.RoleCustomer,
		}

		userRepo.EXPECT().
			FindByEmail(ctx, req.Email).
			Return(existingUser, nil)

		jwtToken.EXPECT().
			Create(mock.Anything).
			Return("dummy.jwt.token", nil)

		redisRepo.EXPECT().
			SaveSession(ctx, "dummy.jwt.token", mock.Anything, jwt.TokenDuration).
			Return(errors.New("redis connection refused"))

		res, err := uc.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "failed to save token to redis")
	})
}

func TestUserUsecaseImpl_GetProfile(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)

	t.Run("success_get_profile", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		existingUser := &entity.User{
			ID:    userID,
			Name:  "Test User",
			Email: "test@mail.com",
		}

		userRepo.EXPECT().
			FindByID(ctx, userID).
			Return(existingUser, nil)

		res, err := uc.GetProfile(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, userID, res.ID)
		assert.Equal(t, existingUser.Name, res.Name)
		assert.Equal(t, existingUser.Email, res.Email)
	})

	t.Run("err_user_not_found", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		userRepo.EXPECT().
			FindByID(ctx, userID).
			Return(nil, apperror.ErrRecordNotFound)

		res, err := uc.GetProfile(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, apperror.ErrUserNotFound)
	})

	t.Run("err_find_by_id_failed", func(t *testing.T) {
		uc, userRepo, _, _ := setupUserUsecase(t)

		userRepo.EXPECT().
			FindByID(ctx, userID).
			Return(nil, errors.New("db error"))

		res, err := uc.GetProfile(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "failed to find user by id")
	})
}

func TestUserUsecaseImpl_Logout(t *testing.T) {
	ctx := context.Background()
	token := "dummy.jwt.token"

	t.Run("success_logout", func(t *testing.T) {
		uc, _, redisRepo, _ := setupUserUsecase(t)

		redisRepo.EXPECT().
			DeleteSession(ctx, token).
			Return(nil)

		err := uc.Logout(ctx, token)

		assert.NoError(t, err)
	})

	t.Run("err_delete_session_failed", func(t *testing.T) {
		uc, _, redisRepo, _ := setupUserUsecase(t)

		redisRepo.EXPECT().
			DeleteSession(ctx, token).
			Return(errors.New("redis error"))

		err := uc.Logout(ctx, token)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete session")
	})
}
