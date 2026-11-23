package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/auth"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/content"
	"github.com/hanamuramiyu/miyu.pw/admin/internal/publish"
)

func publishTestWork() *content.Work {
	return &content.Work{
		ID:        testWorkID,
		Slug:      "monban",
		Title:     "monban",
		Summary:   "Modern Minecraft access control",
		BodyMD:    "## Why\n\nText.\n",
		Date:      "2025-11-07",
		Updated:   "2026-10-22",
		Published: true,
		Tools:     []string{"Java"},
	}
}

func publishTestHandler(t *testing.T) (http.Handler, *auth.Sessions, string, *fakeWorks) {
	t.Helper()

	work := publishTestWork()
	works := &fakeWorks{
		list: []content.WorkSummary{{ID: work.ID, Slug: work.Slug, Published: true}},
		work: work,
	}
	dir := t.TempDir()
	handler, sessions := newHandlerWithPublisher(t, fakePinger{}, works, publish.NewService(works, dir))
	return handler, sessions, dir, works
}

func TestPublishRoutesRequireSession(t *testing.T) {
	handler, _, _, _ := publishTestHandler(t)

	requests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/publish"},
		{http.MethodGet, "/api/publish/status"},
	}

	for _, request := range requests {
		t.Run(request.method+" "+request.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(request.method, request.path, nil))

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", recorder.Code)
			}
		})
	}
}

func TestPublishRequestRequiresCSRF(t *testing.T) {
	handler, sessions, dir, _ := publishTestHandler(t)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost, "/api/publish", "", false))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	if _, err := os.Stat(filepath.Join(dir, publish.TriggerFileName)); !os.IsNotExist(err) {
		t.Errorf("trigger written without CSRF: %v", err)
	}
}

func TestPublishRequestQueuesExport(t *testing.T) {
	handler, sessions, dir, _ := publishTestHandler(t)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost, "/api/publish", "", true))

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202: %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]string
	decodeResponse(t, recorder, &body)
	if body["state"] != "queued" {
		t.Errorf("state = %q, want queued", body["state"])
	}

	if _, err := os.Stat(filepath.Join(dir, publish.TriggerFileName)); err != nil {
		t.Errorf("trigger file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, publish.ExportDirName, "works", "monban", "index.md")); err != nil {
		t.Errorf("export file: %v", err)
	}
}

func TestPublishRequestRejectsInvalidWork(t *testing.T) {
	work := publishTestWork()
	work.Tools = nil
	works := &fakeWorks{
		list: []content.WorkSummary{{ID: work.ID, Slug: work.Slug, Published: true}},
		work: work,
	}
	dir := t.TempDir()
	handler, sessions := newHandlerWithPublisher(t, fakePinger{}, works, publish.NewService(works, dir))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodPost, "/api/publish", "", true))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]string
	decodeResponse(t, recorder, &body)
	if !strings.Contains(body["error"], "at least one tool") {
		t.Errorf("error = %q, want a tool hint", body["error"])
	}

	if _, err := os.Stat(filepath.Join(dir, publish.TriggerFileName)); !os.IsNotExist(err) {
		t.Errorf("trigger written for an invalid work: %v", err)
	}
}

func TestPublishStatusIsIdleWithoutFile(t *testing.T) {
	handler, sessions, _, _ := publishTestHandler(t)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodGet, "/api/publish/status", "", false))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]string
	decodeResponse(t, recorder, &body)
	if body["state"] != "idle" {
		t.Errorf("state = %q, want idle", body["state"])
	}
}

func TestPublishStatusReadsFile(t *testing.T) {
	handler, sessions, dir, _ := publishTestHandler(t)

	if err := os.WriteFile(
		filepath.Join(dir, publish.StatusFileName),
		[]byte(`{"state":"ok","startedAt":"2026-11-22T10:00:00Z","finishedAt":"2026-11-22T10:01:00Z","release":"20261122T100100Z","error":""}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authenticatedRequest(t, sessions, http.MethodGet, "/api/publish/status", "", false))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		State      string `json:"state"`
		StartedAt  string `json:"startedAt"`
		FinishedAt string `json:"finishedAt"`
		Release    string `json:"release"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	if body.State != "ok" || body.Release != "20261122T100100Z" {
		t.Errorf("body = %+v, want ok release 20261122T100100Z", body)
	}
}

func TestPublishMethodNotAllowed(t *testing.T) {
	handler, _, _, _ := publishTestHandler(t)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/publish", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", recorder.Code)
	}
	if got := recorder.Header().Get("Allow"); got != "POST" {
		t.Errorf("Allow = %q, want POST", got)
	}
}
