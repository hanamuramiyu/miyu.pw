package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, record SessionRecord) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, login, csrf_token, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		record.IDHash, record.Login, record.CSRFToken, record.CreatedAt, record.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *PostgresStore) Get(ctx context.Context, idHash string) (SessionRecord, error) {
	var record SessionRecord
	err := s.pool.QueryRow(ctx, `
		SELECT id, login, csrf_token, created_at, expires_at
		FROM sessions
		WHERE id = $1`, idHash,
	).Scan(&record.IDHash, &record.Login, &record.CSRFToken, &record.CreatedAt, &record.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionRecord{}, ErrSessionNotFound
	}
	if err != nil {
		return SessionRecord{}, fmt.Errorf("select session: %w", err)
	}
	return record, nil
}

func (s *PostgresStore) Delete(ctx context.Context, idHash string) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, idHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
