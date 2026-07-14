package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

type sendEmailRequest struct {
	Recipient string `json:"recipient"`
}

type sendEmailResponse struct {
	Message string `json:"message"`
	From    string `json:"from"`
	To      string `json:"to"`
	Domain  string `json:"domain"`
}

type emailServiceRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

func main() {
	log.Println("hi3")
	db, err := openDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if db != nil {
		defer db.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/hello", helloHandler(db))
	mux.HandleFunc("POST /api/email", sendEmailHandler(http.DefaultClient))
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

func sendEmailHandler(client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request sendEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON request."})
			return
		}

		recipient := strings.TrimSpace(request.Recipient)
		if recipient == "" || !strings.Contains(recipient, "@") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Enter a valid recipient email address."})
			return
		}

		platformURL := strings.TrimRight(os.Getenv("PLATFORM_SERVICES_URL"), "/")
		platformKey := os.Getenv("PLATFORM_SERVICES_KEY")
		if platformURL == "" || platformKey == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Platform Services is not configured."})
			return
		}

		namespace, err := currentNamespace()
		if err != nil {
			log.Printf("load namespace: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Namespace metadata is not available."})
			return
		}

		domain := namespace + ".infrapad.ai"
		from := "noreply@" + domain
		serviceRequest := emailServiceRequest{
			From:    from,
			To:      []string{recipient},
			Subject: "Hello",
			Text:    "Hello from your namespace.",
		}

		body, err := json.Marshal(serviceRequest)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to build email request."})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		serviceURL := platformURL + "/infrapad.v1.EmailService/SendNamespaceEmail"
		upstreamRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, serviceURL, bytes.NewReader(body))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to build Platform Services request."})
			return
		}
		upstreamRequest.Header.Set("Authorization", "Bearer "+platformKey)
		upstreamRequest.Header.Set("Connect-Protocol-Version", "1")
		upstreamRequest.Header.Set("Content-Type", "application/json")

		upstreamResponse, err := client.Do(upstreamRequest)
		if err != nil {
			log.Printf("send namespace email: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Platform Services email request failed."})
			return
		}
		defer upstreamResponse.Body.Close()

		if upstreamResponse.StatusCode < 200 || upstreamResponse.StatusCode >= 300 {
			detail, _ := io.ReadAll(io.LimitReader(upstreamResponse.Body, 2048))
			log.Printf("send namespace email: platform services returned %d: %s", upstreamResponse.StatusCode, strings.TrimSpace(string(detail)))
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Platform Services rejected the email request."})
			return
		}

		writeJSON(w, http.StatusOK, sendEmailResponse{
			Message: "Email sent.",
			From:    from,
			To:      recipient,
			Domain:  domain,
		})
	}
}

func currentNamespace() (string, error) {
	if namespace := strings.TrimSpace(os.Getenv("POD_NAMESPACE")); namespace != "" {
		return namespace, nil
	}

	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}

	namespace := strings.TrimSpace(string(data))
	if namespace == "" {
		return "", fmt.Errorf("service account namespace file is empty")
	}

	return namespace, nil
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
