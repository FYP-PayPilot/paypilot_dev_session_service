package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/villageFlower/paypilot_dev_session_service/internal/database"
	"github.com/villageFlower/paypilot_dev_session_service/internal/handlers"
	"github.com/villageFlower/paypilot_dev_session_service/internal/messaging"
	"github.com/villageFlower/paypilot_dev_session_service/internal/middleware"
	"github.com/villageFlower/paypilot_dev_session_service/pkg/config"
	"github.com/villageFlower/paypilot_dev_session_service/pkg/logger"
	"go.uber.org/zap"

	_ "github.com/villageFlower/paypilot_dev_session_service/docs" // swagger docs
)

// @title PayPilot Dev Session Service API
// @version 1.0
// @description API for PayPilot development session service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := logger.Initialize(&cfg.Log); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting PayPilot Dev Session Service")

	// Initialize database
	if err := database.Initialize(&cfg.Database, logger.Log); err != nil {
		logger.Log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	// Run migrations
	if err := database.AutoMigrate(logger.Log); err != nil {
		logger.Log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize RabbitMQ
	rmq, err := messaging.NewRabbitMQ(&cfg.RabbitMQ, logger.Log)
	if err != nil {
		logger.Log.Warn("Failed to initialize RabbitMQ (service will continue without messaging)", zap.Error(err))
		rmq = nil
	} else {
		defer rmq.Close()
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger(logger.Log))
	router.Use(middleware.Recovery(logger.Log))
	router.Use(middleware.CORS())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	userHandler := handlers.NewUserHandler(logger.Log)
	sessionHandler := handlers.NewSessionHandler(logger.Log)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", healthHandler.Check)

		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Session routes
		sessions := v1.Group("/sessions")
		{
			sessions.POST("", sessionHandler.CreateSession)
			sessions.GET("", sessionHandler.ListSessions)
			sessions.GET("/:id", sessionHandler.GetSession)
			sessions.DELETE("/:id", sessionHandler.DeleteSession)
		}
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Log.Info("Server starting", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Start message consumer if RabbitMQ is available
	if rmq != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			err := rmq.Consume(ctx, func(body []byte) error {
				logger.Log.Info("Received message", zap.String("body", string(body)))
				return nil
			})
			if err != nil {
				logger.Log.Error("Consumer stopped", zap.Error(err))
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exited")
}
