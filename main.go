package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/zupolgec/curl-impersonate-service/config"
	"github.com/zupolgec/curl-impersonate-service/handlers"
	"github.com/zupolgec/curl-impersonate-service/metrics"
	"github.com/zupolgec/curl-impersonate-service/middleware"
	"github.com/zupolgec/curl-impersonate-service/models"
	"github.com/zupolgec/curl-impersonate-service/store"
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

	// Open datastore (tokens, settings, usage logs)
	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		log.Fatalf("Failed to create data dir %s: %v", cfg.DataDir, err)
	}
	st, err := store.Open(filepath.Join(cfg.DataDir, "impersonate.db"))
	if err != nil {
		log.Fatalf("Failed to open datastore: %v", err)
	}
	defer func() { _ = st.Close() }()

	// Seed the legacy TOKEN as an API token for backward compatibility.
	if cfg.Token != "" {
		if err := st.SeedToken("legacy-env-token", cfg.Token); err != nil {
			log.Printf("Warning: failed to seed legacy token: %v", err)
		}
	}
	if err := handlers.SeedCORSSetting(st, cfg.CORSAllowedOrigins); err != nil {
		log.Printf("Warning: failed to seed CORS setting: %v", err)
	}

	// Start the usage-log retention janitor.
	retention := time.Duration(cfg.LogRetentionHours) * time.Hour
	stopJanitor := startLogJanitor(st, retention)
	defer stopJanitor()

	// Initialize metrics collector
	collector := metrics.NewCollector()

	// Setup HTTP router
	mux := http.NewServeMux()

	// Public endpoint (no auth)
	mux.HandleFunc("/health", handlers.HealthHandler)

	// Protected API endpoints, authenticated against datastore tokens.
	authMw := middleware.AuthMiddleware(st.ValidateToken)
	mux.Handle("/browsers", authMw(http.HandlerFunc(handlers.BrowsersHandler)))
	mux.Handle("/metrics", authMw(handlers.NewMetricsHandler(collector)))
	mux.Handle("/impersonate", authMw(handlers.NewImpersonateHandler(cfg, collector, st)))

	// API docs at /docs (token-authenticated), toggleable.
	if cfg.APIDocsEnabled {
		mux.Handle("/docs", authMw(handlers.NewDocsHandler(cfg.AdminToken != "")))
		log.Printf("API docs enabled at /docs")
	}

	// Admin UI (enabled only when ADMIN_TOKEN is set), protected by Basic auth.
	if cfg.AdminToken != "" {
		adminMw := middleware.AdminAuthMiddleware(cfg.AdminToken)
		mux.Handle("/admin/", adminMw(handlers.NewAdminHandler(st, collector)))
		log.Printf("Admin UI enabled at /admin/")
	}

	// Apply middleware chain: CORS -> Logging -> Routes. CORS origins are read
	// live from the datastore so the admin UI can update them without a restart.
	corsMw := middleware.CORSMiddleware(handlers.CORSOriginProvider(st, cfg.CORSAllowedOrigins))
	handler := corsMw(middleware.LoggingMiddleware(mux))

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

// startLogJanitor periodically purges usage logs older than the retention
// window. It returns a stop function. A non-positive retention disables purging.
func startLogJanitor(st *store.Store, retention time.Duration) func() {
	if retention <= 0 {
		return func() {}
	}
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		// Purge once at startup, then hourly.
		if n, err := st.PurgeLogsOlderThan(retention); err == nil && n > 0 {
			log.Printf("Purged %d expired usage logs", n)
		}
		for {
			select {
			case <-ticker.C:
				if n, err := st.PurgeLogsOlderThan(retention); err == nil && n > 0 {
					log.Printf("Purged %d expired usage logs", n)
				}
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop) }
}

func verifyBinaries() error {
	// Check a few key wrapper scripts exist
	wrapperScripts := []string{
		"/usr/local/bin/curl_chrome136",
		"/usr/local/bin/curl_firefox135",
	}

	for _, script := range wrapperScripts {
		if _, err := os.Stat(script); os.IsNotExist(err) {
			return fmt.Errorf("wrapper script not found: %s", script)
		}
	}

	return nil
}
