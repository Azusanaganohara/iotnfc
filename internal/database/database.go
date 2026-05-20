package database

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"iot-ktp-api/internal/config"
	"iot-ktp-api/internal/models"
	"iot-ktp-api/internal/utils"
)

var DB *gorm.DB

func Connect() *gorm.DB {
	cfg := config.Get()

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	logLevel := logger.Silent
	if cfg.Env == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db
	log.Printf("Connected to MySQL database: %s", cfg.DBName)
	return db
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.IotDevice{},
		&models.Member{},
		&models.AccessLog{},
	)
	if err != nil {
		log.Fatalf("Auto migration failed: %v", err)
	}
	legacyMemberColumns := []string{"nik", "address", "birth_date"}
	for _, col := range legacyMemberColumns {
		if db.Migrator().HasColumn(&models.Member{}, col) {
			if err := db.Migrator().DropColumn(&models.Member{}, col); err != nil {
				log.Fatalf("Failed to drop members.%s column: %v", col, err)
			}
			log.Printf("Dropped members.%s column", col)
		}
	}
	log.Println("Database migration completed")

	seedAdmin(db)
}

func seedAdmin(db *gorm.DB) {
	cfg := config.Get()
	var count int64
	db.Model(&models.User{}).Count(&count)

	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), 12)
		if err != nil {
			log.Fatalf("Failed to hash admin password: %v", err)
		}

		admin := models.User{
			ID:       utils.GenerateUUID(),
			Name:     "Super Admin",
			Email:    cfg.AdminEmail,
			Password: string(hash),
			Role:     models.RoleAdmin,
			IsActive: true,
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		log.Printf("Created first admin user: %s", cfg.AdminEmail)
	}
}
