package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type helloResponse struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/hello", helloHandler)
	mux.HandleFunc("GET /healthz", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "18081"
	}

	addr := ":" + port
	log.Printf("API server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, logRequests(mux)); err != nil {
		log.Fatal(err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, helloResponse{
		Message: "Hello from Go!",
		Time:    time.Now().UTC(),
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
