package publish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	DefaultDir      = "/var/lib/miyu-admin"
	ExportDirName   = "export"
	TriggerFileName = "publish-request"
	StatusFileName  = "publish-status.json"
)

type Status struct {
	State      string `json:"state"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
	Release    string `json:"release,omitempty"`
	Error      string `json:"error,omitempty"`
}

type triggerRequest struct {
	RequestedAt string `json:"requestedAt"`
	By          string `json:"by"`
}

type Service struct {
	store Store
	dir   string
	mu    sync.Mutex
}

func NewService(store Store, dir string) *Service {
	if dir == "" {
		dir = DefaultDir
	}
	return &Service{store: store, dir: dir}
}

func (s *Service) Request(ctx context.Context, login string) error {
	if s == nil || s.store == nil {
		return errors.New("publish: store is not configured")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return fmt.Errorf("create publish state directory: %w", err)
	}
	if err := Export(ctx, s.store, filepath.Join(s.dir, ExportDirName)); err != nil {
		return err
	}

	request := triggerRequest{
		RequestedAt: time.Now().UTC().Format(time.RFC3339),
		By:          login,
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("encode publish request: %w", err)
	}
	payload = append(payload, '\n')

	tmp, err := os.CreateTemp(s.dir, ".publish-request-")
	if err != nil {
		return fmt.Errorf("create publish request: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write publish request: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write publish request: %w", err)
	}
	if err := os.Chmod(tmpName, 0o640); err != nil {
		return fmt.Errorf("prepare publish request: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(s.dir, TriggerFileName)); err != nil {
		return fmt.Errorf("install publish request: %w", err)
	}
	return nil
}

func (s *Service) Status(_ context.Context) (Status, error) {
	if s == nil || s.dir == "" {
		return Status{}, errors.New("publish: state directory is not configured")
	}

	data, err := os.ReadFile(filepath.Join(s.dir, StatusFileName))
	if errors.Is(err, os.ErrNotExist) {
		return Status{State: "idle"}, nil
	}
	if err != nil {
		return Status{}, fmt.Errorf("read publish status: %w", err)
	}

	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return Status{}, fmt.Errorf("decode publish status: %w", err)
	}
	if status.State == "" {
		return Status{}, fmt.Errorf("decode publish status: state is empty")
	}
	return status, nil
}
