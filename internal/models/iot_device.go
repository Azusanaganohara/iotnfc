package models

import "time"

type DeviceMode string

const (
	DeviceModePending  DeviceMode = "pending"
	DeviceModeActive   DeviceMode = "active"
	DeviceModeRegister DeviceMode = "register"
	DeviceModeInactive DeviceMode = "inactive"
)

type IotDevice struct {
	ID           string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	NodeID       string     `gorm:"type:varchar(36);uniqueIndex;not null" json:"node_id"`
	HardwareID   string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"hardware_id"`
	DeviceName   string     `gorm:"type:varchar(100)" json:"device_name"`
	APIKey       string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	Mode         DeviceMode `gorm:"type:enum('pending','active','register','inactive');default:'pending'" json:"mode"`
	Location     string     `gorm:"type:varchar(200)" json:"location"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	RegisteredAt *time.Time `json:"registered_at"`
	ActivatedBy  *string    `gorm:"type:varchar(36)" json:"activated_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Activator *User `gorm:"foreignKey:ActivatedBy;references:ID" json:"activator,omitempty"`
}
