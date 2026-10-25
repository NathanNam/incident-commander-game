// Command server serves web/ for local play: go run ./cmd/server (or make serve).
// The live site is static files on Vercel; this is only a convenience.
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to listen on")
	flag.Parse()

	files := http.FileServer(http.Dir("web"))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		files.ServeHTTP(w, r)
	}))

	log.Printf("Incident Commander: http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
