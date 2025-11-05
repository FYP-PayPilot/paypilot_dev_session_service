package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/villageFlower/paypilot_dev_session_service/internal/database"
	"github.com/villageFlower/paypilot_dev_session_service/internal/kubernetes"
	"github.com/villageFlower/paypilot_dev_session_service/internal/models"
	"go.uber.org/zap"
)

// SessionHandler handles session-related requests
type SessionHandler struct {
	log      *zap.Logger
	k8sClient *kubernetes.Client
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(log *zap.Logger, k8sClient *kubernetes.Client) *SessionHandler {
	return &SessionHandler{
		log:      log,
		k8sClient: k8sClient,
	}
}

// CreateSession godoc
// @Summary Create a new development session
// @Description Create a new dev session for a project in the no-code app generator
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

	// Set expiration time if not provided (1 year for always-on sessions)
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = time.Now().Add(24 * 365 * time.Hour)
	}

	// Set IP address and user agent from request
	session.IPAddress = c.ClientIP()
	session.UserAgent = c.Request.UserAgent()

	// Set namespace from project UUID if not set
	if session.Namespace == "" && session.ProjectUUID != "" {
		session.Namespace = session.ProjectUUID
	}

	// Set default status if not provided
	if session.Status == "" {
		session.Status = "pending"
	}

	// Create the dev container in Kubernetes if k8s client is available
	if h.k8sClient != nil && session.ProjectUUID != "" {
		ctx := context.Background()
		endpoints, err := h.k8sClient.CreateDevContainer(ctx, session.ProjectUUID, session.ProjectID, session.UserID)
		if err != nil {
			h.log.Error("Failed to create dev container", zap.Error(err))
			session.Status = "error"
		} else {
			session.Status = "running"
			session.ContainerName = "dev-session-" + session.ProjectUUID
			// Populate service endpoints
			if endpoints != nil {
				session.IPAddress = endpoints.ClusterIP
				session.PreviewURL = endpoints.PreviewURL
				session.PreviewPath = endpoints.PreviewPath
				session.ChatURL = endpoints.ChatURL
				session.ChatPath = endpoints.ChatPath
				session.VscodeURL = endpoints.VscodeURL
				session.VscodePath = endpoints.VscodePath
			}
		}
	}

	if err := database.DB.Create(&session).Error; err != nil {
		h.log.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetSession godoc
// @Summary Get a dev session by ID
// @Description Get dev session details by session ID
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
	if err := database.DB.First(&session, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListSessions godoc
// @Summary List all dev sessions
// @Description Get a list of all dev sessions with optional filtering
// @Tags sessions
// @Accept json
// @Produce json
// @Param user_id query int false "Filter by user ID"
// @Param project_id query int false "Filter by project ID"
// @Param status query string false "Filter by status (pending, running, stopped, error)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID := c.Query("user_id")
	projectID := c.Query("project_id")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Session{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
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
// @Summary Delete a dev session
// @Description Delete a dev session by ID (also stops the associated container)
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

	// Delete from Kubernetes if k8s client is available
	if h.k8sClient != nil && session.ProjectUUID != "" {
		ctx := context.Background()
		if err := h.k8sClient.DeleteDevContainer(ctx, session.ProjectUUID); err != nil {
			h.log.Error("Failed to delete dev container from Kubernetes", zap.Error(err))
			// Continue with DB deletion even if K8s deletion fails
		}
	}

	if err := database.DB.Delete(&session).Error; err != nil {
		h.log.Error("Failed to delete session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetOrCreateSessionByProjectUUID godoc
// @Summary Get or create a dev session by project UUID
// @Description Get an existing session for a project UUID, or create a new one if it doesn't exist
// @Tags sessions
// @Accept json
// @Produce json
// @Param project_uuid path string true "Project UUID"
// @Param user_id query int false "User ID"
// @Param project_id query int false "Project ID"
// @Success 200 {object} models.Session
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sessions/project/{project_uuid} [get]
func (h *SessionHandler) GetOrCreateSessionByProjectUUID(c *gin.Context) {
	projectUUID := c.Param("project_uuid")
	if projectUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project UUID is required"})
		return
	}

	// Try to find existing session
	var session models.Session
	err := database.DB.Where("project_uuid = ? AND is_active = ?", projectUUID, true).First(&session).Error
	
	if err == nil {
		// Session exists, return it
		h.log.Info("Found existing session", zap.String("project_uuid", projectUUID))
		c.JSON(http.StatusOK, session)
		return
	}

	// Session doesn't exist, create a new one
	h.log.Info("Creating new session for project", zap.String("project_uuid", projectUUID))

	// Get user_id and project_id from query params or use defaults
	userID, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	projectID, _ := strconv.Atoi(c.DefaultQuery("project_id", "0"))

	// Create new session
	session = models.Session{
		UserID:      userID,
		ProjectID:   projectID,
		ProjectUUID: projectUUID,
		Token:       uuid.New().String(),
		ExpiresAt:   time.Now().Add(24 * 365 * time.Hour), // 1 year expiration for always-on sessions
		Namespace:   projectUUID, // Use project UUID as namespace
		Status:      "pending",
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		IsActive:    true,
	}

	// Create the dev container in Kubernetes if k8s client is available
	if h.k8sClient != nil {
		ctx := context.Background()
		endpoints, err := h.k8sClient.CreateDevContainer(ctx, projectUUID, projectID, userID)
		if err != nil {
			h.log.Error("Failed to create dev container", zap.Error(err))
			session.Status = "error"
		} else {
			session.Status = "running"
			session.ContainerName = "dev-session-" + projectUUID
			// Populate service endpoints
			if endpoints != nil {
				session.IPAddress = endpoints.ClusterIP
				session.PreviewURL = endpoints.PreviewURL
				session.PreviewPath = endpoints.PreviewPath
				session.ChatURL = endpoints.ChatURL
				session.ChatPath = endpoints.ChatPath
				session.VscodeURL = endpoints.VscodeURL
				session.VscodePath = endpoints.VscodePath
			}
		}
	}

	// Save to database
	if err := database.DB.Create(&session).Error; err != nil {
		h.log.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	h.log.Info("Session created successfully",
		zap.String("project_uuid", projectUUID),
		zap.Uint("session_id", session.ID))

	c.JSON(http.StatusOK, session)
}
