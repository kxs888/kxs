// Package audit 仅提供显式 Record。禁止做成「所有 POST 自动记」的全局中间件。
package audit

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/auth"
	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/obs"
)

type Store interface {
	Insert(ctx context.Context, rec *domain.AuditRecord) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

type Record struct {
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Outcome      string
	Detail       map[string]any
	IP           string
	UserAgent    string
	RequestID    string
	TraceID      string
}

// Record 由 service 在关键写路径显式调用。detail_json 不得包含病历全文或 JWT。
func (s *Service) Record(ctx context.Context, rec Record) error {
	if s == nil || s.store == nil {
		return nil
	}
	if rec.ActorID == nil {
		if p, ok := auth.PrincipalFrom(ctx); ok {
			id := p.UserID
			rec.ActorID = &id
		}
	}
	if rec.RequestID == "" {
		rec.RequestID = obs.RequestIDFrom(ctx)
	}
	if rec.TraceID == "" {
		rec.TraceID = obs.TraceIDFrom(ctx)
	}
	if rec.Outcome == "" {
		rec.Outcome = defaultOutcome(rec.Action)
	}
	row := &domain.AuditRecord{
		ActorID:      rec.ActorID,
		Action:       rec.Action,
		ResourceType: rec.ResourceType,
		ResourceID:   rec.ResourceID,
		Outcome:      rec.Outcome,
		Detail:       SanitizeDetail(rec.Detail),
		IP:           rec.IP,
		UserAgent:    rec.UserAgent,
		RequestID:    rec.RequestID,
		TraceID:      rec.TraceID,
	}
	if err := s.store.Insert(ctx, row); err != nil {
		obs.AuditLog().ErrorContext(ctx, "audit.insert_failed", "action", rec.Action, "err", err.Error())
		return err
	}
	obs.AuditLog().InfoContext(ctx, "audit.recorded",
		"action", rec.Action,
		"resource_type", rec.ResourceType,
		"resource_id", rec.ResourceID,
		"outcome", rec.Outcome,
	)
	return nil
}

func defaultOutcome(action string) string {
	if strings.Contains(action, "fail") {
		return "failure"
	}
	return "success"
}

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		part, _, _ := strings.Cut(xff, ",")
		return strings.TrimSpace(part)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
