package publish

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hanamuramiyu/miyu.pw/admin/internal/content"
)

func newTestService(t *testing.T) (*Service, string, *fakeStore) {
	t.Helper()

	dir := t.TempDir()
	work := publishedWork()
	store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}
	return NewService(store, dir), dir, store
}

func TestServiceRequestExportsAndWritesTrigger(t *testing.T) {
	service, dir, _ := newTestService(t)

	if err := service.Request(context.Background(), "octocat"); err != nil {
		t.Fatalf("Request: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ExportDirName, "works", "monban", "index.md")); err != nil {
		t.Fatalf("export file: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, TriggerFileName))
	if err != nil {
		t.Fatalf("read trigger: %v", err)
	}

	var request triggerRequest
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatalf("decode trigger %q: %v", data, err)
	}
	if request.By != "octocat" {
		t.Errorf("by = %q, want octocat", request.By)
	}
	if _, err := time.Parse(time.RFC3339, request.RequestedAt); err != nil {
		t.Errorf("requestedAt = %q, want RFC3339: %v", request.RequestedAt, err)
	}
}

func TestServiceStatusIsIdleWithoutFile(t *testing.T) {
	service, _, _ := newTestService(t)

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.State != "idle" {
		t.Errorf("state = %q, want idle", status.State)
	}
}

func TestServiceStatusReadsFile(t *testing.T) {
	_, dir, _ := newTestService(t)
	service := NewService(nil, dir)

	payload := `{"state":"ok","startedAt":"2026-11-22T10:00:00Z","finishedAt":"2026-11-22T10:01:00Z","release":"20261122T100100Z","error":""}`
	if err := os.WriteFile(filepath.Join(dir, StatusFileName), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.State != "ok" || status.Release != "20261122T100100Z" {
		t.Errorf("status = %+v, want ok release 20261122T100100Z", status)
	}
}

func TestServiceStatusRejectsBrokenFile(t *testing.T) {
	_, dir, _ := newTestService(t)
	service := NewService(nil, dir)

	for _, payload := range []string{"{", `{"startedAt":"now"}`} {
		if err := os.WriteFile(filepath.Join(dir, StatusFileName), []byte(payload), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := service.Status(context.Background()); err == nil {
			t.Fatalf("Status succeeded for payload %q, want error", payload)
		}
	}
}

func TestServiceRequestRequiresStore(t *testing.T) {
	service := NewService(nil, t.TempDir())
	err := service.Request(context.Background(), "octocat")
	if err == nil || !strings.Contains(err.Error(), "store") {
		t.Fatalf("Request error = %v, want a store error", err)
	}
}

func TestServiceRequestKeepsPreviousExportOnFailure(t *testing.T) {
	dir := t.TempDir()
	work := publishedWork()
	work.Tools = nil
	store := &fakeStore{summaries: []content.WorkSummary{summary(work)}, works: map[string]*content.Work{work.ID: work}}
	service := NewService(store, dir)

	previous := filepath.Join(dir, ExportDirName, "works", "previous", "index.md")
	if err := os.MkdirAll(filepath.Dir(previous), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(previous, []byte("previous"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := service.Request(context.Background(), "octocat")
	var invalid *InvalidWorkError
	if !errors.As(err, &invalid) {
		t.Fatalf("Request error = %v, want InvalidWorkError", err)
	}
	if _, statErr := os.Stat(previous); statErr != nil {
		t.Fatalf("previous export was removed: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(dir, TriggerFileName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("trigger status = %v, want not exist", statErr)
	}
}
