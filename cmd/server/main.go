package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kalbhor/tasqueue-ui/internal/api"
	"github.com/kalbhor/tasqueue-ui/internal/config"
	"github.com/kalbhor/tasqueue-ui/internal/service"
)

//go:embed all:web
var staticFS embed.FS

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse command line flags
	var (
		port       = flag.String("port", "8080", "HTTP server port")
		host       = flag.String("host", "0.0.0.0", "HTTP server host")
		brokerType = flag.String("broker", "redis", "Broker type (redis, nats-js, in-memory)")
		redisAddr  = flag.String("redis-addr", "localhost:6379", "Redis address")
		redisPass  = flag.String("redis-pass", "", "Redis password")
		redisDB    = flag.Int("redis-db", 0, "Redis database")
		showVer    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("Tasqueue UI\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Build Date: %s\n", date)
		os.Exit(0)
	}

	// Build configuration
	cfg := config.DefaultConfig()
	cfg.Server.Port = *port
	cfg.Server.Host = *host
	cfg.Broker.Type = *brokerType
	cfg.Broker.Redis.Addr = *redisAddr
	cfg.Broker.Redis.Password = *redisPass
	cfg.Broker.Redis.DB = *redisDB

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Starting Tasqueue UI server...")
	log.Printf("Broker: %s", cfg.Broker.Type)
	if cfg.Broker.Type == "redis" {
		log.Printf("Redis: %s (DB: %d)", cfg.Broker.Redis.Addr, cfg.Broker.Redis.DB)
	}

	// Initialize Tasqueue service
	svc, err := service.NewService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}
	log.Printf("Successfully connected to %s broker", cfg.Broker.Type)

	// Create API handler
	handler := api.NewHandler(svc)

	// Setup routes with embedded static files
	mux := api.SetupRoutes(handler, staticFS)

	// Wrap with middleware
	finalHandler := api.LoggingMiddleware(api.CORSMiddleware(mux))

	// Create HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
