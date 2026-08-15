package repository

import (
	"context"
	"errors"

	"github.com/Mpayy/e-commerce/monolith/internal/user/entity"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.GetTxFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.ErrDuplicatedKey
		}
		return err
	}
	return nil
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}
