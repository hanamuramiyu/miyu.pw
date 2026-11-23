package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireSessionRejectsMissingCookie(t *testing.T) {
	middleware := &Middleware{Sessions: NewSessions(newFakeStore(), SessionTTL)}
	handler := middleware.RequireSession(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

func TestCSRFRequiresHeaderOnUnsafeMethods(t *testing.T) {
	store := newFakeStore()
	sessions := NewSessions(store, SessionTTL)
	token, session, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	middleware := &Middleware{Sessions: sessions}
	handler := middleware.RequireSession(CSRF(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	newRequest := func(method, csrf string) *http.Request {
		request := httptest.NewRequest(method, "/auth/logout", nil)
		request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
		if csrf != "" {
			request.Header.Set(CSRFTokenHeader, csrf)
		}
		return request
	}

	t.Run("post without header", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, newRequest(http.MethodPost, ""))
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", recorder.Code)
		}
	})

	t.Run("post with wrong header", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, newRequest(http.MethodPost, "wrong"))
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", recorder.Code)
		}
	})

	t.Run("post with valid header", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, newRequest(http.MethodPost, session.CSRFToken))
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}
	})

	t.Run("get does not need header", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, newRequest(http.MethodGet, ""))
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}
	})
}

func TestCSRFWithoutSessionIsUnauthorized(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

func TestCookieAttributes(t *testing.T) {
	sessionCookie := SessionCookie("token", true, 7*24*time.Hour)
	if !sessionCookie.HttpOnly || !sessionCookie.Secure {
		t.Error("session cookie must be HttpOnly and Secure")
	}
	if sessionCookie.SameSite != http.SameSiteStrictMode {
		t.Error("session cookie must be SameSite=Strict")
	}
	if sessionCookie.Path != "/" {
		t.Errorf("session cookie Path = %q, want /", sessionCookie.Path)
	}
	if sessionCookie.MaxAge != int((7*24*time.Hour)/time.Second) {
		t.Errorf("session cookie MaxAge = %d", sessionCookie.MaxAge)
	}

	stateCookie := StateCookie("state", false, 10*time.Minute)
	if !stateCookie.HttpOnly {
		t.Error("state cookie must be HttpOnly")
	}
	if stateCookie.SameSite != http.SameSiteLaxMode {
		t.Error("state cookie must be SameSite=Lax")
	}
	if stateCookie.Secure {
		t.Error("state cookie Secure = true, want false")
	}
	if stateCookie.MaxAge != 600 {
		t.Errorf("state cookie MaxAge = %d, want 600", stateCookie.MaxAge)
	}
}
