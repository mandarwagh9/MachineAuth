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

	"agentauth/internal/config"
	"agentauth/internal/db"
	"agentauth/internal/handlers"
	"agentauth/internal/middleware"
	"agentauth/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	tokenService, err := services.NewTokenService(cfg)
	if err != nil {
		log.Fatalf("failed to create token service: %v", err)
	}

	agentService := services.NewAgentService(database)
	auditService := services.NewAuditService(database)
	webhookService := services.NewWebhookService(database)

	// Wire up webhook triggering in audit service
	auditService.SetWebhookService(webhookService)

	// Start webhook delivery worker
	webhookWorker := services.NewDeliveryWorker(webhookService, cfg.WebhookWorkerCount)
	webhookWorker.Start()
	defer webhookWorker.Stop()

	authHandler := handlers.NewAuthHandler(agentService, tokenService)
	agentsHandler := handlers.NewAgentsHandler(agentService, auditService)
	webhookHandler := handlers.NewWebhookHandler(webhookService, auditService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/oauth/token", authHandler.Token)

	mux.HandleFunc("/.well-known/jwks.json", tokenService.JWKS)

	mux.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			agentsHandler.List(w, r)
		case http.MethodPost:
			agentsHandler.Create(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/agents/", agentsHandler.HandleAgent)

	// Webhook routes
	mux.HandleFunc("/api/webhooks", webhookHandler.ListAndCreate)
	mux.HandleFunc("/api/webhooks/", webhookHandler.HandleWebhook)
	mux.HandleFunc("/api/webhook-events", webhookHandler.HandleEvents)

	loggedMux := middleware.Logging(mux)
	corsMux := middleware.CORS(cfg.AllowedOrigins, loggedMux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      corsMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
