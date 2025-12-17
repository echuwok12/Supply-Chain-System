package main

import (
	"appointment-booking/internal/config"
	"appointment-booking/internal/domain"
	"appointment-booking/internal/handler"
	"appointment-booking/internal/middleware"
	"appointment-booking/internal/repository"
	"appointment-booking/internal/service"
	"appointment-booking/internal/websocket"

	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// --------------------
	// Config & DB
	// --------------------
	cfg := config.LoadConfig()

	db := config.ConnectDB(cfg)

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Appointment{},
		&domain.Availability{},
	); err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Database migration completed")

	wsHandler := websocket.NewHandler()
	go wsHandler.Run()

	// --------------------
	// Redis
	// --------------------
	redisClient := config.ConnectRedis(cfg)

	// --------------------
	// Repositories
	// --------------------
	userRepo := repository.NewUserRepository(db)
	apptRepo := repository.NewAppointmentRepository(db)
	availRepo := repository.NewAvailabilityRepository(db)

	// --------------------
	// Services
	// --------------------
	authService := service.NewAuthService(userRepo)

	notifyService := service.NewNotificationService()
	apptService := service.NewAppointmentService(
		apptRepo,
		notifyService,
		wsHandler,
		redisClient,
	)

	availService := service.NewAvailabilityService(availRepo, apptRepo, redisClient)

	reportService := service.NewReportService(apptRepo)
	// --------------------
	// Handlers
	// --------------------
	authHandler := handler.NewAuthHandler(authService)
	apptHandler := handler.NewAppointmentHandler(apptService)
	availHandler := handler.NewAvailabilityHandler(availService)
	adminHandler := handler.NewAdminHandler(reportService)

	// --------------------
	// Router
	// --------------------
	router := gin.Default()

	// WebSocket route
	router.GET("/ws", wsHandler.HandleConnection)

	// Public routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"env":    cfg.AppEnv,
		})
	})

	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)

	router.GET("/providers/:providerID/slots", availHandler.GetSlots)

	// --------------------
	// Protected routes
	// --------------------
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			role, _ := c.Get("role")

			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"role":    role,
				"message": "You are authenticated!",
			})
		})

		protected.POST("/appointments", apptHandler.Create)
		protected.PUT("/appointments/:id/cancel", apptHandler.Cancel)
		protected.PUT("/appointments/:id/reschedule", apptHandler.Reschedule)

		protected.POST("/availability", availHandler.SetAvailability)
	}

	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.RequireRole("admin"))
	{
		adminGroup.GET("/dashboard", adminHandler.GetDashboard)
	}

	// --------------------
	// HTTP Server
	// --------------------
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// --------------------
	// Graceful Shutdown
	// --------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
