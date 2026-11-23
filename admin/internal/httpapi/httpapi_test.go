package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/auth"
)

type memStore struct {
	mu      sync.Mutex
	records map[string]auth.SessionRecord
}

func newMemStore() *memStore {
	return &memStore{records: make(map[string]auth.SessionRecord)}
}

func (s *memStore) Create(_ context.Context, record auth.SessionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.IDHash] = record
	return nil
}

func (s *memStore) Get(_ context.Context, idHash string) (auth.SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[idHash]
	if !ok {
		return auth.SessionRecord{}, auth.ErrSessionNotFound
	}
	return record, nil
}

func (s *memStore) Delete(_ context.Context, idHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, idHash)
	return nil
}

type fakePinger struct {
	err error
}

func (p fakePinger) Ping(context.Context) error {
	return p.err
}

func newHandler(t *testing.T, health Pinger) (http.Handler, *auth.Sessions) {
	t.Helper()
	return newHandlerWithWorks(t, health, nil)
}

func newHandlerWithWorks(t *testing.T, health Pinger, works WorkStore) (http.Handler, *auth.Sessions) {
	t.Helper()
	return newHandlerWithUI(t, health, works, nil)
}

func newHandlerWithUI(t *testing.T, health Pinger, works WorkStore, ui http.Handler) (http.Handler, *auth.Sessions) {
	t.Helper()
	return newHandlerFull(t, health, works, ui, nil)
}

func newHandlerWithPublisher(t *testing.T, health Pinger, works WorkStore, publisher Publisher) (http.Handler, *auth.Sessions) {
	t.Helper()
	return newHandlerFull(t, health, works, nil, publisher)
}

func newHandlerFull(t *testing.T, health Pinger, works WorkStore, ui http.Handler, publisher Publisher) (http.Handler, *auth.Sessions) {
	t.Helper()

	sessions := auth.NewSessions(newMemStore(), auth.SessionTTL)
	handlers := auth.NewHandlers(auth.HandlerOptions{
		Sessions: sessions,
		States:   auth.NewStateStore(auth.StateTTL),
		Logger:   log.New(io.Discard, "", 0),
	})

	return New(Options{
		Sessions:  sessions,
		Auth:      handlers,
		Limiter:   auth.NewRateLimiter(100, time.Minute),
		Health:    health,
		Works:     works,
		Publisher: publisher,
		UI:        ui,
		Logger:    log.New(io.Discard, "", 0),
	}), sessions
}

func TestHealthz(t *testing.T) {
	handler, _ := newHandler(t, fakePinger{})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var body map[string]bool
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body["ok"] {
		t.Errorf("body = %v, want ok=true", body)
	}
}

func TestReadyz(t *testing.T) {
	t.Run("database reachable", func(t *testing.T) {
		handler, _ := newHandler(t, fakePinger{})

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}
	})

	t.Run("database unavailable", func(t *testing.T) {
		handler, _ := newHandler(t, fakePinger{err: errors.New("connection refused")})

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", recorder.Code)
		}
	})
}

func TestUnknownRouteIsJSONNotFound(t *testing.T) {
	handler, _ := newHandler(t, fakePinger{})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/unknown", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
}

func TestUnknownAPIPathIsJSONNotFound(t *testing.T) {
	handler, _ := newHandler(t, fakePinger{})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}

	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "not found" {
		t.Errorf("error = %q, want not found", body["error"])
	}
}

func TestUnsupportedMethodIsJSONMethodNotAllowed(t *testing.T) {
	handler, _ := newHandler(t, fakePinger{})

	tests := []struct {
		method string
		path   string
		allow  string
	}{
		{http.MethodDelete, "/api/me", "GET"},
		{http.MethodPost, "/healthz", "GET"},
		{http.MethodGet, "/auth/logout", "POST"},
		{http.MethodPatch, "/api/works/0190e0e0-0000-7000-8000-000000000000", "DELETE, GET, PUT"},
	}

	for _, test := range tests {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, nil))

		if recorder.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s %s status = %d, want 405", test.method, test.path, recorder.Code)
		}
		if got := recorder.Header().Get("Allow"); got != test.allow {
			t.Errorf("%s %s Allow = %q, want %q", test.method, test.path, got, test.allow)
		}
		if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("%s %s Content-Type = %q", test.method, test.path, got)
		}

		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s %s decode body: %v", test.method, test.path, err)
		}
		if body["error"] != "method not allowed" {
			t.Errorf("%s %s error = %q, want method not allowed", test.method, test.path, body["error"])
		}
	}
}

func TestPanicReturnsJSONInternalError(t *testing.T) {
	ui := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	handler, _ := newHandlerWithUI(t, fakePinger{}, nil, ui)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := recorder.Body.String(); strings.Contains(got, "boom") || strings.Contains(got, "goroutine") {
		t.Errorf("response leaks the panic stack: %q", got)
	}

	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "internal error" {
		t.Errorf("error = %q, want internal error", body["error"])
	}
}

func TestUIServesClientRoutesAndAPIKeepsPriority(t *testing.T) {
	ui := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, "admin ui")
	})
	handler, _ := newHandlerWithUI(t, fakePinger{}, nil, ui)

	clientRoutes := []string{"/", "/works/new", "/works/0190e0e0-0000-7000-8000-000000000000"}
	for _, target := range clientRoutes {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", target, recorder.Code)
		}
		if got := recorder.Body.String(); got != "admin ui" {
			t.Errorf("%s body = %q, want admin ui", target, got)
		}
	}

	jsonRoutes := []string{"/api/unknown", "/auth/unknown"}
	for _, target := range jsonRoutes {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", target, recorder.Code)
		}
		if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("%s Content-Type = %q, want JSON", target, got)
		}
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/works", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("/api/works status = %d, want 401", recorder.Code)
	}
}

func TestMeWithoutSession(t *testing.T) {
	handler, _ := newHandler(t, fakePinger{})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

func TestLogoutDeletesSession(t *testing.T) {
	handler, sessions := newHandler(t, fakePinger{})
	token, session, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	request.Header.Set(auth.CSRFTokenHeader, session.CSRFToken)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if _, err := sessions.Lookup(context.Background(), token); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("Lookup after logout error = %v, want ErrUnauthenticated", err)
	}
}

func TestLogoutRequiresCSRF(t *testing.T) {
	handler, sessions := newHandler(t, fakePinger{})
	token, _, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	if _, err := sessions.Lookup(context.Background(), token); err != nil {
		t.Fatalf("session was deleted without CSRF: %v", err)
	}
}

func TestMeReturnsLoginAndCSRFToken(t *testing.T) {
	handler, sessions := newHandler(t, fakePinger{})
	token, session, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var body struct {
		Login     string `json:"login"`
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Login != "octocat" {
		t.Errorf("login = %q, want octocat", body.Login)
	}
	if body.CSRFToken != session.CSRFToken {
		t.Errorf("csrfToken = %q, want %q", body.CSRFToken, session.CSRFToken)
	}
}
