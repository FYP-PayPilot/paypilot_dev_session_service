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
	ProjectUUID   string         `gorm:"uniqueIndex;not null" json:"project_uuid" binding:"required"` // UUID from another service
	Token         string         `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt     time.Time      `json:"expires_at"`
	ContainerName string         `json:"container_name"`
	Namespace     string         `json:"namespace"`                       // Uses project_uuid as namespace
	Status        string         `gorm:"default:'pending'" json:"status"` // pending, running, stopped, error
	IPAddress     string         `json:"ip_address"`
	UserAgent     string         `json:"user_agent"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	// Service endpoints
	PreviewURL  string `json:"preview_url"`  // Preview application endpoint
	PreviewPath string `json:"preview_path"` // Path redirect for preview
	ChatURL     string `json:"chat_url"`     // Chat/AI agents endpoint
	ChatPath    string `json:"chat_path"`    // Path redirect for chat
	VscodeURL   string `json:"vscode_url"`   // VS Code web endpoint
	VscodePath  string `json:"vscode_path"`  // Path redirect for vscode
}

// TableName overrides the table name
func (Session) TableName() string {
	return "sessions"
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
