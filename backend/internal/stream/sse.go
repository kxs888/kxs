package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/obs"
)

// DefaultTicketTTL 是 SSE ticket 默认有效期（C6：expires_at 默认 24h）。
const DefaultTicketTTL = 24 * time.Hour

type Store interface {
	InsertTicket(ctx context.Context, userID uuid.UUID, ttlSeconds int) (*domain.StreamTicket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*domain.StreamTicket, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func New(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) Issue(ctx context.Context, userID uuid.UUID) (*domain.StreamTicket, error) {
	return s.store.InsertTicket(ctx, userID, int(DefaultTicketTTL.Seconds()))
}

func (s *Service) Validate(ctx context.Context, raw string) (*domain.StreamTicket, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, errcode.Unauthorized("invalid stream ticket")
	}
	t, err := s.store.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if !t.ExpiresAt.After(s.now()) {
		return nil, errcode.Unauthorized("stream ticket expired")
	}
	return t, nil
}

// WritePlaceholder 推送非敏感占位事件。禁止推送病历/评估全文。
func WritePlaceholder(w http.ResponseWriter, r *http.Request, ticket *domain.StreamTicket) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeEvent(w, "ready", map[string]any{
		"ok":      true,
		"code":    "OK",
		"channel": "tasks",
		"note":    "placeholder; no PHI",
	})
	fl.Flush()

	writeEvent(w, "task", map[string]any{
		"id":     "00000000-0000-0000-0000-000000000000",
		"status": "placeholder",
	})
	fl.Flush()

	writeEvent(w, "heartbeat", map[string]any{"ts": time.Now().UTC().Format(time.RFC3339)})
	fl.Flush()
	obs.App().InfoContext(r.Context(), "sse.placeholder_done", "ticket", ticket.ID.String())
}

func writeEvent(w http.ResponseWriter, event string, data any) {
	b, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
}
