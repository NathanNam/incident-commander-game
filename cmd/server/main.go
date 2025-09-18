package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	requestCounter     metric.Int64Counter
	requestDuration    metric.Float64Histogram
	healthCheckCount   metric.Int64Counter
	clientEventCounter metric.Int64Counter
	gameMetricsGauge   metric.Float64Gauge
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

// clientTelemetryEventsHandler handles client-side telemetry events
func clientTelemetryEventsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.GetLogger()
	tracer := telemetry.GetTracer()

	// Start span for processing client telemetry
	ctx, span := tracer.Start(ctx, "process_client_telemetry_event")
	defer span.End()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read request body")
		logger.ErrorContext(ctx, "Failed to read telemetry event body", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse client event
	var clientEvent telemetry.ClientEvent
	if err := json.Unmarshal(body, &clientEvent); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to unmarshal client event")
		logger.ErrorContext(ctx, "Failed to unmarshal client event", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract correlation headers
	sessionID := r.Header.Get("X-Session-ID")
	correlationID := r.Header.Get("X-Correlation-ID")

	// Add span attributes
	span.SetAttributes(
		attribute.String("client.event.type", clientEvent.Type),
		attribute.String("client.session_id", sessionID),
		attribute.String("client.correlation_id", correlationID),
		attribute.Int("client.event.level", clientEvent.Level),
		attribute.Int("client.event.score", clientEvent.Score),
	)

	// Log the client event with correlation info
	logger.InfoContext(ctx, "Client telemetry event received",
		"event_type", clientEvent.Type,
		"session_id", sessionID,
		"correlation_id", correlationID,
		"level", clientEvent.Level,
		"score", clientEvent.Score,
		"data", clientEvent.Data,
		"client_timestamp", clientEvent.Timestamp,
	)

	// Increment client event counter
	clientEventCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("event_type", clientEvent.Type),
		attribute.String("session_id", sessionID),
	))

	// Record business metrics based on event type
	switch clientEvent.Type {
	case "level_change":
		gameMetricsGauge.Record(ctx, float64(clientEvent.Level), metric.WithAttributes(
			attribute.String("metric_type", "current_level"),
			attribute.String("session_id", sessionID),
		))
	case "score_change":
		gameMetricsGauge.Record(ctx, float64(clientEvent.Score), metric.WithAttributes(
			attribute.String("metric_type", "current_score"),
			attribute.String("session_id", sessionID),
		))
	case "game_start":
		gameMetricsGauge.Record(ctx, 1, metric.WithAttributes(
			attribute.String("metric_type", "game_sessions"),
			attribute.String("session_id", sessionID),
		))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

// clientTelemetryMetricsHandler handles client-side metrics
func clientTelemetryMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.GetLogger()
	tracer := telemetry.GetTracer()

	// Start span for processing client metrics
	ctx, span := tracer.Start(ctx, "process_client_telemetry_metric")
	defer span.End()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read request body")
		logger.ErrorContext(ctx, "Failed to read telemetry metric body", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse client metric
	var clientMetric telemetry.ClientMetric
	if err := json.Unmarshal(body, &clientMetric); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to unmarshal client metric")
		logger.ErrorContext(ctx, "Failed to unmarshal client metric", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Extract correlation headers
	sessionID := r.Header.Get("X-Session-ID")
	correlationID := r.Header.Get("X-Correlation-ID")

	// Add span attributes
	span.SetAttributes(
		attribute.String("client.metric.name", clientMetric.Name),
		attribute.String("client.metric.type", clientMetric.Type),
		attribute.Float64("client.metric.value", clientMetric.Value),
		attribute.String("client.session_id", sessionID),
		attribute.String("client.correlation_id", correlationID),
	)

	// Log the client metric
	logger.InfoContext(ctx, "Client telemetry metric received",
		"metric_name", clientMetric.Name,
		"metric_type", clientMetric.Type,
		"metric_value", clientMetric.Value,
		"session_id", sessionID,
		"correlation_id", correlationID,
		"client_timestamp", clientMetric.Timestamp,
	)

	// Record the metric in our telemetry system
	gameMetricsGauge.Record(ctx, clientMetric.Value, metric.WithAttributes(
		attribute.String("metric_type", clientMetric.Name),
		attribute.String("session_id", sessionID),
		attribute.String("source", "client"),
	))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
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

	clientEventCounter, err = meter.Int64Counter("client_events_total",
		metric.WithDescription("Total number of client telemetry events received"))
	if err != nil {
		log.Fatal("Failed to create client event counter:", err)
	}

	gameMetricsGauge, err = meter.Float64Gauge("game_metrics",
		metric.WithDescription("Various game-related metrics"))
	if err != nil {
		log.Fatal("Failed to create game metrics gauge:", err)
	}

	logger := telemetry.GetLogger()
	logger.Info("OpenTelemetry metrics initialized")

	// Set up instrumented routes
	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(serveIndex), "GET /"))
	http.Handle("/health", otelhttp.NewHandler(http.HandlerFunc(healthCheckHandler), "GET /health"))

	// Client telemetry API endpoints
	http.Handle("/api/telemetry/events", otelhttp.NewHandler(corsMiddleware(http.HandlerFunc(clientTelemetryEventsHandler)), "POST /api/telemetry/events"))
	http.Handle("/api/telemetry/metrics", otelhttp.NewHandler(corsMiddleware(http.HandlerFunc(clientTelemetryMetricsHandler)), "POST /api/telemetry/metrics"))

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
