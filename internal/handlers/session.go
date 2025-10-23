package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/villageFlower/paypilot_dev_session_service/internal/database"
	"github.com/villageFlower/paypilot_dev_session_service/internal/models"
	"go.uber.org/zap"
)

// SessionHandler handles session-related requests
type SessionHandler struct {
	log *zap.Logger
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(log *zap.Logger) *SessionHandler {
	return &SessionHandler{log: log}
}

// CreateSession godoc
// @Summary Create a new session
// @Description Create a new session for a user
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body models.Session true "Session information"
// @Success 201 {object} models.Session
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var session models.Session
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate token if not provided
	if session.Token == "" {
		session.Token = uuid.New().String()
	}

	// Set expiration time if not provided (24 hours from now)
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = time.Now().Add(24 * time.Hour)
	}

	// Set IP address and user agent from request
	session.IPAddress = c.ClientIP()
	session.UserAgent = c.Request.UserAgent()

	if err := database.DB.Create(&session).Error; err != nil {
		h.log.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetSession godoc
// @Summary Get a session by ID
// @Description Get session details by session ID
// @Tags sessions
// @Accept json
// @Produce json
// @Param id path int true "Session ID"
// @Success 200 {object} models.Session
// @Failure 404 {object} map[string]interface{}
// @Router /sessions/{id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var session models.Session
	if err := database.DB.Preload("User").First(&session, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListSessions godoc
// @Summary List all sessions
// @Description Get a list of all sessions
// @Tags sessions
// @Accept json
// @Produce json
// @Param user_id query int false "Filter by user ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Session{}).Preload("User")
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var sessions []models.Session
	var total int64

	query.Count(&total)
	query.Limit(pageSize).Offset(offset).Find(&sessions)

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// DeleteSession godoc
// @Summary Delete a session
// @Description Delete a session by ID
// @Tags sessions
// @Accept json
// @Produce json
// @Param id path int true "Session ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Router /sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var session models.Session
	if err := database.DB.First(&session, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	if err := database.DB.Delete(&session).Error; err != nil {
		h.log.Error("Failed to delete session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}

	c.Status(http.StatusNoContent)
}
