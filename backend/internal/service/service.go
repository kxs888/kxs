package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/audit"
	"github.com/kxs888/kxs/backend/internal/auth"
	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/event"
	"github.com/kxs888/kxs/backend/internal/obs"
)

type UserStore interface {
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Insert(ctx context.Context, u *domain.User) error
	Count(ctx context.Context) (int64, error)
}

type PingStore interface {
	Insert(ctx context.Context, message string) (*domain.PingWrite, error)
}

type AuthService struct {
	users  UserStore
	audit  *audit.Service
	secret string
	ttl    time.Duration
	now    func() time.Time
}

func NewAuth(users UserStore, aud *audit.Service, secret string, ttl time.Duration) *AuthService {
	return &AuthService{users: users, audit: aud, secret: secret, ttl: ttl, now: time.Now}
}

type LoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *AuthService) Login(ctx context.Context, username, password, ip string) (*LoginResult, error) {
	fail := func() (*LoginResult, error) {
		_ = s.audit.Record(ctx, audit.Record{
			Action:       "auth.login_failed",
			ResourceType: "user",
			ResourceID:   username,
			Detail:       map[string]any{"username": username, "result": "invalid_credentials"},
			IP:           ip,
		})
		return nil, errcode.InvalidCredentials()
	}
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return fail()
	}
	if !auth.VerifyPassword(password, u.PasswordHash) {
		return fail()
	}
	tok, exp, err := auth.IssueAccess(s.secret, s.ttl, u.ID, u.Username, s.now())
	if err != nil {
		return nil, errcode.Internal("issue token")
	}
	_ = s.audit.Record(ctx, audit.Record{
		ActorID:      &u.ID,
		Action:       "auth.login",
		ResourceType: "user",
		ResourceID:   u.ID.String(),
		Detail:       map[string]any{"username": u.Username, "result": "ok"},
		IP:           ip,
	})
	return &LoginResult{AccessToken: tok, TokenType: "Bearer", ExpiresIn: exp}, nil
}

func (s *AuthService) Me(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *AuthService) EnsureBootstrap(ctx context.Context, username, password string) error {
	n, err := s.users.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	u := &domain.User{Username: username, PasswordHash: hash, DisplayName: "Demo User"}
	if err := s.users.Insert(ctx, u); err != nil {
		return err
	}
	obs.App().InfoContext(ctx, "bootstrap.user_created", "username", username)
	return nil
}

type PingService struct {
	store  PingStore
	audit  *audit.Service
	outbox *event.Publisher
}

func NewPing(store PingStore, aud *audit.Service, outbox *event.Publisher) *PingService {
	return &PingService{store: store, audit: aud, outbox: outbox}
}

func (s *PingService) Create(ctx context.Context, message, ip string) (*domain.PingWrite, error) {
	if message == "" {
		return nil, errcode.InvalidArgument("message required")
	}
	if len(message) > 256 {
		return nil, errcode.InvalidArgument("message too long")
	}
	row, err := s.store.Insert(ctx, message)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, audit.Record{
		Action:       "ping_write.create",
		ResourceType: "ping_write",
		ResourceID:   row.ID.String(),
		Detail:       map[string]any{"id": row.ID.String(), "result": "ok"},
		IP:           ip,
	})
	_ = s.outbox.Enqueue(ctx, "ping_write.created", map[string]any{"id": row.ID.String()})
	return row, nil
}
