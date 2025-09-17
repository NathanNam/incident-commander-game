package telemetry

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Global telemetry instances
var (
	appTracer trace.Tracer
	appMeter  metric.Meter
	appLogger *slog.Logger
)

// parseOTLPEndpoint parses the OTLP endpoint URL and returns the host, base path, and whether to use HTTPS
func parseOTLPEndpoint(endpoint string) (host string, basePath string, useHTTPS bool, err error) {
	// If no protocol is specified, assume localhost:port format
	if !strings.Contains(endpoint, "://") {
		return endpoint, "", false, nil
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", false, err
	}

	host = u.Host
	if u.Port() == "" {
		if u.Scheme == "https" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	// Preserve the path (e.g., "/v2/otel" for Observe)
	basePath = strings.TrimSuffix(u.Path, "/")
	useHTTPS = u.Scheme == "https"
	return host, basePath, useHTTPS, nil
}

// buildOTLPHeaders creates the standard headers for OTLP exporters.
func buildOTLPHeaders(targetPackage, bearerToken string) map[string]string {
	headers := map[string]string{
		"x-observe-target-package": targetPackage,
	}
	if bearerToken != "" {
		headers["Authorization"] = "Bearer " + bearerToken
	}
	return headers
}

// setupTracing configures OpenTelemetry tracing with OTLP HTTP exporter.
func setupTracing(ctx context.Context, res *resource.Resource, otlpEndpoint, bearerToken string) (*sdktrace.TracerProvider, error) {
	host, basePath, useHTTPS, err := parseOTLPEndpoint(otlpEndpoint)
	if err != nil {
		return nil, err
	}

	headers := buildOTLPHeaders("Tracing", bearerToken)
	urlPath := basePath + "/v1/traces"
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(host),
		otlptracehttp.WithURLPath(urlPath),
		otlptracehttp.WithHeaders(headers),
	}

	if !useHTTPS {
		options = append(options, otlptracehttp.WithInsecure())
	}

	traceExporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp, nil
}

// setupMetrics configures OpenTelemetry metrics with OTLP HTTP exporter.
func setupMetrics(ctx context.Context, res *resource.Resource, otlpEndpoint, bearerToken string) (*sdkmetric.MeterProvider, error) {
	host, basePath, useHTTPS, err := parseOTLPEndpoint(otlpEndpoint)
	if err != nil {
		return nil, err
	}

	headers := buildOTLPHeaders("Metrics", bearerToken)
	urlPath := basePath + "/v1/metrics"
	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(host),
		otlpmetrichttp.WithURLPath(urlPath),
		otlpmetrichttp.WithHeaders(headers),
	}

	if !useHTTPS {
		options = append(options, otlpmetrichttp.WithInsecure())
	}

	metricExporter, err := otlpmetrichttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return mp, nil
}

// setupLogging configures OpenTelemetry logging with OTLP HTTP exporter and structured logging.
func setupLogging(ctx context.Context, res *resource.Resource, otlpEndpoint, bearerToken, serviceName string) (*sdklog.LoggerProvider, error) {
	host, basePath, useHTTPS, err := parseOTLPEndpoint(otlpEndpoint)
	if err != nil {
		return nil, err
	}

	headers := buildOTLPHeaders("Logs", bearerToken)
	urlPath := basePath + "/v1/logs"
	options := []otlploghttp.Option{
		otlploghttp.WithEndpoint(host),
		otlploghttp.WithURLPath(urlPath),
		otlploghttp.WithHeaders(headers),
	}

	if !useHTTPS {
		options = append(options, otlploghttp.WithInsecure())
	}

	logExporter, err := otlploghttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	// Create structured logger that will send logs to OTLP
	otelHandler := otelslog.NewHandler(serviceName)
	appLogger = slog.New(otelHandler)

	return lp, nil
}

// SetupInstrumentation initializes OpenTelemetry with tracing, metrics, and logging.
// Returns a cleanup function that should be called before application shutdown.
func SetupInstrumentation(serviceName string) func() {
	ctx := context.Background()

	// Get OTLP endpoint from environment or use default
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4318"
	}

	// Get bearer token from environment
	bearerToken := os.Getenv("OTEL_EXPORTER_OTLP_BEARER_TOKEN")

	// Create resource with service identification
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		slog.Error("failed to create resource", "error", err)
		panic(err)
	}

	// Setup tracing
	tp, err := setupTracing(ctx, res, otlpEndpoint, bearerToken)
	if err != nil {
		slog.Error("failed to setup tracing", "error", err)
		panic(err)
	}
	appTracer = otel.Tracer(serviceName)

	// Setup metrics
	mp, err := setupMetrics(ctx, res, otlpEndpoint, bearerToken)
	if err != nil {
		slog.Error("failed to setup metrics", "error", err)
		panic(err)
	}
	appMeter = otel.Meter(serviceName)

	// Setup logging
	lp, err := setupLogging(ctx, res, otlpEndpoint, bearerToken, serviceName)
	if err != nil {
		slog.Error("failed to setup logging", "error", err)
		panic(err)
	}

	appLogger.Info("OpenTelemetry instrumentation initialized",
		"service", serviceName,
		"endpoint", otlpEndpoint)

	// Return cleanup function
	return func() {
		appLogger.Info("Shutting down OpenTelemetry instrumentation")

		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown tracer provider", "error", err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown meter provider", "error", err)
		}
		if err := lp.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown logger provider", "error", err)
		}
	}
}

// GetTracer returns the global tracer instance.
// Call SetupInstrumentation first.
func GetTracer() trace.Tracer {
	return appTracer
}

// GetMeter returns the global meter instance.
// Call SetupInstrumentation first.
func GetMeter() metric.Meter {
	return appMeter
}

// GetLogger returns the global structured logger instance.
// Call SetupInstrumentation first.
func GetLogger() *slog.Logger {
	return appLogger
}
