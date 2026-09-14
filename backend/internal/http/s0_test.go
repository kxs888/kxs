package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/audit"
	"github.com/kxs888/kxs/backend/internal/auth"
	"github.com/kxs888/kxs/backend/internal/config"
	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/event"
	"github.com/kxs888/kxs/backend/internal/handler"
	httpx "github.com/kxs888/kxs/backend/internal/http"
	"github.com/kxs888/kxs/backend/internal/http/middleware"
	"github.com/kxs888/kxs/backend/internal/obs"
	"github.com/kxs888/kxs/backend/internal/respond"
	"github.com/kxs888/kxs/backend/internal/service"
	"github.com/kxs888/kxs/backend/internal/sm"
	"github.com/kxs888/kxs/backend/internal/stream"
)

const testSecret = "test-jwt-secret-please-use-32b!!"
const testSMKey = "0123456789abcdeffedcba9876543210"

type env struct {
	ok      bool            `json:"ok"`
	code    string          `json:"code"`
	message string          `json:"message"`
	data    json.RawMessage `json:"data"`
}

func decodeEnv(t *testing.T, b []byte) respond.Envelope {
	t.Helper()
	var e respond.Envelope
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatalf("json %s: %v", b, err)
	}
	return e
}

type memUsers struct {
	mu     sync.Mutex
	byName map[string]*domain.User
	byID   map[uuid.UUID]*domain.User
}

func (m *memUsers) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byName[username]
	if !ok {
		return nil, errcode.NotFound("user")
	}
	cp := *u
	return &cp, nil
}

func (m *memUsers) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, errcode.NotFound("user")
	}
	cp := *u
	return &cp, nil
}

func (m *memUsers) Insert(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.CreatedAt = time.Now().UTC()
	cp := *u
	m.byName[u.Username] = &cp
	m.byID[u.ID] = &cp
	return nil
}

func (m *memUsers) Count(context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.byID)), nil
}

type memPing struct {
	mu   sync.Mutex
	rows []domain.PingWrite
}

func (m *memPing) Insert(_ context.Context, message string) (*domain.PingWrite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := domain.PingWrite{ID: uuid.New(), Message: message, CreatedAt: time.Now().UTC()}
	m.rows = append(m.rows, p)
	return &p, nil
}

type memAudit struct {
	mu   sync.Mutex
	rows []domain.AuditRecord
}

func (m *memAudit) Insert(_ context.Context, rec *domain.AuditRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec.ID = uuid.New()
	rec.CreatedAt = time.Now().UTC()
	m.rows = append(m.rows, *rec)
	return nil
}

type memOutbox struct {
	mu   sync.Mutex
	rows []domain.OutboxEvent
}

func (m *memOutbox) Enqueue(_ context.Context, topic string, payload map[string]any) (*domain.OutboxEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ev := domain.OutboxEvent{ID: uuid.New(), Topic: topic, Payload: payload, Status: domain.OutboxPending, CreatedAt: time.Now().UTC()}
	m.rows = append(m.rows, ev)
	return &ev, nil
}

func (m *memOutbox) ClaimPending(context.Context, int) ([]domain.OutboxEvent, error) { return nil, nil }
func (m *memOutbox) MarkPublished(context.Context, uuid.UUID) error                  { return nil }

type memIdem struct {
	mu   sync.Mutex
	data map[string]*domain.IdempotencyRecord
}

func (m *memIdem) BeginOrGet(_ context.Context, key, fingerprint string) (*domain.IdempotencyRecord, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = map[string]*domain.IdempotencyRecord{}
	}
	if rec, ok := m.data[key]; ok {
		if rec.Fingerprint != fingerprint {
			return nil, false, errcode.New(errcode.IdempotencyKeyConflict, 409, "idempotency key reused with different payload")
		}
		cp := *rec
		return &cp, false, nil
	}
	rec := &domain.IdempotencyRecord{Key: key, Fingerprint: fingerprint, Completed: false}
	m.data[key] = rec
	cp := *rec
	return &cp, true, nil
}

func (m *memIdem) Complete(_ context.Context, key string, status int, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.data[key]
	rec.StatusCode = status
	rec.ResponseBody = append([]byte(nil), body...)
	rec.Completed = true
	return nil
}

type memTickets struct {
	mu   sync.Mutex
	data map[uuid.UUID]*domain.StreamTicket
}

func (m *memTickets) InsertTicket(_ context.Context, userID uuid.UUID, ttlSeconds int) (*domain.StreamTicket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = map[uuid.UUID]*domain.StreamTicket{}
	}
	t := &domain.StreamTicket{ID: uuid.New(), UserID: userID, ExpiresAt: time.Now().Add(time.Duration(ttlSeconds) * time.Second), CreatedAt: time.Now()}
	m.data[t.ID] = t
	cp := *t
	return &cp, nil
}

func (m *memTickets) GetTicket(_ context.Context, id uuid.UUID) (*domain.StreamTicket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.data[id]
	if !ok {
		return nil, errcode.Unauthorized("invalid stream ticket")
	}
	cp := *t
	return &cp, nil
}

type okPing struct{}

func (okPing) Ping(context.Context) error { return nil }

type fx struct {
	h      http.Handler
	users  *memUsers
	pings  *memPing
	audits *memAudit
	idemp  *memIdem
}

func newFX(t *testing.T, smOn bool) *fx {
	t.Helper()
	hash, err := auth.HashPassword("demo-pass-change-me")
	if err != nil {
		t.Fatal(err)
	}
	u := &domain.User{ID: uuid.New(), Username: "demo", PasswordHash: hash, DisplayName: "Demo User"}
	users := &memUsers{byName: map[string]*domain.User{u.Username: u}, byID: map[uuid.UUID]*domain.User{u.ID: u}}
	pings := &memPing{}
	audits := &memAudit{}
	outbox := &memOutbox{}
	idemp := &memIdem{data: map[string]*domain.IdempotencyRecord{}}
	tickets := &memTickets{data: map[uuid.UUID]*domain.StreamTicket{}}
	aud := audit.New(audits)
	pub := event.New(outbox)
	authSvc := service.NewAuth(users, aud, testSecret, 15*time.Minute)
	pingSvc := service.NewPing(pings, aud, pub)
	streamSvc := stream.New(tickets)
	api := handler.New(okPing{}, authSvc, pingSvc, streamSvc)
	cfg := &config.Config{
		HTTPTimeout:     5 * time.Second,
		HTTPBodyLimit:   1 << 20,
		JWTSecret:       testSecret,
		JWTTTL:          15 * time.Minute,
		SMCryptoEnabled: smOn,
		SM4KeyHex:       testSMKey,
	}
	return &fx{
		h:      httpx.NewRouter(httpx.Deps{Cfg: cfg, API: api, Idempotency: idemp}),
		users:  users,
		pings:  pings,
		audits: audits,
		idemp:  idemp,
	}
}

func (f *fx) do(t *testing.T, method, path, body, token, idempKey string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if idempKey != "" {
		req.Header.Set("Idempotency-Key", idempKey)
	}
	rr := httptest.NewRecorder()
	f.h.ServeHTTP(rr, req)
	return rr
}

func (f *fx) login(t *testing.T) string {
	t.Helper()
	rr := f.do(t, http.MethodPost, "/api/v1/auth/login", `{"username":"demo","password":"demo-pass-change-me"}`, "", "")
	if rr.Code != 200 {
		t.Fatalf("login %d %s", rr.Code, rr.Body.String())
	}
	e := decodeEnv(t, rr.Body.Bytes())
	var data struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(e.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.AccessToken == "" {
		t.Fatal("empty token")
	}
	return data.AccessToken
}

func TestHealthReadyLoginMe(t *testing.T) {
	f := newFX(t, false)
	rr := f.do(t, http.MethodGet, "/healthz", "", "", "")
	if rr.Code != 200 {
		t.Fatalf("healthz %d", rr.Code)
	}
	e := decodeEnv(t, rr.Body.Bytes())
	if !e.OK || e.Code != errcode.OK {
		t.Fatalf("%+v", e)
	}
	rr = f.do(t, http.MethodGet, "/readyz", "", "", "")
	if rr.Code != 200 {
		t.Fatalf("readyz %d %s", rr.Code, rr.Body.String())
	}
	tok := f.login(t)
	rr = f.do(t, http.MethodGet, "/api/v1/me", "", tok, "")
	if rr.Code != 200 {
		t.Fatalf("me %d %s", rr.Code, rr.Body.String())
	}
	rr = f.do(t, http.MethodGet, "/api/v1/me", "", "", "")
	if rr.Code != 401 {
		t.Fatalf("me unauth %d", rr.Code)
	}
	e = decodeEnv(t, rr.Body.Bytes())
	if e.Code != errcode.AuthUnauthorized {
		t.Fatalf("code %s", e.Code)
	}
}

func TestPingWritesIdempotent(t *testing.T) {
	f := newFX(t, false)
	tok := f.login(t)
	body := `{"message":"hello-s0"}`
	rr1 := f.do(t, http.MethodPost, "/api/v1/ping-writes", body, tok, "idem-key-abc")
	if rr1.Code != 201 {
		t.Fatalf("first %d %s", rr1.Code, rr1.Body.String())
	}
	rr2 := f.do(t, http.MethodPost, "/api/v1/ping-writes", body, tok, "idem-key-abc")
	if rr2.Code != 201 {
		t.Fatalf("replay %d %s", rr2.Code, rr2.Body.String())
	}
	if rr2.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("want replay header")
	}
	if rr1.Body.String() != rr2.Body.String() {
		t.Fatalf("body mismatch\n%s\n%s", rr1.Body.String(), rr2.Body.String())
	}
	f.pings.mu.Lock()
	n := len(f.pings.rows)
	f.pings.mu.Unlock()
	if n != 1 {
		t.Fatalf("ping_writes rows %d", n)
	}
	rr3 := f.do(t, http.MethodPost, "/api/v1/ping-writes", `{"message":"other"}`, tok, "idem-key-abc")
	if rr3.Code != 409 {
		t.Fatalf("conflict %d %s", rr3.Code, rr3.Body.String())
	}
}

func TestTasksEmptyAndNoPatientCRUD(t *testing.T) {
	f := newFX(t, false)
	tok := f.login(t)
	rr := f.do(t, http.MethodGet, "/api/v1/tasks", "", tok, "")
	if rr.Code != 200 {
		t.Fatalf("tasks %d", rr.Code)
	}
	e := decodeEnv(t, rr.Body.Bytes())
	if !bytes.Contains(e.Data, []byte(`"items"`)) || !bytes.Contains(e.Data, []byte(`[]`)) {
		t.Fatalf("tasks data %s", e.Data)
	}
	rr = f.do(t, http.MethodGet, "/api/v1/patients", "", tok, "")
	if rr.Code != 404 {
		t.Fatalf("patients should not exist, got %d", rr.Code)
	}
}

func TestAuditExplicitNotGlobalPOST(t *testing.T) {
	f := newFX(t, false)
	tok := f.login(t)
	rr := f.do(t, http.MethodPost, "/api/v1/stream/tickets", "", tok, "")
	if rr.Code != 200 {
		t.Fatalf("ticket %d %s", rr.Code, rr.Body.String())
	}
	f.audits.mu.Lock()
	n := len(f.audits.rows)
	actions := make([]string, 0, n)
	for _, r := range f.audits.rows {
		actions = append(actions, r.Action)
	}
	f.audits.mu.Unlock()
	if n != 1 || actions[0] != "auth.login" {
		t.Fatalf("audit should only be explicit login, got %v", actions)
	}
}

func TestSSEPlaceholder(t *testing.T) {
	f := newFX(t, false)
	tok := f.login(t)
	rr := f.do(t, http.MethodPost, "/api/v1/stream/tickets", "", tok, "")
	e := decodeEnv(t, rr.Body.Bytes())
	var data struct {
		Ticket string `json:"ticket"`
	}
	if err := json.Unmarshal(e.Data, &data); err != nil {
		t.Fatal(err)
	}
	rr = f.do(t, http.MethodGet, "/api/v1/stream?ticket="+data.Ticket, "", "", "")
	if rr.Code != 200 {
		t.Fatalf("stream %d %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("ct %s", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: ready") || !strings.Contains(body, `"status":"placeholder"`) {
		t.Fatalf("sse body %s", body)
	}
	if strings.Contains(strings.ToLower(body), "病历") || strings.Contains(body, "medical_record") {
		t.Fatal("sse leaked phi")
	}
}

func TestSMCryptoOffPassthrough(t *testing.T) {
	f := newFX(t, false)
	rr := f.do(t, http.MethodPost, "/api/v1/auth/login", `{"username":"demo","password":"demo-pass-change-me"}`, "", "")
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	if rr.Header().Get("X-SM-Crypto") != "" {
		t.Fatal("sm header should be absent when disabled")
	}
	e := decodeEnv(t, rr.Body.Bytes())
	if !e.OK {
		t.Fatalf("%+v", e)
	}
}

func TestSMCryptoOnEncryptsAPINotHealth(t *testing.T) {
	f := newFX(t, true)
	rr := f.do(t, http.MethodGet, "/healthz", "", "", "")
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	e := decodeEnv(t, rr.Body.Bytes())
	if !e.OK {
		t.Fatal("healthz must stay plaintext")
	}

	key, err := sm.ParseKeyHex(testSMKey)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`{"username":"demo","password":"demo-pass-change-me"}`)
	enc, err := sm.Encrypt(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := sm.MarshalEnvelope(enc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("login sm %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-SM-Crypto") != "sm4-cbc" {
		t.Fatal("want sm header")
	}
	outEnv, err := sm.UnmarshalEnvelope(w.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	pt, err := sm.Decrypt(key, outEnv)
	if err != nil {
		t.Fatal(err)
	}
	inner := decodeEnv(t, pt)
	if !inner.OK {
		t.Fatalf("inner %s", pt)
	}
}

func TestSkipSMPath(t *testing.T) {
	if !middleware.SkipSMPath("/healthz") || !middleware.SkipSMPath("/readyz") || !middleware.SkipSMPath("/metrics") {
		t.Fatal("probes")
	}
	if !middleware.SkipSMPath("/api/v1/stream") || !middleware.SkipSMPath("/stream") {
		t.Fatal("stream")
	}
	if middleware.SkipSMPath("/api/v1/auth/login") {
		t.Fatal("login should encrypt when layer 5 on")
	}
}

func TestAccessLogNoJWT(t *testing.T) {
	var buf bytes.Buffer
	obs.Configure(&buf, 0)
	t.Cleanup(func() { obs.Configure(io.Discard, 0) })
	f := newFX(t, false)
	_ = f.login(t)
	s := buf.String()
	if strings.Contains(s, "eyJ") {
		t.Fatalf("jwt in logs: %s", s)
	}
}
