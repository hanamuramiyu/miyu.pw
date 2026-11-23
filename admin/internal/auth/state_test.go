package auth

import (
	"testing"
	"time"
)

func TestStateIsSingleUse(t *testing.T) {
	store := NewStateStore(StateTTL)
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	state, err := store.Issue()
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if state == "" {
		t.Fatal("Issue returned an empty state")
	}

	if !store.Consume(state) {
		t.Fatal("first Consume returned false")
	}
	if store.Consume(state) {
		t.Fatal("second Consume returned true, state is reusable")
	}
}

func TestStateRejectsEmptyAndUnknown(t *testing.T) {
	store := NewStateStore(StateTTL)

	if store.Consume("") {
		t.Fatal("Consume accepted an empty state")
	}
	if store.Consume("unknown") {
		t.Fatal("Consume accepted an unknown state")
	}
}

func TestStateExpires(t *testing.T) {
	store := NewStateStore(StateTTL)
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	state, err := store.Issue()
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(StateTTL + time.Second)
	if store.Consume(state) {
		t.Fatal("Consume accepted an expired state")
	}
}
