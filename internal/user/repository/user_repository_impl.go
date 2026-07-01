package userrepository

import (
	"context"
	"errors"

	"github.com/Mpayy/e-commerce/internal/user/entity"
	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/transaction"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{DB: db}
}

func (r *UserRepositoryImpl) GetTx(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.GetTxFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return r.DB.WithContext(ctx)
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	err := r.GetTx(ctx).Create(user).Error
	if err != nil {
        var mysqlErr *mysql.MySQLError
        if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
            return apperror.ErrDuplicatedKey
        }
        return err
    }

	return nil
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.GetTx(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.GetTx(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
