package repo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/errcode"
)

type PingWriteRepo struct {
	pool *pgxpool.Pool
}

func NewPingWriteRepo(pool *pgxpool.Pool) *PingWriteRepo {
	return &PingWriteRepo{pool: pool}
}

func (r *PingWriteRepo) Insert(ctx context.Context, message string) (*domain.PingWrite, error) {
	var p domain.PingWrite
	err := r.pool.QueryRow(ctx, `
		INSERT INTO ping_writes (message) VALUES ($1)
		RETURNING id, message, created_at`, message).
		Scan(&p.ID, &p.Message, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PingWriteRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ping_writes`).Scan(&n)
	return n, err
}

type IdempotencyRepo struct {
	pool *pgxpool.Pool
}

func NewIdempotencyRepo(pool *pgxpool.Pool) *IdempotencyRepo {
	return &IdempotencyRepo{pool: pool}
}

func (r *IdempotencyRepo) BeginOrGet(ctx context.Context, key, fingerprint string) (*domain.IdempotencyRecord, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var rec domain.IdempotencyRecord
	var body []byte
	err = tx.QueryRow(ctx, `
		SELECT key, request_fingerprint, COALESCE(status_code, 0), COALESCE(response_body, ''), completed, expires_at
		FROM idempotency_keys WHERE key = $1 FOR UPDATE`, key).
		Scan(&rec.Key, &rec.Fingerprint, &rec.StatusCode, &body, &rec.Completed, &rec.ExpiresAt)
	if err == nil {
		if !rec.ExpiresAt.IsZero() && !rec.ExpiresAt.After(time.Now()) {
			if _, delErr := tx.Exec(ctx, `DELETE FROM idempotency_keys WHERE key = $1`, key); delErr != nil {
				return nil, false, delErr
			}
		} else {
			if rec.Fingerprint != fingerprint {
				return nil, false, errcode.IdempotencyConflict("idempotency key reused with different payload")
			}
			rec.ResponseBody = body
			if err := tx.Commit(ctx); err != nil {
				return nil, false, err
			}
			return &rec, false, nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO idempotency_keys (key, request_fingerprint, completed, expires_at)
		VALUES ($1, $2, FALSE, now() + interval '24 hours')`, key, fingerprint)
	if err != nil {
		return nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return &domain.IdempotencyRecord{
		Key:         key,
		Fingerprint: fingerprint,
		Completed:   false,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, true, nil
}

func (r *IdempotencyRepo) Complete(ctx context.Context, key string, status int, responseBody []byte) error {
	if len(responseBody) == 0 {
		responseBody = []byte("null")
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE idempotency_keys
		SET status_code = $2, response_body = $3, completed = TRUE
		WHERE key = $1 AND completed = FALSE`, key, status, string(responseBody))
	return err
}

type OutboxRepo struct {
	pool *pgxpool.Pool
}

func NewOutboxRepo(pool *pgxpool.Pool) *OutboxRepo {
	return &OutboxRepo{pool: pool}
}

func (r *OutboxRepo) Enqueue(ctx context.Context, topic string, payload map[string]any) (*domain.OutboxEvent, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var ev domain.OutboxEvent
	var raw []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO outbox_events (topic, payload_json, status)
		VALUES ($1, $2::jsonb, $3)
		RETURNING id, topic, payload_json, status, created_at`,
		topic, b, domain.OutboxPending,
	).Scan(&ev.ID, &ev.Topic, &raw, &ev.Status, &ev.CreatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(raw, &ev.Payload)
	return &ev, nil
}

func (r *OutboxRepo) ClaimPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, topic, payload_json, status, created_at
		FROM outbox_events
		WHERE status = $1
		ORDER BY created_at
		LIMIT $2`, domain.OutboxPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.OutboxEvent
	for rows.Next() {
		var ev domain.OutboxEvent
		var raw []byte
		if err := rows.Scan(&ev.ID, &ev.Topic, &raw, &ev.Status, &ev.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &ev.Payload)
		out = append(out, ev)
	}
	return out, rows.Err()
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_events SET status = $2, published_at = now(), attempts = attempts + 1
		WHERE id = $1`, id, domain.OutboxPublished)
	return err
}

type StreamRepo struct {
	pool *pgxpool.Pool
}

func NewStreamRepo(pool *pgxpool.Pool) *StreamRepo {
	return &StreamRepo{pool: pool}
}

func (r *StreamRepo) InsertTicket(ctx context.Context, userID uuid.UUID, ttlSeconds int) (*domain.StreamTicket, error) {
	var t domain.StreamTicket
	err := r.pool.QueryRow(ctx, `
		INSERT INTO stream_tickets (user_id, expires_at)
		VALUES ($1, now() + make_interval(secs => $2::int))
		RETURNING id, user_id, expires_at, created_at`, userID, ttlSeconds).
		Scan(&t.ID, &t.UserID, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *StreamRepo) GetTicket(ctx context.Context, id uuid.UUID) (*domain.StreamTicket, error) {
	var t domain.StreamTicket
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, created_at FROM stream_tickets WHERE id = $1`, id).
		Scan(&t.ID, &t.UserID, &t.ExpiresAt, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errcode.Unauthorized("invalid stream ticket")
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
