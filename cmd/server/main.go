package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NathanNam/incident-commander-game/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// Global metrics
var (
	requestCounter   metric.Int64Counter
	requestDuration  metric.Float64Histogram
	healthCheckCount metric.Int64Counter
)

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create span for health check
	tracer := telemetry.GetTracer()
	ctx, span := tracer.Start(ctx, "health_check")
	defer span.End()

	// Log health check request
	logger := telemetry.GetLogger()
	logger.InfoContext(ctx, "Health check requested",
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent())

	// Increment health check counter
	healthCheckCount.Add(ctx, 1)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	health := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "incident-commander-game",
	}

	// Add span attributes
	span.SetAttributes(
		attribute.String("health.status", health.Status),
		attribute.String("health.service", health.Service),
	)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to encode health response")
		logger.ErrorContext(ctx, "Failed to encode health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.InfoContext(ctx, "Health check completed successfully")
}

// corsMiddleware adds CORS headers for WebAssembly
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := telemetry.GetLogger()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			logger.InfoContext(ctx, "CORS preflight request",
				"path", r.URL.Path,
				"origin", r.Header.Get("Origin"))
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// serveIndex serves the main HTML page
func serveIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create span for serving index
	tracer := telemetry.GetTracer()
	ctx, span := tracer.Start(ctx, "serve_index")
	defer span.End()

	logger := telemetry.GetLogger()
	logger.InfoContext(ctx, "Serving index page",
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent())

	// Add span attributes
	span.SetAttributes(
		attribute.String("http.route", "/"),
		attribute.String("file.path", "web/index.html"),
	)

	http.ServeFile(w, r, "web/index.html")
	logger.InfoContext(ctx, "Index page served successfully")
}

func main() {
	// Initialize OpenTelemetry
	cleanup := telemetry.SetupInstrumentation("incident-commander-server")
	defer cleanup()

	// Initialize metrics
	meter := telemetry.GetMeter()
	var err error

	requestCounter, err = meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		log.Fatal("Failed to create request counter:", err)
	}

	requestDuration, err = meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"))
	if err != nil {
		log.Fatal("Failed to create request duration histogram:", err)
	}

	healthCheckCount, err = meter.Int64Counter("health_checks_total",
		metric.WithDescription("Total number of health check requests"))
	if err != nil {
		log.Fatal("Failed to create health check counter:", err)
	}

	logger := telemetry.GetLogger()
	logger.Info("OpenTelemetry metrics initialized")

	// Set up instrumented routes
	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(serveIndex), "GET /"))
	http.Handle("/health", otelhttp.NewHandler(http.HandlerFunc(healthCheckHandler), "GET /health"))

	// Serve static files with CORS headers and instrumentation
	fileServer := http.FileServer(http.Dir("web/"))
	http.Handle("/web/", otelhttp.NewHandler(corsMiddleware(http.StripPrefix("/web/", fileServer)), "GET /web/*"))
	http.Handle("/static/", otelhttp.NewHandler(corsMiddleware(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/")))), "GET /static/*"))
	http.Handle("/images/", otelhttp.NewHandler(corsMiddleware(http.StripPrefix("/images/", http.FileServer(http.Dir("web/images/")))), "GET /images/*"))

	logger.Info("🎮 Incident Commander Game Server starting on :8080")
	logger.Info("🌐 Open http://localhost:8080 to play!")
	logger.Info("🔍 Health check available at http://localhost:8080/health")
	logger.Info("🎯 Each browser session gets its own game instance")

	// Also print to stdout for compatibility
	fmt.Println("🎮 Incident Commander Game Server starting on :8080")
	fmt.Println("🌐 Open http://localhost:8080 to play!")
	fmt.Println("🔍 Health check available at http://localhost:8080/health")
	fmt.Println("🎯 Each browser session gets its own game instance")

	logger.Info("Server starting to listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
