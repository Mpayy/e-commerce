package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/entity"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type UserRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (name, email, password, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			RETURNING id, created_at, updated_at`
	
	row := r.db.QueryRowxContext(ctx, query, user.Name, user.Email, user.Password, user.Role)
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperror.ErrDuplicatedKey
		}
		return err
	}
	return nil
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, name, email, password, role, created_at, updated_at 
			FROM users
			WHERE email = $1`

	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, name, email, password, role, created_at, updated_at 
			FROM users
			WHERE id = $1`

	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}