package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0510t/intelligence-terminal/apps/api/internal/auth"
	"github.com/0510t/intelligence-terminal/apps/api/internal/config"
	handler "github.com/0510t/intelligence-terminal/apps/api/internal/http"
	"github.com/0510t/intelligence-terminal/apps/api/internal/jobs"
	"github.com/0510t/intelligence-terminal/apps/api/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	// Repositories
	macroRepo := storage.NewMacroRepo(pool)
	cotRepo := storage.NewCOTRepo(pool)
	snapshotRepo := storage.NewSnapshotRepo(pool)
	signalRepo := storage.NewSignalRepo(pool)
	configRepo := storage.NewConfigRepo(pool)
	sourceRunRepo := storage.NewSourceRunRepo(pool)

	// Jobs
	fredClient := jobs.NewFREDClient(cfg.FREDAPIKey)
	fredIngestJob := jobs.NewFREDIngestJob(fredClient, macroRepo, sourceRunRepo)

	cftcClient := jobs.NewCFTCClient()
	cftcIngestJob := jobs.NewCFTCIngestJob(cftcClient, cotRepo, sourceRunRepo)

	snapshotJob := jobs.NewSnapshotJob(macroRepo, snapshotRepo, signalRepo, cotRepo, configRepo, sourceRunRepo)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints
		r.Get("/health", healthHandler(pool))

		// Internal endpoints (auth required)
		r.Route("/internal", func(r chi.Router) {
			r.Use(auth.InternalAuth(cfg.InternalAuthToken))

			r.Post("/jobs/ingest-macro", ingestMacroHandler(fredIngestJob))
			r.Post("/jobs/ingest-cot", ingestCOTHandler(cftcIngestJob))
			r.Post("/jobs/rebuild-snapshots", rebuildSnapshotsHandler(snapshotJob))
		})
	})

	// Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second, // longer for ingestion endpoints
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down server")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	slog.Info("server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := pool.Ping(r.Context())
		if err != nil {
			handler.WriteError(w, r, http.StatusServiceUnavailable, "INTERNAL_ERROR", "database unreachable")
			return
		}
		handler.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
			"db":     "connected",
		})
	}
}

type ingestMacroRequest struct {
	LookbackYears int `json:"lookback_years"`
}

type jobResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func ingestMacroHandler(job *jobs.FREDIngestJob) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ingestMacroRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req.LookbackYears = 5 // default
		}
		if req.LookbackYears <= 0 || req.LookbackYears > 20 {
			req.LookbackYears = 5
		}

		// Run ingestion synchronously for now; will move to async worker later
		if err := job.Run(r.Context(), req.LookbackYears); err != nil {
			slog.Error("macro ingestion failed", "error", err)
			handler.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "macro ingestion failed: "+err.Error())
			return
		}

		handler.WriteJSON(w, http.StatusOK, jobResponse{
			Status:  "success",
			Message: "macro ingestion completed",
		})
	}
}

func ingestCOTHandler(job *jobs.CFTCIngestJob) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ingestMacroRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			req.LookbackYears = 5
		}
		if req.LookbackYears <= 0 || req.LookbackYears > 20 {
			req.LookbackYears = 5
		}

		if err := job.Run(r.Context(), req.LookbackYears); err != nil {
			slog.Error("COT ingestion failed", "error", err)
			handler.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "COT ingestion failed: "+err.Error())
			return
		}

		handler.WriteJSON(w, http.StatusOK, jobResponse{
			Status:  "success",
			Message: "COT ingestion completed",
		})
	}
}

func rebuildSnapshotsHandler(job *jobs.SnapshotJob) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := job.Run(r.Context()); err != nil {
			slog.Error("snapshot rebuild failed", "error", err)
			handler.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "snapshot rebuild failed: "+err.Error())
			return
		}

		handler.WriteJSON(w, http.StatusOK, jobResponse{
			Status:  "success",
			Message: "snapshot rebuild completed",
		})
	}
}
