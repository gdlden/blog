package model

import "time"

type VerificationCode struct {
	ID        uint       `gorm:"primarykey"`
	Email     string     `gorm:"column:email;type:varchar(255);not null;index"`
	Code      string     `gorm:"column:code;type:varchar(10);not null"`
	Type      string     `gorm:"column:type;type:varchar(20);not null;default:'register'"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	UsedAt    *time.Time `gorm:"column:used_at;default:null"`
	CreatedAt time.Time
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}
