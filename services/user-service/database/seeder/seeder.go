package seeder

import (
	"time"

	"github.com/Mpayy/e-commerce/pkg/logger"
	"github.com/Mpayy/e-commerce/services/user-service/internal/user/entity"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func RunSeeder(log *logger.Logger, db *sqlx.DB) {
	log.Info("Memulai proses database seeding...")

	seedAdmin(log, db)

	log.Info("Database seeding selesai dengan sukses!")
}

func seedAdmin(log *logger.Logger, db *sqlx.DB) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("AdminRahasia123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("[Seeder] Failed to generate admin password hash: %v", err)
	}

	admin := entity.User{
		Name:      "Super Admin",
		Email:     "admin@mail.com",
		Password:  string(hashedPassword),
		Role:      entity.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `INSERT INTO users (name, email, password, role, created_at, updated_at)
			VALUES (:name, :email, :password, :role, :created_at, :updated_at)
			ON CONFLICT (email) DO NOTHING`

	result, err := db.NamedExec(query, &admin)
	if err != nil {
		log.Fatalf("[Seeder] Failed to create admin user: %v", err)
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		log.Info("[Seeder] Admin user already exists, skip...")
		return
	}

	log.Info("[Seeder] Admin user created successfully!")
}
