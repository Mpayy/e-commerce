package seeder

import (
	"github.com/Mpayy/e-commerce/internal/user/entity"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RunSeeder(log *logrus.Logger, db *gorm.DB) {
	log.Info("Memulai proses database seeding...")

	seedAdmin(log, db)

	log.Info("Database seeding selesai dengan sukses!")
}

func seedAdmin(log *logrus.Logger, db *gorm.DB) {
	var count int64
	db.Model(&entity.User{}).Where("role = ?", entity.RoleAdmin).Count(&count)

	if count > 0 {
		log.Info("[Seeder] Admin data already exists, skip...")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("AdminRahasia123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("[Seeder] Failed to generate admin password hash: %v", err)
	}

	admin := entity.User{
		Name:     "Super Admin",
		Email:    "admin@mail.com",
		Password: string(hashedPassword),
		Role:     entity.RoleAdmin,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("[Seeder] Failed to create admin user: %v", err)
	}

	log.Info("[Seeder] Admin user created successfully!")
}
