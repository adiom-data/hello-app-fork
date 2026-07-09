package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type helloResponse struct {
	Message   string     `json:"message"`
	Time      time.Time  `json:"time"`
	DBEnabled bool       `json:"dbEnabled"`
	HitCount  *int64     `json:"hitCount,omitempty"`
	LastHitAt *time.Time `json:"lastHitAt,omitempty"`
	DBError   string     `json:"dbError,omitempty"`
}

func main() {
	log.Println("hi1")
	db, err := openDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/hello", helloHandler(db))
	mux.HandleFunc("GET /healthz", healthHandler(db))

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

func helloHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := helloResponse{
			Message:   helloMessage(),
			Time:      time.Now().UTC(),
			DBEnabled: db != nil,
		}

		if db != nil {
			hitCount, lastHitAt, err := recordHelloHit(r.Context(), db)
			if err != nil {
				log.Printf("record hello hit: %v", err)
				response.DBError = err.Error()
			} else {
				response.HitCount = &hitCount
				response.LastHitAt = &lastHitAt
			}
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func helloMessage() string {
	secret := os.Getenv("MY_SECRET")
	if secret == "" {
		secret = "(MY_SECRET is not set!)"
	}
	return fmt.Sprintf("Hello from Go! MY_SECRET!: %s", secret)
}

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{
					"status": "degraded",
					"error":  err.Error(),
				})
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func openDatabase() (*sql.DB, error) {
	if os.Getenv("PGHOST") == "" {
		return nil, nil
	}

	db, err := sql.Open("pgx", postgresDSN())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func recordHelloHit(ctx context.Context, db *sql.DB) (int64, time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, `
		create table if not exists hello_hits (
			id bigserial primary key,
			hit_at timestamptz not null default now()
		)
	`); err != nil {
		return 0, time.Time{}, err
	}

	var hitCount int64
	var lastHitAt time.Time
	err := db.QueryRowContext(ctx, `
		with inserted as (
			insert into hello_hits default values
			returning hit_at
		)
		select count(*), max(hit_at)
		from hello_hits
	`).Scan(&hitCount, &lastHitAt)
	if err != nil {
		return 0, time.Time{}, err
	}

	return hitCount, lastHitAt.UTC(), nil
}

func postgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		envOrDefault("PGHOST", "localhost"),
		envOrDefault("PGPORT", "5432"),
		envOrDefault("PGDATABASE", "postgres"),
		envOrDefault("PGUSER", "postgres"),
		os.Getenv("PGPASSWORD"),
		envOrDefault("PGSSLMODE", "disable"),
	)
}

func envOrDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
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
