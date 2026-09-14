package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/kxs888/kxs/backend/internal/obs"
)

type ctxKey int

const keyPrincipal ctxKey = 1

type Principal struct {
	UserID   uuid.UUID
	Username string
}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	ctx = context.WithValue(ctx, keyPrincipal, p)
	ctx = obs.WithUserID(ctx, p.UserID.String())
	return ctx
}

func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(keyPrincipal).(Principal)
	return p, ok
}

func BearerToken(header string) (string, error) {
	h := strings.TrimSpace(header)
	if h == "" {
		return "", errcode.Unauthorized("missing authorization")
	}
	const prefix = "Bearer "
	if len(h) < len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", errcode.Unauthorized("authorization must be Bearer")
	}
	tok := strings.TrimSpace(h[len(prefix):])
	if tok == "" {
		return "", errcode.Unauthorized("missing access token")
	}
	return tok, nil
}
