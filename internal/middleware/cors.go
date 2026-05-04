package middleware

import (
	"fmt"
	"strings"
	"time"

	"iot-ktp-api/internal/config"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	cfg := config.Get()
	allowed := make(map[string]bool)
	for _, o := range strings.Split(cfg.AllowedOrigins, ",") {
		allowed[strings.TrimSpace(o)] = true
	}

	maxAge := fmt.Sprintf("%d", int(12*time.Hour/time.Second))

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key, X-Node-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", maxAge)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
