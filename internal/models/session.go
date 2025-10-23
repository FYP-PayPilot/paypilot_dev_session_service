package models

import (
	"time"

	"gorm.io/gorm"
)

// Session represents a development session for a no-code app project
// This manages dev containers in Kubernetes namespaces
type Session struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	UserID        int            `gorm:"not null;index" json:"user_id" binding:"required"`
	ProjectID     int            `gorm:"not null;index" json:"project_id" binding:"required"`
	Token         string         `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt     time.Time      `json:"expires_at"`
	ContainerName string         `json:"container_name"`
	Namespace     string         `json:"namespace"`
	Status        string         `gorm:"default:'pending'" json:"status"` // pending, running, stopped, error
	IPAddress     string         `json:"ip_address"`
	UserAgent     string         `json:"user_agent"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
}

// TableName overrides the table name
func (Session) TableName() string {
	return "sessions"
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
