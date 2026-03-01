package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"machineauth/internal/config"
	"machineauth/internal/db"
	"machineauth/internal/handlers"
	"machineauth/internal/middleware"
	"machineauth/internal/services"
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

	tokenService, err := services.NewTokenService(cfg, database)
	if err != nil {
		log.Fatalf("failed to create token service: %v", err)
	}

	agentService := services.NewAgentService(database)
	auditService := services.NewAuditService(database)
	orgService := services.NewOrganizationService(database)
	teamService := services.NewTeamService(database)

	authHandler := handlers.NewAuthHandler(agentService, tokenService)
	agentsHandler := handlers.NewAgentsHandler(agentService, auditService)
	agentSelfHandler := handlers.NewAgentSelfHandler(agentService)
	orgHandler := handlers.NewOrganizationHandler(orgService, teamService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"endpoints":"/oauth/token, /oauth/introspect, /oauth/revoke, /oauth/refresh, /api/agents, /.well-known/jwks.json","service":"MachineAuth","status":"running","version":"1.0.0"}`))
	})

	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		agentCount, _ := agentService.Count()
		w.Write([]byte(fmt.Sprintf(`{"status":"ok","timestamp":"%s","agents_count":%d}`, time.Now().UTC().Format(time.RFC3339), agentCount)))
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		stats, _ := agentService.GetStats()
		tokensRefreshed, tokensRevoked := tokenService.GetMetrics()
		w.Write([]byte(fmt.Sprintf(`{"requests":%d,"tokens_issued":%d,"tokens_refreshed":%d,"tokens_revoked":%d,"active_tokens":%d,"total_agents":%d}`, stats.TotalRequests, stats.TokensIssued, tokensRefreshed, tokensRevoked, stats.ActiveTokens, stats.TotalAgents)))
	})

	mux.HandleFunc("/oauth/token", authHandler.Token)
	mux.HandleFunc("/oauth/introspect", authHandler.Introspect)
	mux.HandleFunc("/oauth/revoke", authHandler.Revoke)
	mux.HandleFunc("/oauth/refresh", authHandler.Refresh)

	mux.HandleFunc("/.well-known/jwks.json", tokenService.JWKS)

	mux.HandleFunc("/api/organizations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			orgHandler.ListOrganizations(w, r)
		case http.MethodPost:
			orgHandler.CreateOrganization(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/organizations/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.Contains(path, "/teams") {
			switch r.Method {
			case http.MethodGet:
				orgHandler.ListTeams(w, r)
			case http.MethodPost:
				orgHandler.CreateTeam(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			switch r.Method {
			case http.MethodGet:
				orgHandler.GetOrganization(w, r)
			case http.MethodPut, http.MethodPatch:
				orgHandler.UpdateOrganization(w, r)
			case http.MethodDelete:
				orgHandler.DeleteOrganization(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})

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

	jwtAuth := middleware.JWTAuth(tokenService)

	mux.Handle("/api/agents/me", jwtAuth(http.HandlerFunc(agentSelfHandler.GetMe)))
	mux.Handle("/api/agents/me/usage", jwtAuth(http.HandlerFunc(agentSelfHandler.GetUsage)))
	mux.Handle("/api/agents/me/rotate", jwtAuth(http.HandlerFunc(agentSelfHandler.RotateCredentials)))
	mux.Handle("/api/agents/me/deactivate", jwtAuth(http.HandlerFunc(agentSelfHandler.Deactivate)))
	mux.Handle("/api/agents/me/reactivate", jwtAuth(http.HandlerFunc(agentSelfHandler.Reactivate)))
	mux.Handle("/api/agents/me/delete", jwtAuth(http.HandlerFunc(agentSelfHandler.Delete)))

	mux.Handle("/api/verify", jwtAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentID, ok := middleware.GetAgentIDFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		agent, err := agentService.GetByID(agentID)
		if err != nil {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":       true,
			"agent_id":    agent.ID,
			"client_id":   agent.ClientID,
			"name":        agent.Name,
			"scopes":      agent.Scopes,
			"is_active":   agent.IsActive,
			"token_count": agent.TokenCount,
		})
	})))

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
