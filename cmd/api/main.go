package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"inventory-system/internal/config"
	"inventory-system/internal/database"
	"inventory-system/internal/domain"
	"inventory-system/internal/handler"
	"inventory-system/internal/infrastructure"
	"inventory-system/internal/middleware"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"inventory-system/test/mocks"

	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Configurar modo de Gin
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
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

	// ========== Inicializar Repositorios ==========
	productRepo := repository.NewProductRepository(db)
	stockRepo := repository.NewStockRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// ========== Inicializar Event Publisher (Pub/Sub) ==========
	publisher, err := initializeEventPublisher(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize event publisher: %v", err)
	}
	defer publisher.Close()

	// ========== Inicializar Servicios ==========
	productService := service.NewProductService(productRepo, eventRepo)
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)
	reservationService := service.NewReservationService(reservationRepo, stockRepo, productRepo, eventRepo, publisher)
	eventSyncService := service.NewEventSyncService(eventRepo, publisher) // ✅ Inyectar publisher para re-intentos

	// ========== Inicializar Handlers ==========
	productHandler := handler.NewProductHandler(productService)
	stockHandler := handler.NewStockHandler(stockService)
	reservationHandler := handler.NewReservationHandler(reservationService)

	// ========== Crear Router ==========
	router := gin.New()

	// ========== Middlewares Globales ==========
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// ========== Health Check ==========
	router.GET("/health", func(c *gin.Context) {
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

	// ========== API v1 Routes ==========
	v1 := router.Group("/api/v1")
	{
		// Product endpoints (lectura pública, escritura protegida)
		products := v1.Group("/products")
		{
			// Públicos (sin API Key)
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/sku/:sku", productHandler.GetProductBySKU)

			// Protegidos (requieren API Key)
			products.POST("", middleware.APIKeyAuth(cfg.APIKeys), productHandler.CreateProduct)
			products.PUT("/:id", middleware.APIKeyAuth(cfg.APIKeys), productHandler.UpdateProduct)
			products.DELETE("/:id", middleware.APIKeyAuth(cfg.APIKeys), productHandler.DeleteProduct)
		}

		// Stock endpoints (todos protegidos)
		stock := v1.Group("/stock", middleware.APIKeyAuth(cfg.APIKeys))
		{
			stock.POST("", stockHandler.InitializeStock)
			stock.GET("/product/:productId", stockHandler.GetAllStockByProduct)
			stock.GET("/store/:storeId", stockHandler.GetAllStockByStore)
			stock.GET("/low-stock", stockHandler.GetLowStockItems)
			stock.GET("/:productId/:storeId", stockHandler.GetStockByProductAndStore)
			stock.GET("/:productId/:storeId/availability", stockHandler.CheckAvailability)
			stock.PUT("/:productId/:storeId", stockHandler.UpdateStock)
			stock.POST("/:productId/:storeId/adjust", stockHandler.AdjustStock)
		}

		// Stock transfer endpoint (protegido)
		v1.POST("/stock/transfer", middleware.APIKeyAuth(cfg.APIKeys), stockHandler.TransferStock)

		// Reservation endpoints (todos protegidos)
		reservations := v1.Group("/reservations", middleware.APIKeyAuth(cfg.APIKeys))
		{
			reservations.POST("", reservationHandler.CreateReservation)
			reservations.GET("/:id", reservationHandler.GetReservation)
			reservations.POST("/:id/confirm", reservationHandler.ConfirmReservation)
			reservations.POST("/:id/cancel", reservationHandler.CancelReservation)
			reservations.GET("/store/:storeId/pending", reservationHandler.GetPendingByStore)
			reservations.GET("/product/:productId/store/:storeId", reservationHandler.GetReservationsByProduct)
			reservations.GET("/stats", reservationHandler.GetReservationStats)
		}
	}

	// ========== Background Workers ==========
	// Worker para expirar reservas (cada 1 minuto)
	go startReservationExpirationWorker(reservationService)

	// Worker para sincronizar eventos (cada 10 segundos)
	go startEventSyncWorker(eventSyncService)

	// ========== Servidor HTTP ==========
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s (instance: %s)", cfg.ServerPort, cfg.InstanceID)
		log.Printf("📊 Database driver: %s", cfg.DatabaseDriver)
		log.Printf("🔒 Log level: %s, format: %s", cfg.LogLevel, cfg.LogFormat)
		log.Printf("� API Keys loaded: %d", len(cfg.APIKeys))
		log.Printf("��📡 API available at http://localhost:%s/api/v1", cfg.ServerPort)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// ========== Graceful Shutdown ==========
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

// initializeEventPublisher crea el publisher según la configuración.
// Soporta múltiples implementaciones: redis, kafka, none.
func initializeEventPublisher(cfg *config.Config) (domain.EventPublisher, error) {
	broker := strings.ToLower(cfg.MessageBroker)

	switch broker {
	case "redis":
		// Redis Streams (opción por defecto - simple y rápida)
		addr := fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
		publisher, err := infrastructure.NewRedisPublisher(infrastructure.RedisPublisherConfig{
			Addr:       addr,
			StreamName: "inventory-events",
			MaxLen:     100000, // Retener últimos 100k eventos
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis publisher: %w", err)
		}
		log.Printf("✅ Using Redis Streams as message broker (%s)", addr)
		return publisher, nil

	case "kafka":
		// Implementación futura para Apache Kafka
		return nil, fmt.Errorf("Kafka publisher not implemented yet. Set MESSAGE_BROKER=redis")

	case "none", "":
		// No publisher (solo logging)
		log.Printf("⚠️  No message broker configured (MESSAGE_BROKER=none)")
		return mocks.NewNoOpPublisher(), nil

	default:
		return nil, fmt.Errorf("unknown message broker: %s (options: redis, kafka, none)", broker)
	}
}

// startReservationExpirationWorker worker para expirar reservas
func startReservationExpirationWorker(service *service.ReservationService) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("⏰ Reservation expiration worker started")

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		count, err := service.ProcessExpiredReservations(ctx)
		cancel()

		if err != nil {
			log.Printf("Error processing expired reservations: %v", err)
		} else if count > 0 {
			log.Printf("✅ Expired %d reservations", count)
		}
	}
}

// startEventSyncWorker worker para sincronizar eventos
func startEventSyncWorker(service *service.EventSyncService) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("📡 Event synchronization worker started")

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		count, err := service.SyncPendingEvents(ctx, 100)
		cancel()

		if err != nil {
			log.Printf("Error syncing events: %v", err)
		} else if count > 0 {
			log.Printf("✅ Synced %d events", count)
		}
	}
}
