package idempotency

import (
	"context"
	"sync"
	"time"

	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/errcode"
)

// Memory 供单测与无库回退。生产走 Postgres repo。
type Memory struct {
	mu   sync.Mutex
	data map[string]*domain.IdempotencyRecord
	now  func() time.Time
}

func NewMemory() *Memory {
	return &Memory{data: map[string]*domain.IdempotencyRecord{}, now: time.Now}
}

func (m *Memory) BeginOrGet(_ context.Context, key, fingerprint string) (*domain.IdempotencyRecord, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = map[string]*domain.IdempotencyRecord{}
	}
	now := m.now()
	if rec, ok := m.data[key]; ok {
		if !rec.ExpiresAt.IsZero() && !rec.ExpiresAt.After(now) {
			delete(m.data, key)
		} else {
			if rec.Fingerprint != fingerprint {
				return nil, false, errcode.IdempotencyConflict("idempotency key reused with different payload")
			}
			cp := *rec
			return &cp, false, nil
		}
	}
	rec := &domain.IdempotencyRecord{
		Key:         key,
		Fingerprint: fingerprint,
		Completed:   false,
		ExpiresAt:   now.Add(DefaultTTL),
	}
	m.data[key] = rec
	cp := *rec
	return &cp, true, nil
}

func (m *Memory) Complete(_ context.Context, key string, status int, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.data[key]
	if !ok || rec.Completed {
		return nil
	}
	rec.StatusCode = status
	rec.ResponseBody = append([]byte(nil), body...)
	rec.Completed = true
	return nil
}
