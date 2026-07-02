package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloMessage(t *testing.T) {
	t.Run("uses configured secret", func(t *testing.T) {
		t.Setenv("MY_SECRET", "test-secret")

		got := helloMessage()
		want := "Hello from Go! MY_SECRET!: test-secret"
		if got != want {
			t.Fatalf("helloMessage() = %q, want %q", got, want)
		}
	})

	t.Run("uses fallback when secret is unset", func(t *testing.T) {
		t.Setenv("MY_SECRET", "")

		got := helloMessage()
		want := "Hello from Go! MY_SECRET!: (MY_SECRET is not set)"
		if got != want {
			t.Fatalf("helloMessage() = %q, want %q", got, want)
		}
	})
}

func TestHelloHandlerWithoutDatabase(t *testing.T) {
	t.Setenv("MY_SECRET", "handler-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()

	helloHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	var response helloResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Message != "Hello from Go! MY_SECRET!: handler-secret" {
		t.Fatalf("message = %q", response.Message)
	}
	if response.DBEnabled {
		t.Fatal("DBEnabled = true, want false")
	}
	if response.HitCount != nil {
		t.Fatalf("HitCount = %v, want nil", *response.HitCount)
	}
	if response.LastHitAt != nil {
		t.Fatalf("LastHitAt = %v, want nil", *response.LastHitAt)
	}
	if response.Time.IsZero() {
		t.Fatal("Time was not set")
	}
}
