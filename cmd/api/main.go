package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-system/internal/config"
	"inventory-system/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Configurar modo de Gin
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	//Initialize database
	db, err := database.NewDatabaseClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Aplicar migraciones
	log.Println("📊 Applying database migrations...")
	if err := database.InitializeSchema(db, cfg); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}
	log.Println("✅ Database migrations applied successfully")

	// Crear router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Check database health
		dbStatus := "healthy"
		if err := database.HealthCheck(db); err != nil {
			dbStatus = "unhealthy: " + err.Error()
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"timestamp":   time.Now().Format(time.RFC3339),
			"instance_id": cfg.InstanceID,
			"version":     "1.0.0",
			"database":    dbStatus,
			"db_driver":   cfg.DatabaseDriver,
		})
	})

	// Servidor HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s (instance: %s)", cfg.ServerPort, cfg.InstanceID)
		log.Printf("📊 Database driver: %s", cfg.DatabaseDriver)
		log.Printf("🔒 Log level: %s, format: %s", cfg.LogLevel, cfg.LogFormat)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
