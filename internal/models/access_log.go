package models

import "time"

type LogAction string

const (
	ActionGranted           LogAction = "granted"
	ActionDenied            LogAction = "denied"
	ActionRegistered        LogAction = "registered"
	ActionAlreadyRegistered LogAction = "already_registered"
)

type AccessLog struct {
	ID       string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	NodeID   string    `gorm:"type:varchar(36);not null;index" json:"node_id"`
	NIK      string    `gorm:"type:varchar(16);not null;index" json:"nik"`
	MemberID *string   `gorm:"type:varchar(36)" json:"member_id"`
	Action   LogAction `gorm:"type:enum('granted','denied','registered','already_registered');not null" json:"action"`
	Reason   string    `gorm:"type:varchar(255)" json:"reason"`
	TappedAt time.Time `gorm:"autoCreateTime" json:"tapped_at"`

	Member *Member `gorm:"foreignKey:MemberID;references:ID" json:"member,omitempty"`
}
