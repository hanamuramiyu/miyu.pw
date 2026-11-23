package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeStore struct {
	mu      sync.Mutex
	records map[string]SessionRecord
}

func newFakeStore() *fakeStore {
	return &fakeStore{records: make(map[string]SessionRecord)}
}

func (s *fakeStore) Create(_ context.Context, record SessionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.IDHash] = record
	return nil
}

func (s *fakeStore) Get(_ context.Context, idHash string) (SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[idHash]
	if !ok {
		return SessionRecord{}, ErrSessionNotFound
	}
	return record, nil
}

func (s *fakeStore) Delete(_ context.Context, idHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, idHash)
	return nil
}

func (s *fakeStore) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.records)
}

func TestSessionsStoreOnlyTokenHash(t *testing.T) {
	store := newFakeStore()
	sessions := NewSessions(store, SessionTTL)
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	sessions.now = func() time.Time { return now }

	token, session, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if token == "" || session.CSRFToken == "" {
		t.Fatal("Create returned empty token")
	}
	if token == session.CSRFToken {
		t.Fatal("session and CSRF tokens are identical")
	}

	if _, ok := store.records[token]; ok {
		t.Fatal("plaintext session token is stored in the database")
	}
	record, ok := store.records[hashToken(token)]
	if !ok {
		t.Fatal("session record not found by token hash")
	}
	if record.Login != "octocat" {
		t.Errorf("stored login = %q, want octocat", record.Login)
	}
	if record.CSRFToken != session.CSRFToken {
		t.Errorf("stored CSRF token = %q, want %q", record.CSRFToken, session.CSRFToken)
	}
	if want := now.Add(SessionTTL); !record.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", record.ExpiresAt, want)
	}

	found, err := sessions.Lookup(context.Background(), token)
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if found.Login != "octocat" || found.CSRFToken != session.CSRFToken {
		t.Errorf("Lookup = %+v, want %+v", found, session)
	}
}

func TestSessionsRejectExpiredToken(t *testing.T) {
	store := newFakeStore()
	sessions := NewSessions(store, SessionTTL)
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	sessions.now = func() time.Time { return now }

	token, _, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	now = now.Add(SessionTTL + time.Second)

	if _, err := sessions.Lookup(context.Background(), token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Lookup error = %v, want ErrUnauthenticated", err)
	}
	if store.count() != 0 {
		t.Errorf("expired session was not deleted, %d records left", store.count())
	}
}

func TestSessionsRejectUnknownToken(t *testing.T) {
	sessions := NewSessions(newFakeStore(), SessionTTL)

	if _, err := sessions.Lookup(context.Background(), "unknown"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Lookup error = %v, want ErrUnauthenticated", err)
	}
	if _, err := sessions.Lookup(context.Background(), ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Lookup empty error = %v, want ErrUnauthenticated", err)
	}
}

func TestSessionsDelete(t *testing.T) {
	store := newFakeStore()
	sessions := NewSessions(store, SessionTTL)

	token, _, err := sessions.Create(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := sessions.Delete(context.Background(), token); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if store.count() != 0 {
		t.Errorf("Delete left %d records", store.count())
	}
}
