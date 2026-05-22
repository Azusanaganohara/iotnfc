package models

import "time"

type PendingRegistration struct {
	ID           string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	DeviceNodeID string     `gorm:"type:varchar(36);index;not null" json:"device_node_id"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	Phone        string     `gorm:"type:varchar(20);not null" json:"phone"`
	RequestedBy  *string    `gorm:"type:varchar(36)" json:"requested_by"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
