package models

import (
	"time"

	"gorm.io/gorm"
)

// Session represents a user session
type Session struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"user_id" binding:"required"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
}

// TableName overrides the table name
func (Session) TableName() string {
	return "sessions"
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
