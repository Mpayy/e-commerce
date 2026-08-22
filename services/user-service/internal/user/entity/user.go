package entity

import "time"

const AuthPrefix = "auth:session:"

const (
	RoleCustomer string = "customer"
	RoleAdmin    string = "admin"
)

type User struct {
	ID        uint      `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
