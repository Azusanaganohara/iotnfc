package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"iot-ktp-api/internal/config"
	"iot-ktp-api/internal/models"
)

func main() {
	cfg := config.Get()
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("ded0d83d514f5f500e1a7c58"), 12)
    res := db.Model(&models.User{}).Where("email = ?", "admin@example.com").Update("password", string(hash))
	
	fmt.Printf("Updated %d rows\n", res.RowsAffected)
}
