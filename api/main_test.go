package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendEmailHandler(t *testing.T) {
	t.Run("sends namespace email through platform services", func(t *testing.T) {
		t.Setenv("PLATFORM_SERVICES_KEY", "test-key")
		t.Setenv("POD_NAMESPACE", "matched")

		var gotRequest emailServiceRequest
		var gotAuth string
		var gotConnectVersion string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/infrapad.v1.EmailService/SendNamespaceEmail" {
				t.Fatalf("path = %q, want email service path", r.URL.Path)
			}
			gotAuth = r.Header.Get("Authorization")
			gotConnectVersion = r.Header.Get("Connect-Protocol-Version")
			if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
				t.Fatalf("decode upstream request: %v", err)
			}
			writeJSON(w, http.StatusOK, map[string]string{})
		}))
		defer server.Close()
		t.Setenv("PLATFORM_SERVICES_URL", server.URL)

		req := httptest.NewRequest(http.MethodPost, "/api/email", strings.NewReader(`{"recipient":"user@example.com"}`))
		rec := httptest.NewRecorder()

		sendEmailHandler(server.Client()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if gotAuth != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want bearer token", gotAuth)
		}
		if gotConnectVersion != "1" {
			t.Fatalf("Connect-Protocol-Version = %q, want 1", gotConnectVersion)
		}
		if gotRequest.From != "noreply@matched.infrapad.ai" {
			t.Fatalf("From = %q, want namespace sender", gotRequest.From)
		}
		if len(gotRequest.To) != 1 || gotRequest.To[0] != "user@example.com" {
			t.Fatalf("To = %#v, want single recipient", gotRequest.To)
		}
		if gotRequest.Subject != "Hello" {
			t.Fatalf("Subject = %q, want Hello", gotRequest.Subject)
		}
		if gotRequest.Text != "Hello from your namespace." {
			t.Fatalf("Text = %q, want canned text", gotRequest.Text)
		}

		var response sendEmailResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.From != "noreply@matched.infrapad.ai" || response.Domain != "matched.infrapad.ai" {
			t.Fatalf("response = %#v, want namespace email details", response)
		}
	})

	t.Run("requires platform services config", func(t *testing.T) {
		t.Setenv("PLATFORM_SERVICES_URL", "")
		t.Setenv("PLATFORM_SERVICES_KEY", "")

		req := httptest.NewRequest(http.MethodPost, "/api/email", strings.NewReader(`{"recipient":"user@example.com"}`))
		rec := httptest.NewRecorder()

		sendEmailHandler(http.DefaultClient).ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
	})
}

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
		want := "Hello from Go! MY_SECRET!: (MY_SECRET is not set!)"
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
