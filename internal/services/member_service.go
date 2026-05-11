package services

import (
	"errors"

	"gorm.io/gorm"

	"iot-ktp-api/internal/models"
	"iot-ktp-api/internal/utils"
)

type MemberService struct {
	db *gorm.DB
}

func NewMemberService(db *gorm.DB) *MemberService {
	return &MemberService{db: db}
}

type CreateMemberInput struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Phone    string `json:"phone" binding:"required,min=8,max=20"`
	PhotoURL string `json:"photo_url"`
}

type UpdateMemberInput struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	PhotoURL string `json:"photo_url"`
	IsActive *bool  `json:"is_active"`
}

func (s *MemberService) Create(input CreateMemberInput, registeredByUser string) (*models.Member, error) {
	unixID := utils.GenerateUnixID()
	member := &models.Member{
		ID:               utils.GenerateUUID(),
		UnixID:           unixID,
		Name:             input.Name,
		Phone:            input.Phone,
		PhotoURL:         input.PhotoURL,
		IsActive:         false,
		RegisteredByUser: &registeredByUser,
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, err
	}
	return member, nil
}

func (s *MemberService) GetAll(page, limit int, search string) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	q := s.db.Model(&models.Member{})
	if search != "" {
		q = q.Where("unix_id LIKE ? OR name LIKE ? OR phone LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	q.Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&members).Error
	return members, total, err
}

func (s *MemberService) GetByID(id string) (*models.Member, error) {
	var member models.Member
	if err := s.db.First(&member, "id = ?", id).Error; err != nil {
		return nil, errors.New("member not found")
	}
	return &member, nil
}

func (s *MemberService) GetByUnixID(unixID string) (*models.Member, error) {
	var member models.Member
	if err := s.db.Where("unix_id = ?", unixID).First(&member).Error; err != nil {
		return nil, errors.New("member not found")
	}
	return &member, nil
}

func (s *MemberService) Update(id string, input UpdateMemberInput) (*models.Member, error) {
	var member models.Member
	if err := s.db.First(&member, "id = ?", id).Error; err != nil {
		return nil, errors.New("member not found")
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Phone != "" {
		updates["phone"] = input.Phone
	}
	if input.PhotoURL != "" {
		updates["photo_url"] = input.PhotoURL
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if err := s.db.Model(&member).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (s *MemberService) Delete(id string) error {
	result := s.db.Delete(&models.Member{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return errors.New("member not found")
	}
	return result.Error
}
