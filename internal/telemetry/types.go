package telemetry

import "time"

// ClientEvent represents a client-side telemetry event (shared type)
type ClientEvent struct {
	Type          string                 `json:"type"`
	Timestamp     time.Time              `json:"timestamp"`
	SessionID     string                 `json:"session_id"`
	CorrelationID string                 `json:"correlation_id"`
	Level         int                    `json:"level,omitempty"`
	Score         int                    `json:"score,omitempty"`
	Data          string                 `json:"data,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
	SpanID        string                 `json:"span_id,omitempty"`
}

// ClientMetric represents a client-side metric (shared type)
type ClientMetric struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Type      string                 `json:"type"` // counter, gauge, histogram
	Timestamp time.Time              `json:"timestamp"`
	SessionID string                 `json:"session_id"`
	Labels    map[string]interface{} `json:"labels,omitempty"`
}