package auth

import (
	"context"
	"errors"
	"time"
)

const SessionTTL = 7 * 24 * time.Hour

var ErrSessionNotFound = errors.New("session not found")

var ErrUnauthenticated = errors.New("unauthenticated")

type SessionRecord struct {
	IDHash    string
	Login     string
	CSRFToken string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Session struct {
	Login     string
	CSRFToken string
	ExpiresAt time.Time
}

type Store interface {
	Create(ctx context.Context, record SessionRecord) error
	Get(ctx context.Context, idHash string) (SessionRecord, error)
	Delete(ctx context.Context, idHash string) error
}

type Sessions struct {
	store Store
	ttl   time.Duration
	now   func() time.Time
}

func NewSessions(store Store, ttl time.Duration) *Sessions {
	return &Sessions{
		store: store,
		ttl:   ttl,
		now:   time.Now,
	}
}

func (s *Sessions) Create(ctx context.Context, login string) (string, Session, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", Session{}, err
	}
	csrfToken, err := randomToken(32)
	if err != nil {
		return "", Session{}, err
	}

	now := s.now()
	record := SessionRecord{
		IDHash:    hashToken(token),
		Login:     login,
		CSRFToken: csrfToken,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}
	if err := s.store.Create(ctx, record); err != nil {
		return "", Session{}, err
	}

	return token, Session{
		Login:     login,
		CSRFToken: csrfToken,
		ExpiresAt: record.ExpiresAt,
	}, nil
}

func (s *Sessions) Lookup(ctx context.Context, token string) (Session, error) {
	if token == "" {
		return Session{}, ErrUnauthenticated
	}

	idHash := hashToken(token)
	record, err := s.store.Get(ctx, idHash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return Session{}, ErrUnauthenticated
		}
		return Session{}, err
	}

	if !record.ExpiresAt.After(s.now()) {
		_ = s.store.Delete(ctx, idHash)
		return Session{}, ErrUnauthenticated
	}

	return Session{
		Login:     record.Login,
		CSRFToken: record.CSRFToken,
		ExpiresAt: record.ExpiresAt,
	}, nil
}

func (s *Sessions) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.Delete(ctx, hashToken(token))
}
