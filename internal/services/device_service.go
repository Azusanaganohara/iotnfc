package services

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"iot-ktp-api/internal/config"
	"iot-ktp-api/internal/models"
	"iot-ktp-api/internal/utils"
)

type DeviceService struct {
	db *gorm.DB
}

func NewDeviceService(db *gorm.DB) *DeviceService {
	return &DeviceService{db: db}
}

type ProvisionInput struct {
	HardwareID   string `json:"hardware_id"`
	DeviceName   string `json:"device_name" binding:"required,min=2"`
	ProvisionKey string `json:"provision_key"`
}

type ProvisionResult struct {
	NodeID     string `json:"node_id"`
	APIKey     string `json:"api_key"`
	DeviceName string `json:"device_name"`
	Mode       string `json:"mode"`
	Message    string `json:"message"`
}

type SetModeInput struct {
	Mode     string `json:"mode" binding:"required,oneof=active register inactive pending"`
	Location string `json:"location"`
}

func (s *DeviceService) Provision(input ProvisionInput) (*ProvisionResult, error) {
	cfg := config.Get()
	validProvisionKey := false
	if cfg.ProvisionSecret != "" && input.ProvisionKey == cfg.ProvisionSecret {
		validProvisionKey = true
	}
	if cfg.DeviceAPIKey != "" && input.ProvisionKey == cfg.DeviceAPIKey {
		validProvisionKey = true
	}
	if !validProvisionKey {
		return nil, errors.New("invalid provision key")
	}

	if strings.TrimSpace(input.HardwareID) != "" {
		var existing models.IotDevice
		if err := s.db.Where("hardware_id = ?", input.HardwareID).First(&existing).Error; err == nil {
			newAPIKey, err := utils.GenerateAPIKey()
			if err != nil {
				return nil, err
			}
			s.db.Model(&existing).Updates(map[string]interface{}{
				"api_key":     newAPIKey,
				"device_name": input.DeviceName,
			})
			return &ProvisionResult{
				NodeID:     existing.NodeID,
				APIKey:     newAPIKey,
				DeviceName: input.DeviceName,
				Mode:       string(existing.Mode),
				Message:    "Device re-provisioned with new API key",
			}, nil
		}
	}
	nodeID := utils.GenerateNodeID()
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	hardwareID := strings.TrimSpace(input.HardwareID)
	if hardwareID == "" {
		hardwareID = "PENDING-" + nodeID
	}

	device := &models.IotDevice{
		ID:           utils.GenerateUUID(),
		NodeID:       nodeID,
		HardwareID:   hardwareID,
		DeviceName:   input.DeviceName,
		APIKey:       apiKey,
		Mode:         models.DeviceModePending,
		RegisteredAt: &now,
	}

	if err := s.db.Create(device).Error; err != nil {
		return nil, err
	}

	return &ProvisionResult{
		NodeID:     nodeID,
		APIKey:     apiKey,
		DeviceName: device.DeviceName,
		Mode:       string(device.Mode),
		Message:    "Device provisioned. Waiting for admin activation.",
	}, nil
}

func (s *DeviceService) GetAll() ([]models.IotDevice, error) {
	var devices []models.IotDevice
	err := s.db.Preload("Activator").Order("created_at DESC").Find(&devices).Error
	return devices, err
}

func (s *DeviceService) GetByNodeID(nodeID string) (*models.IotDevice, error) {
	var device models.IotDevice
	if err := s.db.Preload("Activator").Where("node_id = ?", nodeID).First(&device).Error; err != nil {
		return nil, errors.New("device not found")
	}
	return &device, nil
}

func (s *DeviceService) SetMode(nodeID, userID string, input SetModeInput) (*models.IotDevice, error) {
	var device models.IotDevice
	if err := s.db.Where("node_id = ?", nodeID).First(&device).Error; err != nil {
		return nil, errors.New("device not found")
	}

	updates := map[string]interface{}{
		"mode":         models.DeviceMode(input.Mode),
		"activated_by": userID,
	}
	if input.Location != "" {
		updates["location"] = input.Location
	}

	if err := s.db.Model(&device).Updates(updates).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Activator").Where("node_id = ?", nodeID).First(&device)
	return &device, nil
}

func (s *DeviceService) UpdateDevice(nodeID string, name, location string) (*models.IotDevice, error) {
	var device models.IotDevice
	if err := s.db.Where("node_id = ?", nodeID).First(&device).Error; err != nil {
		return nil, errors.New("device not found")
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["device_name"] = name
	}
	if location != "" {
		updates["location"] = location
	}

	s.db.Model(&device).Updates(updates)
	s.db.Preload("Activator").Where("node_id = ?", nodeID).First(&device)
	return &device, nil
}

func (s *DeviceService) Delete(nodeID string) error {
	result := s.db.Where("node_id = ?", nodeID).Delete(&models.IotDevice{})
	if result.RowsAffected == 0 {
		return errors.New("device not found")
	}
	return result.Error
}

func (s *DeviceService) GetStatus(nodeID string) (*models.IotDevice, error) {
	return s.GetByNodeID(nodeID)
}
