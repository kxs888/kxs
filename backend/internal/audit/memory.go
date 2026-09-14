package audit

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/domain"
)

// Memory 是 audit_logs 的内存实现，供单测与无库场景。生产走 DB repo。
type Memory struct {
	mu   sync.Mutex
	rows []domain.AuditRecord
}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Insert(_ context.Context, rec *domain.AuditRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *rec
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = time.Now().UTC()
	}
	if cp.Detail == nil {
		cp.Detail = map[string]any{}
	}
	rec.ID = cp.ID
	rec.CreatedAt = cp.CreatedAt
	m.rows = append(m.rows, cp)
	return nil
}

func (m *Memory) Snapshot() []domain.AuditRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.AuditRecord, len(m.rows))
	copy(out, m.rows)
	return out
}
