package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"iot-ktp-api/internal/handlers"
	"iot-ktp-api/internal/middleware"
)

func SetupAPI(
	r *gin.Engine,
	db *gorm.DB,
	authH *handlers.AuthHandler,
	deviceH *handlers.DeviceHandler,
	memberH *handlers.MemberHandler,
	ktpH *handlers.KTPHandler,
) {
	// Forbidden handlers for direct access
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "forbidden.html", nil)
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusForbidden, "forbidden.html", nil)
	})

	r.GET("/docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "docs.html", nil)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "iot-ktp-api",
			"version": "1.0.0",
		})
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.POST("/refresh", authH.Refresh)
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthRequired())
			{
				authProtected.GET("/me", authH.Me)
				authProtected.POST("/logout", authH.Logout)
				authProtected.POST("/users", middleware.AdminRequired(), authH.Register)
			}
		}

		devices := v1.Group("/devices")
		{
			devices.POST("/provision", deviceH.Provision)
			deviceAuth := devices.Group("")
			deviceAuth.Use(middleware.DeviceAuth(db))
			{
				deviceAuth.GET("/me/status", deviceH.GetMyStatus)
			}

			adminDevices := devices.Group("")
			adminDevices.Use(middleware.AuthRequired())
			{
				adminDevices.GET("", deviceH.GetAll)
				adminDevices.GET("/:node_id", deviceH.GetOne)
				adminDevices.PUT("/:node_id", deviceH.Update)
				adminDevices.PUT("/:node_id/mode", middleware.AdminRequired(), deviceH.SetMode)
				adminDevices.DELETE("/:node_id", middleware.AdminRequired(), deviceH.Delete)
			}
		}
		members := v1.Group("/members")
		members.Use(middleware.AuthRequired())
		{
			members.GET("", memberH.GetAll)
			members.POST("", memberH.Create)
			members.GET("/unix/:unix_id", memberH.GetByUnixID)
			members.GET("/:id", memberH.GetByID)
			members.PUT("/:id", memberH.Update)
			members.DELETE("/:id", middleware.AdminRequired(), memberH.Delete)
		}
		card := v1.Group("/card")
		{
			cardDevice := card.Group("")
			cardDevice.Use(middleware.DeviceAuth(db))
			{
				cardDevice.POST("/tap", ktpH.Tap)
			}
			cardAdmin := card.Group("")
			cardAdmin.Use(middleware.AuthRequired())
			{
				cardAdmin.GET("/logs", ktpH.GetLogs)
				cardAdmin.GET("/logs/device/:node_id", ktpH.GetLogsByDevice)
				cardAdmin.GET("/logs/member/:unix_id", ktpH.GetLogsByUnixID)
			}
		}
		ktp := v1.Group("/ktp")
		{
			ktpDevice := ktp.Group("")
			ktpDevice.Use(middleware.DeviceAuth(db))
			{
				ktpDevice.POST("/tap", ktpH.Tap)
			}
			ktpAdmin := ktp.Group("")
			ktpAdmin.Use(middleware.AuthRequired())
			{
				ktpAdmin.GET("/logs", ktpH.GetLogs)
				ktpAdmin.GET("/logs/device/:node_id", ktpH.GetLogsByDevice)
				ktpAdmin.GET("/logs/member/:unix_id", ktpH.GetLogsByUnixID)
			}
		}
	}
}
