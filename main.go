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

	"github.com/zupolgec/curl-impersonate-service/config"
	"github.com/zupolgec/curl-impersonate-service/handlers"
	"github.com/zupolgec/curl-impersonate-service/metrics"
	"github.com/zupolgec/curl-impersonate-service/middleware"
	"github.com/zupolgec/curl-impersonate-service/models"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting curl-impersonate-service v%s", config.Version)
	log.Printf("Port: %s", cfg.Port)
	log.Printf("Log Level: %s", cfg.LogLevel)

	// Load browsers configuration
	if err := models.LoadBrowsers(cfg.BrowsersJSONPath); err != nil {
		log.Fatalf("Failed to load browsers.json: %v", err)
	}
	log.Printf("Loaded %d browser configurations", len(models.GetAllBrowsers()))

	// Verify curl-impersonate binaries exist
	if err := verifyBinaries(); err != nil {
		log.Fatalf("Binary verification failed: %v", err)
	}
	log.Printf("Verified curl-impersonate binaries")

	// Initialize metrics collector
	collector := metrics.NewCollector()

	// Setup HTTP router
	mux := http.NewServeMux()

	// Public endpoint (no auth)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Protected endpoints
	authMw := middleware.AuthMiddleware(cfg.Token)
	mux.Handle("/browsers", authMw(http.HandlerFunc(handlers.BrowsersHandler)))
	mux.Handle("/metrics", authMw(handlers.NewMetricsHandler(collector)))
	mux.Handle("/impersonate", authMw(handlers.NewImpersonateHandler(cfg, collector)))

	// Apply middleware chain: CORS -> Logging -> Routes
	handler := middleware.CORSMiddleware(middleware.LoggingMiddleware(mux))

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.MaxTimeout+10) * time.Second,
		WriteTimeout: time.Duration(cfg.MaxTimeout+10) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func verifyBinaries() error {
	// Check a few key wrapper scripts exist
	wrapperScripts := []string{
		"/usr/local/bin/curl_chrome116",
		"/usr/local/bin/curl_ff109",
	}

	for _, script := range wrapperScripts {
		if _, err := os.Stat(script); os.IsNotExist(err) {
			return fmt.Errorf("wrapper script not found: %s", script)
		}
	}

	return nil
}
