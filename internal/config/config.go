package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	ServerPort string
	InstanceID string // Identificador de esta instancia de API (para logs/métricas)

	// Database
	DatabaseDriver   string // "postgres" o "sqlite"
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	SQLitePath       string // Para SQLite: ":memory:" o ruta a archivo

	// Redis
	RedisHost string
	RedisPort int

	// NATS
	NATSUrl string

	// Business
	ReservationTTL int // segundos

	// Security (API Key Authentication)
	APIKeys           map[string]string // key -> store_name
	RateLimitRequests int               // requests per minute

	// Observability
	LogLevel      string // debug, info, warn, error
	LogFormat     string // json, text
	EnableMetrics bool
}

func Load() *Config {
	// Cargar .env si existe (ignora error si no existe)
	_ = godotenv.Load()

	postgresPort, _ := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	reservationTTL, _ := strconv.Atoi(getEnv("RESERVATION_TTL", "600"))
	rateLimitRequests, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	enableMetrics, _ := strconv.ParseBool(getEnv("ENABLE_METRICS", "true"))

	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		InstanceID:        getEnv("INSTANCE_ID", "api-001"),
		DatabaseDriver:    getEnv("DATABASE_DRIVER", "sqlite"), // Default: SQLite para desarrollo
		PostgresHost:      getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:      postgresPort,
		PostgresUser:      getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:  getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:        getEnv("POSTGRES_DB", "inventory"),
		SQLitePath:        getEnv("SQLITE_PATH", ":memory:"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         redisPort,
		NATSUrl:           getEnv("NATS_URL", "nats://localhost:4222"),
		ReservationTTL:    reservationTTL,
		APIKeys:           loadAPIKeys(),
		RateLimitRequests: rateLimitRequests,
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),
		EnableMetrics:     enableMetrics,
	}
}

func loadAPIKeys() map[string]string {
	keys := make(map[string]string)

	// Leer de variable de entorno API_KEYS
	// Formato: key1:name1,key2:name2
	apiKeysEnv := getEnv("API_KEYS", "")
	if apiKeysEnv != "" {
		pairs := strings.Split(apiKeysEnv, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				keys[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Si no hay keys configuradas, usar keys por defecto para desarrollo
	if len(keys) == 0 {
		keys = map[string]string{
			"dev-key-store-001": "Store Madrid",
			"dev-key-store-002": "Store Barcelona",
			"dev-key-admin":     "Admin",
		}
	}

	return keys
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
