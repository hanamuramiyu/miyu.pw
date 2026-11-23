package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsTenAndRejectsEleventh(t *testing.T) {
	limiter := NewRateLimiter(10, time.Minute)
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := func(remoteAddr string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, "/auth/github", nil)
		httpRequest.RemoteAddr = remoteAddr
		handler.ServeHTTP(recorder, httpRequest)
		return recorder
	}

	for attempt := 1; attempt <= 10; attempt++ {
		if got := request("192.0.2.10:5000").Code; got != http.StatusOK {
			t.Fatalf("attempt %d status = %d, want 200", attempt, got)
		}
	}

	if got := request("192.0.2.10:5000").Code; got != http.StatusTooManyRequests {
		t.Fatalf("eleventh attempt status = %d, want 429", got)
	}

	if got := request("192.0.2.11:5000").Code; got != http.StatusOK {
		t.Fatalf("other client status = %d, want 200", got)
	}

	now = now.Add(time.Minute + time.Second)
	if got := request("192.0.2.10:5000").Code; got != http.StatusOK {
		t.Fatalf("after window status = %d, want 200", got)
	}
}

func TestRateLimiterUsesForwardedFor(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	limiter.now = func() time.Time { return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC) }

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := func(header, value string) int {
		recorder := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, "/auth/github", nil)
		httpRequest.RemoteAddr = "127.0.0.1:5000"
		httpRequest.Header.Set(header, value)
		handler.ServeHTTP(recorder, httpRequest)
		return recorder.Code
	}

	if got := request("X-Real-IP", "198.51.100.1"); got != http.StatusOK {
		t.Fatalf("first status = %d, want 200", got)
	}
	if got := request("X-Real-IP", "198.51.100.1"); got != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want 429", got)
	}
	if got := request("X-Real-IP", "198.51.100.2"); got != http.StatusOK {
		t.Fatalf("other real ip status = %d, want 200", got)
	}

	if got := request("X-Forwarded-For", "203.0.113.5"); got != http.StatusOK {
		t.Fatalf("forwarded first status = %d, want 200", got)
	}
	if got := request("X-Forwarded-For", "203.0.113.5"); got != http.StatusTooManyRequests {
		t.Fatalf("forwarded second status = %d, want 429", got)
	}
	if got := request("X-Forwarded-For", "spoofed, 203.0.113.6"); got != http.StatusOK {
		t.Fatalf("last forwarded address status = %d, want 200", got)
	}
}
