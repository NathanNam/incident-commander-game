//go:build js && wasm
// +build js,wasm

package telemetry

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"
)

// ClientTelemetry handles client-side telemetry collection and transmission
type ClientTelemetry struct {
	sessionID     string
	correlationID string
	serverURL     string
	events        []ClientEvent
	metrics       map[string]interface{}
}

// NewClientTelemetry creates a new client telemetry instance
func NewClientTelemetry(serverURL string) *ClientTelemetry {
	sessionID := generateSessionID()
	correlationID := generateCorrelationID()

	return &ClientTelemetry{
		sessionID:     sessionID,
		correlationID: correlationID,
		serverURL:     serverURL,
		events:        make([]ClientEvent, 0),
		metrics:       make(map[string]interface{}),
	}
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	// Use JavaScript crypto API for better randomness
	crypto := js.Global().Get("crypto")
	if !crypto.IsUndefined() {
		array := js.Global().Get("Uint8Array").New(16)
		crypto.Call("getRandomValues", array)

		// Convert to hex string
		result := ""
		for i := 0; i < 16; i++ {
			b := array.Index(i).Int()
			result += fmt.Sprintf("%02x", b)
		}
		return "session_" + result
	}

	// Fallback to timestamp-based ID
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// generateCorrelationID creates a correlation ID for request tracing
func generateCorrelationID() string {
	return fmt.Sprintf("corr_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// LogEvent records a client-side event
func (ct *ClientTelemetry) LogEvent(eventType string, level, score int, data string, attributes map[string]interface{}) {
	event := ClientEvent{
		Type:          eventType,
		Timestamp:     time.Now(),
		SessionID:     ct.sessionID,
		CorrelationID: ct.correlationID,
		Level:         level,
		Score:         score,
		Data:          data,
		Attributes:    attributes,
	}

	ct.events = append(ct.events, event)

	// Log to browser console for debugging
	js.Global().Get("console").Call("log",
		fmt.Sprintf("[CLIENT_TELEMETRY] %s - Session: %s, Level: %d, Score: %d",
			eventType, ct.sessionID, level, score))

	// Send event immediately for real-time telemetry
	ct.sendEvent(event)
}

// RecordMetric records a client-side metric
func (ct *ClientTelemetry) RecordMetric(name string, value float64, metricType string, labels map[string]interface{}) {
	metric := ClientMetric{
		Name:      name,
		Value:     value,
		Type:      metricType,
		Timestamp: time.Now(),
		SessionID: ct.sessionID,
		Labels:    labels,
	}

	// Store locally
	ct.metrics[name] = value

	// Send metric
	ct.sendMetric(metric)
}

// StartSpan creates a new trace span (simplified implementation)
func (ct *ClientTelemetry) StartSpan(operationName string) *ClientSpan {
	return &ClientSpan{
		TraceID:       generateTraceID(),
		SpanID:        generateSpanID(),
		OperationName: operationName,
		StartTime:     time.Now(),
		SessionID:     ct.sessionID,
		telemetry:     ct,
	}
}

// sendEvent sends an event to the server telemetry endpoint
func (ct *ClientTelemetry) sendEvent(event ClientEvent) {
	go ct.sendToServer("/api/telemetry/events", event)
}

// sendMetric sends a metric to the server telemetry endpoint
func (ct *ClientTelemetry) sendMetric(metric ClientMetric) {
	go ct.sendToServer("/api/telemetry/metrics", metric)
}

// sendToServer sends telemetry data to server endpoint
func (ct *ClientTelemetry) sendToServer(endpoint string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		js.Global().Get("console").Call("error", "Failed to marshal telemetry data:", err.Error())
		return
	}

	// Use fetch API to send data
	url := ct.serverURL + endpoint
	headers := map[string]interface{}{
		"Content-Type":      "application/json",
		"X-Session-ID":      ct.sessionID,
		"X-Correlation-ID":  ct.correlationID,
	}

	options := map[string]interface{}{
		"method": "POST",
		"headers": headers,
		"body": string(jsonData),
	}

	// Convert Go map to JavaScript object
	jsOptions := js.ValueOf(options)

	fetch := js.Global().Get("fetch")
	promise := fetch.Invoke(url, jsOptions)

	// Handle promise (fire and forget for now)
	promise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		js.Global().Get("console").Call("error", "Failed to send telemetry:", args[0])
		return nil
	}))
}

// GetSessionID returns the current session ID
func (ct *ClientTelemetry) GetSessionID() string {
	return ct.sessionID
}

// GetCorrelationID returns the current correlation ID
func (ct *ClientTelemetry) GetCorrelationID() string {
	return ct.correlationID
}

// SetCorrelationID updates the correlation ID (for request correlation)
func (ct *ClientTelemetry) SetCorrelationID(correlationID string) {
	ct.correlationID = correlationID
}

// ClientSpan represents a trace span on the client side
type ClientSpan struct {
	TraceID       string
	SpanID        string
	ParentSpanID  string
	OperationName string
	StartTime     time.Time
	EndTime       time.Time
	SessionID     string
	Attributes    map[string]interface{}
	telemetry     *ClientTelemetry
}

// SetAttribute adds an attribute to the span
func (cs *ClientSpan) SetAttribute(key string, value interface{}) {
	if cs.Attributes == nil {
		cs.Attributes = make(map[string]interface{})
	}
	cs.Attributes[key] = value
}

// End finishes the span and sends it
func (cs *ClientSpan) End() {
	cs.EndTime = time.Now()
	duration := cs.EndTime.Sub(cs.StartTime)

	// Create span event
	event := ClientEvent{
		Type:          "span",
		Timestamp:     cs.StartTime,
		SessionID:     cs.SessionID,
		CorrelationID: cs.telemetry.correlationID,
		Data:          cs.OperationName,
		TraceID:       cs.TraceID,
		SpanID:        cs.SpanID,
		Attributes: map[string]interface{}{
			"duration_ms":    duration.Milliseconds(),
			"operation_name": cs.OperationName,
		},
	}

	// Merge span attributes
	for k, v := range cs.Attributes {
		event.Attributes[k] = v
	}

	cs.telemetry.sendEvent(event)
}

// Helper functions for generating IDs
func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func generateSpanID() string {
	return fmt.Sprintf("span_%d", time.Now().UnixNano())
}