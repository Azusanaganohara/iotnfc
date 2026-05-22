package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"iot-ktp-api/internal/config"
	"iot-ktp-api/internal/models"
)

func DeviceAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		nodeID := c.GetHeader("X-Node-ID")

		if apiKey == "" || nodeID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Header X-API-Key dan X-Node-ID wajib diisi",
				"hint":    "Set X-API-Key dari .env (DEVICE_API_KEY) dan X-Node-ID dari provisioning",
			})
			return
		}

		cfg := config.Get()

		if cfg.DeviceMasterKey != "" && apiKey == cfg.DeviceMasterKey {
			var device models.IotDevice
			if err := db.Where("node_id = ?", nodeID).First(&device).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Node ID tidak ditemukan",
				})
				return
			}
			attachDevice(c, db, &device)
			return
		}

		if cfg.DeviceAPIKey != "" && apiKey == cfg.DeviceAPIKey {
			var device models.IotDevice
			if err := db.Where("node_id = ?", nodeID).First(&device).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Node ID tidak ditemukan. Pastikan device sudah di-provision.",
				})
				return
			}
			if device.Mode == models.DeviceModeInactive {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "Device dinonaktifkan. Hubungi administrator.",
				})
				return
			}
			attachDevice(c, db, &device)
			return
		}

		var device models.IotDevice
		if err := db.Where("node_id = ? AND api_key = ?", nodeID, apiKey).First(&device).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "API Key atau Node ID tidak valid",
				"hint":    "Gunakan DEVICE_API_KEY dari .env atau api_key hasil provisioning",
			})
			return
		}

		if device.Mode == models.DeviceModeInactive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Device dinonaktifkan. Hubungi administrator.",
			})
			return
		}

		attachDevice(c, db, &device)
	}
}

func attachDevice(c *gin.Context, db *gorm.DB, device *models.IotDevice) {
	now := time.Now()
	_ = db.Model(device).Update("last_seen_at", now)

	hardwareID := strings.TrimSpace(c.GetHeader("X-Hardware-ID"))
	if hardwareID != "" && (device.HardwareID == "" || strings.HasPrefix(device.HardwareID, "PENDING-")) {
		if err := db.Model(device).Update("hardware_id", hardwareID).Error; err == nil {
			device.HardwareID = hardwareID
		}
	}

	c.Set("device", device)
	c.Set("device_node_id", device.NodeID)
	c.Set("device_mode", string(device.Mode))
	c.Next()
}
