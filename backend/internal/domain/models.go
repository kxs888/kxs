package domain

import (
	"time"

	"github.com/google/uuid"
)

// User 是登录占位主体。S0 不做完整账号 CRUD。
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-" pii:"mask"`
	DisplayName  string    `json:"display_name"`
	Permissions  []string  `json:"permissions,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type PingWrite struct {
	ID        uuid.UUID `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditRecord struct {
	ID           uuid.UUID      `json:"id"`
	ActorID      *uuid.UUID     `json:"actor_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id,omitempty"`
	Outcome      string         `json:"outcome,omitempty"`
	Detail       map[string]any `json:"detail"`
	IP           string         `json:"ip,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	RequestID    string         `json:"request_id,omitempty"`
	TraceID      string         `json:"trace_id,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type OutboxEvent struct {
	ID        uuid.UUID      `json:"id"`
	Topic     string         `json:"topic"`
	Payload   map[string]any `json:"payload"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

const (
	OutboxPending   = "pending"
	OutboxPublished = "published"
	OutboxFailed    = "failed"
)

type StreamTicket struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type IdempotencyRecord struct {
	Key          string
	Fingerprint  string
	StatusCode   int
	ResponseBody []byte
	Completed    bool
	ExpiresAt    time.Time
}
