package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	health := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "incident-commander-game",
	}
	
	json.NewEncoder(w).Encode(health)
}

// corsMiddleware adds CORS headers for WebAssembly
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// serveIndex serves the main HTML page
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func main() {
	// Set up routes
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/health", healthCheckHandler)
	
	// Serve static files with CORS headers
	fileServer := http.FileServer(http.Dir("web/"))
	http.Handle("/web/", corsMiddleware(http.StripPrefix("/web/", fileServer)))
	http.Handle("/static/", corsMiddleware(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/")))))
	http.Handle("/images/", corsMiddleware(http.StripPrefix("/images/", http.FileServer(http.Dir("web/images/")))))

	fmt.Println("🎮 Incident Commander Game Server starting on :8080")
	fmt.Println("🌐 Open http://localhost:8080 to play!")
	fmt.Println("🔍 Health check available at http://localhost:8080/health")
	fmt.Println("🎯 Each browser session gets its own game instance")
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}