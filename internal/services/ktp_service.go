package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"iot-ktp-api/internal/models"
	"iot-ktp-api/internal/utils"
)

type KTPService struct {
	db *gorm.DB
}

func NewKTPService(db *gorm.DB) *KTPService {
	return &KTPService{db: db}
}

type TapInput struct {
	NIK  string `json:"nik" binding:"required,len=16,numeric"`
	Name string `json:"name"`
}

type TapResult struct {
	Action     string         `json:"action"`
	Message    string         `json:"message"`
	Member     *models.Member `json:"member,omitempty"`
	DeviceMode string         `json:"device_mode"`
}

func (s *KTPService) ProcessTap(device *models.IotDevice, input TapInput) (*TapResult, error) {
	switch device.Mode {
	case models.DeviceModeActive:
		return s.handleActiveTap(device, input)
	case models.DeviceModeRegister:
		return s.handleRegisterTap(device, input)
	case models.DeviceModePending:
		return nil, errors.New("device is pending activation by admin")
	default:
		return nil, errors.New("device is in an invalid mode")
	}
}

func (s *KTPService) handleActiveTap(device *models.IotDevice, input TapInput) (*TapResult, error) {
	var member models.Member
	err := s.db.Where("nik = ?", input.NIK).First(&member).Error

	log := &models.AccessLog{
		ID:     utils.GenerateUUID(),
		NodeID: device.NodeID,
		NIK:    input.NIK,
	}

	if err != nil {
		log.Action = models.ActionDenied
		log.Reason = "NIK not registered as member"
		s.db.Create(log)
		return &TapResult{
			Action:     string(models.ActionDenied),
			Message:    "Access denied: NIK not registered",
			DeviceMode: string(device.Mode),
		}, nil
	}

	if !member.IsActive {
		log.Action = models.ActionDenied
		log.Reason = "Member is inactive"
		log.MemberID = &member.ID
		s.db.Create(log)
		return &TapResult{
			Action:     string(models.ActionDenied),
			Message:    "Access denied: member account is inactive",
			DeviceMode: string(device.Mode),
		}, nil
	}

	log.Action = models.ActionGranted
	log.MemberID = &member.ID
	s.db.Create(log)

	return &TapResult{
		Action:     string(models.ActionGranted),
		Message:    "Access granted",
		Member:     &member,
		DeviceMode: string(device.Mode),
	}, nil
}

func (s *KTPService) handleRegisterTap(device *models.IotDevice, input TapInput) (*TapResult, error) {
	var existing models.Member
	err := s.db.Where("nik = ?", input.NIK).First(&existing).Error

	log := &models.AccessLog{
		ID:     utils.GenerateUUID(),
		NodeID: device.NodeID,
		NIK:    input.NIK,
	}

	if err == nil {
		log.Action = models.ActionAlreadyRegistered
		log.MemberID = &existing.ID
		log.Reason = "NIK already registered"
		s.db.Create(log)
		return &TapResult{
			Action:     string(models.ActionAlreadyRegistered),
			Message:    "NIK is already a registered member",
			Member:     &existing,
			DeviceMode: string(device.Mode),
		}, nil
	}
	name := input.Name
	if name == "" {
		name = "Member " + input.NIK[len(input.NIK)-4:]
	}

	now := time.Now()
	_ = now
	newMember := &models.Member{
		ID:                 utils.GenerateUUID(),
		NIK:                input.NIK,
		Name:               name,
		IsActive:           true,
		RegisteredByDevice: &device.NodeID,
	}

	if err := s.db.Create(newMember).Error; err != nil {
		return nil, err
	}

	log.Action = models.ActionRegistered
	log.MemberID = &newMember.ID
	s.db.Create(log)

	return &TapResult{
		Action:     string(models.ActionRegistered),
		Message:    "KTP registered as new member successfully",
		Member:     newMember,
		DeviceMode: string(device.Mode),
	}, nil
}

func (s *KTPService) GetLogs(nodeID, nik string, page, limit int) ([]models.AccessLog, int64, error) {
	var logs []models.AccessLog
	var total int64

	q := s.db.Model(&models.AccessLog{}).Preload("Member")
	if nodeID != "" {
		q = q.Where("node_id = ?", nodeID)
	}
	if nik != "" {
		q = q.Where("nik = ?", nik)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Order("tapped_at DESC").Find(&logs).Error
	return logs, total, err
}
