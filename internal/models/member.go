package models

import "time"

type Member struct {
	ID                 string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UnixID             string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"unix_id"`
	Name               string    `gorm:"type:varchar(100);not null" json:"name"`
	Phone              string    `gorm:"type:varchar(20)" json:"phone"`
	PhotoURL           string    `gorm:"type:varchar(500)" json:"photo_url"`
	IsActive           bool      `gorm:"default:true" json:"is_active"`
	RegisteredByDevice *string   `gorm:"type:varchar(36)" json:"registered_by_device"`
	RegisteredByUser   *string   `gorm:"type:varchar(36)" json:"registered_by_user"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
