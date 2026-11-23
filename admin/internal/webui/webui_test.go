package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func builtFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!doctype html><title>admin</title>")},
		"assets/app-123.js": &fstest.MapFile{
			Data: []byte("console.log('admin')"),
		},
		"assets/app-123.css": &fstest.MapFile{Data: []byte("body{}")},
	}
}

func TestServesIndex(t *testing.T) {
	handler := New(builtFS())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "admin") {
		t.Errorf("body = %q", recorder.Body.String())
	}
	if got := recorder.Header().Get("Cache-Control"); got != indexCache {
		t.Errorf("Cache-Control = %q, want %q", got, indexCache)
	}
}

func TestServesFaviconAsPNG(t *testing.T) {
	fsys := builtFS()
	fsys["favicon.png"] = &fstest.MapFile{Data: []byte("\x89PNG\r\n\x1a\n")}
	handler := New(fsys)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/favicon.png", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != indexCache {
		t.Errorf("Cache-Control = %q, want %q", got, indexCache)
	}
}

func TestServesAssetsAsImmutable(t *testing.T) {
	handler := New(builtFS())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/app-123.js", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); got != assetCache {
		t.Errorf("Cache-Control = %q, want %q", got, assetCache)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/javascript") {
		t.Errorf("Content-Type = %q", got)
	}
}

func TestMissingAssetIs404(t *testing.T) {
	handler := New(builtFS())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestClientRouteFallsBackToIndex(t *testing.T) {
	handler := New(builtFS())

	for _, target := range []string{"/works/new", "/works/0190e0e0-0000-7000-8000-000000000000", "/unknown"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", target, recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), "admin") {
			t.Errorf("%s did not receive index.html", target)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := New(builtFS())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	want := map[string]string{
		"X-Frame-Options":         "DENY",
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "no-referrer",
		"X-Robots-Tag":            "noindex",
		"Content-Security-Policy": contentSecurityPolicy,
	}
	for name, value := range want {
		if got := recorder.Header().Get(name); got != value {
			t.Errorf("%s = %q, want %q", name, got, value)
		}
	}
}

func TestPlaceholderWhenFrontendIsMissing(t *testing.T) {
	handler := New(fstest.MapFS{})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "frontend not built") {
		t.Errorf("body does not mention missing frontend: %q", recorder.Body.String())
	}
}

func TestNilFilesystemUsesPlaceholder(t *testing.T) {
	handler := New(nil)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/works/new", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "frontend not built") {
		t.Errorf("body does not mention missing frontend")
	}
}

func TestRejectsUnsafeMethods(t *testing.T) {
	handler := New(builtFS())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", recorder.Code)
	}
	if got := recorder.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q", got)
	}
}

func TestHeadHasNoBody(t *testing.T) {
	handler := New(builtFS())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodHead, "/assets/app-123.css", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body length = %d, want 0", recorder.Body.Len())
	}
}
