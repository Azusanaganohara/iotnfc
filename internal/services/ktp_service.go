package services

import (
	"errors"
	"strings"
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
	UnixID string `json:"unix_id" binding:"required,numeric"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
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
	err := s.db.Where("unix_id = ?", input.UnixID).First(&member).Error

	log := &models.AccessLog{
		ID:     utils.GenerateUUID(),
		NodeID: device.NodeID,
		UnixID: input.UnixID,
	}

	if err != nil {
		log.Action = models.ActionDenied
		log.Reason = "Unix ID not registered as member"
		s.db.Create(log)
		return &TapResult{
			Action:     string(models.ActionDenied),
			Message:    "Access denied: Unix ID not registered",
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
	var member models.Member
	err := s.db.Where("unix_id = ?", input.UnixID).First(&member).Error

	log := &models.AccessLog{
		ID:     utils.GenerateUUID(),
		NodeID: device.NodeID,
		UnixID: input.UnixID,
	}

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		pending, pendingErr := s.getPendingRegistration(device.NodeID)
		if pendingErr != nil {
			log.Action = models.ActionDenied
			log.Reason = "No pending registration for device"
			s.db.Create(log)
			return &TapResult{
				Action:     string(models.ActionDenied),
				Message:    "Registration denied: no pending admin data",
				DeviceMode: string(device.Mode),
			}, nil
		}

		newMember := &models.Member{
			ID:                 utils.GenerateUUID(),
			UnixID:             input.UnixID,
			Name:               pending.Name,
			Phone:              pending.Phone,
			IsActive:           true,
			RegisteredByDevice: &device.NodeID,
			RegisteredByUser:   pending.RequestedBy,
		}
		if err := s.db.Create(newMember).Error; err != nil {
			return nil, err
		}
		s.db.Delete(&models.PendingRegistration{}, "id = ?", pending.ID)

		log.Action = models.ActionRegistered
		log.MemberID = &newMember.ID
		log.Reason = "Member created via register mode"
		s.db.Create(log)

		return &TapResult{
			Action:     string(models.ActionRegistered),
			Message:    "Card registered, member created",
			Member:     newMember,
			DeviceMode: string(device.Mode),
		}, nil
	}

	if member.IsActive {
		log.Action = models.ActionAlreadyRegistered
		log.MemberID = &member.ID
		log.Reason = "Member already active"
		s.db.Create(log)
		return &TapResult{
			Action:     string(models.ActionAlreadyRegistered),
			Message:    "Member already active",
			Member:     &member,
			DeviceMode: string(device.Mode),
		}, nil
	}

	if input.Name != "" && !strings.EqualFold(strings.TrimSpace(input.Name), strings.TrimSpace(member.Name)) {
		log.Action = models.ActionDenied
		log.MemberID = &member.ID
		log.Reason = "Member name mismatch"
		s.db.Create(log)
		return &TapResult{
			Action:     string(models.ActionDenied),
			Message:    "Registration denied: member data mismatch",
			DeviceMode: string(device.Mode),
		}, nil
	}

	if input.Phone != "" && normalizePhone(input.Phone) != normalizePhone(member.Phone) {
		log.Action = models.ActionDenied
		log.MemberID = &member.ID
		log.Reason = "Member phone mismatch"
		s.db.Create(log)
		return &TapResult{
			Action:     string(models.ActionDenied),
			Message:    "Registration denied: member data mismatch",
			DeviceMode: string(device.Mode),
		}, nil
	}

	if err := s.db.Model(&member).Updates(map[string]interface{}{
		"is_active":            true,
		"registered_by_device": device.NodeID,
	}).Error; err != nil {
		return nil, err
	}

	log.Action = models.ActionRegistered
	log.MemberID = &member.ID
	log.Reason = "Member activated via card registration"
	s.db.Create(log)

	member.IsActive = true
	member.RegisteredByDevice = &device.NodeID
	return &TapResult{
		Action:     string(models.ActionRegistered),
		Message:    "Card registered, member activated",
		Member:     &member,
		DeviceMode: string(device.Mode),
	}, nil
}

func (s *KTPService) getPendingRegistration(nodeID string) (*models.PendingRegistration, error) {
	var pending models.PendingRegistration
	if err := s.db.Where("device_node_id = ?", nodeID).Order("created_at DESC").First(&pending).Error; err != nil {
		return nil, err
	}
	if pending.ExpiresAt != nil && time.Now().After(*pending.ExpiresAt) {
		s.db.Delete(&pending)
		return nil, gorm.ErrRecordNotFound
	}
	return &pending, nil
}

func (s *KTPService) GetLogs(nodeID, unixID string, page, limit int) ([]models.AccessLog, int64, error) {
	var logs []models.AccessLog
	var total int64

	q := s.db.Model(&models.AccessLog{}).Preload("Member")
	if nodeID != "" {
		q = q.Where("node_id = ?", nodeID)
	}
	if unixID != "" {
		q = q.Where("unix_id = ?", unixID)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Order("tapped_at DESC").Find(&logs).Error
	return logs, total, err
}

func normalizePhone(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}
