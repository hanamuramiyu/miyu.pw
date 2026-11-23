package auth

import (
	"sync"
	"time"
)

type StateStore struct {
	mu     sync.Mutex
	states map[string]time.Time
	ttl    time.Duration
	now    func() time.Time
}

func NewStateStore(ttl time.Duration) *StateStore {
	return &StateStore{
		states: make(map[string]time.Time),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *StateStore) Issue() (string, error) {
	state, err := randomToken(32)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	s.states[state] = s.now().Add(s.ttl)
	return state, nil
}

func (s *StateStore) Consume(state string) bool {
	if state == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()

	expiresAt, ok := s.states[state]
	if !ok {
		return false
	}
	delete(s.states, state)
	return expiresAt.After(s.now())
}

func (s *StateStore) pruneLocked() {
	now := s.now()
	for state, expiresAt := range s.states {
		if !expiresAt.After(now) {
			delete(s.states, state)
		}
	}
}
