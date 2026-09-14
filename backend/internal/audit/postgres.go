package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kxs888/kxs/backend/internal/domain"
)

// Postgres 将审计写入 audit_logs。S1 前必须走此实现（有 DATABASE_URL 时）。
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	if pool == nil {
		return nil
	}
	return &Postgres{pool: pool}
}

func (p *Postgres) Insert(ctx context.Context, rec *domain.AuditRecord) error {
	if p == nil || p.pool == nil {
		return nil
	}
	if rec.Detail == nil {
		rec.Detail = map[string]any{}
	}
	detail, err := json.Marshal(SanitizeDetail(rec.Detail))
	if err != nil {
		return err
	}
	rec.Detail = SanitizeDetail(rec.Detail)
	return p.pool.QueryRow(ctx, `
		INSERT INTO audit_logs (
			actor_id, action, resource_type, resource_id, outcome,
			detail_json, ip, user_agent, request_id, trace_id
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)
		RETURNING id, created_at`,
		rec.ActorID, rec.Action, rec.ResourceType, rec.ResourceID, rec.Outcome,
		detail, rec.IP, rec.UserAgent, rec.RequestID, rec.TraceID,
	).Scan(&rec.ID, &rec.CreatedAt)
}
