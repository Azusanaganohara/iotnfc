package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"iot-ktp-api/internal/config"
	"iot-ktp-api/internal/database"
	"iot-ktp-api/internal/handlers"
	"iot-ktp-api/internal/middleware"
	"iot-ktp-api/internal/routes"
	"iot-ktp-api/internal/services"
)

func main() {
	cfg := config.Load()
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	db := database.Connect()
	database.Migrate(db)
	authSvc := services.NewAuthService(db)
	deviceSvc := services.NewDeviceService(db)
	memberSvc := services.NewMemberService(db)
	ktpSvc := services.NewKTPService(db)
	authH := handlers.NewAuthHandler(authSvc)
	deviceH := handlers.NewDeviceHandler(deviceSvc)
	memberH := handlers.NewMemberHandler(memberSvc)
	ktpH := handlers.NewKTPHandler(ktpSvc)
	rl := middleware.NewRateLimiter(100, time.Minute)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(rl.Middleware())

	r.LoadHTMLGlob("templates/*")

	routes.SetupAPI(r, db, authH, deviceH, memberH, ktpH)

	log.Printf("IoT KTP API running on port %s (env: %s)", cfg.Port, cfg.Env)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
