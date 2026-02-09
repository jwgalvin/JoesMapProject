# Step 08: Configuration and Main Application

## Objective
Implement configuration loading and wire all components together in main.go with graceful shutdown.

## Tasks

### 1. Create Configuration Package
Create `internal/config/config.go`:
```go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	USGS     USGSConfig
	Logging  LoggingConfig
	CORS     CORSConfig
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Path string
}

// USGSConfig contains USGS API configuration
type USGSConfig struct {
	Endpoint     string
	PollInterval time.Duration
	Timeout      time.Duration
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level string
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	Enabled        bool
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DATABASE_PATH", "./data/geopulse.db"),
		},
		USGS: USGSConfig{
			Endpoint:     getEnv("USGS_ENDPOINT", "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson"),
			PollInterval: getDurationEnv("USGS_POLL_INTERVAL", 5*time.Minute),
			Timeout:      getDurationEnv("USGS_TIMEOUT", 10*time.Second),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		CORS: CORSConfig{
			Enabled:        getBoolEnv("ENABLE_CORS", true),
			AllowedOrigins: getSliceEnv("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
	}

	return config, nil
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv gets boolean environment variable with default
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// getDurationEnv gets duration environment variable with default
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		// Try parsing as minutes first
		if minutes, err := strconv.Atoi(value); err == nil {
			return time.Duration(minutes) * time.Minute
		}
		// Try parsing as duration string
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

// getSliceEnv gets slice environment variable with default
func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.Database.Path == "" {
		return fmt.Errorf("DATABASE_PATH is required")
	}
	if c.USGS.Endpoint == "" {
		return fmt.Errorf("USGS_ENDPOINT is required")
	}
	return nil
}
```

### 2. Update Environment Example
Update `configs/.env.example`:
```env
# Server Configuration
PORT=8080

# Database Configuration
DATABASE_PATH=./data/geopulse.db

# USGS API Configuration
USGS_ENDPOINT=https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson
USGS_POLL_INTERVAL=5
USGS_TIMEOUT=10s

# Logging
LOG_LEVEL=info

# CORS
ENABLE_CORS=true
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### 3. Implement Main Application
Update `cmd/api/main.go`:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/geopulse/internal/application/ingest"
	"github.com/yourusername/geopulse/internal/application/query"
	"github.com/yourusername/geopulse/internal/config"
	"github.com/yourusername/geopulse/internal/infrastructure/persistence"
	httpinfra "github.com/yourusername/geopulse/internal/infrastructure/http"
	"github.com/yourusername/geopulse/internal/infrastructure/usgs"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Setup logger
	logger := setupLogger(cfg.Logging.Level)
	logger.Info("starting GeoPulse API",
		"port", cfg.Server.Port,
		"db_path", cfg.Database.Path,
		"poll_interval", cfg.USGS.PollInterval,
	)

	// Initialize database
	db, err := persistence.NewDatabase(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	logger.Info("database connected")

	// Initialize repository
	repo := persistence.NewSQLiteRepository(db)

	// Initialize USGS client
	usgsClient := usgs.NewClient(cfg.USGS.Endpoint, cfg.USGS.Timeout)

	// Initialize services
	ingestService := ingest.NewService(usgsClient, repo, logger)
	queryService := query.NewService(repo)

	// Start background ingestion scheduler
	scheduler := ingest.NewScheduler(ingestService, cfg.USGS.PollInterval, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go scheduler.Start(ctx)

	// Setup HTTP server
	handler := httpinfra.NewHandler(queryService)
	
	var corsOrigins []string
	if cfg.CORS.Enabled {
		corsOrigins = cfg.CORS.AllowedOrigins
	}
	
	router := httpinfra.NewRouter(handler, logger, corsOrigins)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)

		// Stop scheduler
		scheduler.Stop()

		// Graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			return server.Close()
		}

		logger.Info("server stopped gracefully")
	}

	return nil
}

// setupLogger creates a configured logger
func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}
```

*Note: Update import paths to match your module*

### 4. Copy Environment Example
```powershell
Copy-Item configs\.env.example .env
```

Edit `.env` if needed for local development.

### 5. Build and Test Application
```powershell
# Build the application
go build -o geopulse.exe ./cmd/api

# Run the application
.\geopulse.exe
```

In another terminal, test the API:
```powershell
# Test health endpoint
curl http://localhost:8080/v1/health

# Test events endpoint (after a few minutes for data ingestion)
curl http://localhost:8080/v1/events?limit=5

# Test GeoJSON endpoint
curl http://localhost:8080/v1/events/geojson?minMagnitude=4.0
```

### 6. Test Graceful Shutdown
Run the application and press `Ctrl+C`. You should see:
```
shutdown signal received
scheduler stopped
server stopped gracefully
```

### 7. Create Run Script
Create `scripts/run.ps1`:
```powershell
#!/usr/bin/env pwsh

# Load environment and run application
$ErrorActionPreference = "Stop"

Write-Host "Starting GeoPulse API..." -ForegroundColor Green

# Ensure data directory exists
if (-not (Test-Path "data")) {
    New-Item -ItemType Directory -Path "data" | Out-Null
}

# Run application
go run ./cmd/api
```

Make it executable and test:
```powershell
.\scripts\run.ps1
```

## Success Criteria
- ✓ Configuration loading from environment
- ✓ All components wired together
- ✓ Graceful shutdown implemented
- ✓ Background scheduler running
- ✓ HTTP server starts successfully
- ✓ API endpoints respond correctly
- ✓ Logs are structured and informative

## Next Step
Proceed to **Step09-Testing.md** to add comprehensive integration tests.
