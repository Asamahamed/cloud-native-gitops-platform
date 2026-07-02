// Package main implements a small production-style HTTP service that exposes
// Prometheus metrics, Kubernetes health/readiness probes and structured logging.
// It is intentionally lightweight so it can be used to demonstrate the full
// GitOps + observability delivery pipeline end to end.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, by path and status.",
		},
		[]string{"path", "method", "status"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency distribution in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	// ready flips to true once startup work is done; the readiness probe reads it.
	ready atomic.Bool
)

// version is set at build time via -ldflags "-X main.version=...".
// It falls back to the APP_VERSION env var, then "dev".
var version = "dev"

// instrument wraps a handler to record Prometheus metrics for every request.
func instrument(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		httpDuration.WithLabelValues(path).Observe(time.Since(start).Seconds())
		httpRequests.WithLabelValues(path, r.Method, http.StatusText(rec.status)).Inc()
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := getEnv("PORT", "8080")
	if v := os.Getenv("APP_VERSION"); v != "" {
		version = v
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", instrument("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"service": "gitops-demo-app",
			"version": version,
			"message": "Cloud-native GitOps platform demo service is healthy 🚀",
		})
	}))

	// Liveness: is the process up?
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Readiness: is the process ready to receive traffic?
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if !ready.Load() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "starting"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Simulate startup work (warm caches, DB connections, etc.), then mark ready.
	go func() {
		time.Sleep(2 * time.Second)
		ready.Store(true)
		slog.Info("service is ready", "version", version)
	}()

	// Graceful shutdown on SIGTERM (important for zero-downtime rollouts).
	go func() {
		slog.Info("starting server", "port", port, "version", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
