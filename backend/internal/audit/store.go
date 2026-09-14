package audit

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewStore：有 pgx pool 则持久化 Postgres；pool 为 nil（无 DATABASE_URL）则 Memory 兜底。
func NewStore(pool *pgxpool.Pool) Store {
	if pool == nil {
		return NewMemory()
	}
	return NewPostgres(pool)
}

// Open 按 DATABASE_URL 选择实现。url 为空时返回 Memory，不连库。
func Open(ctx context.Context, databaseURL string) (Store, func(), error) {
	databaseURL = strings.TrimSpace(databaseURL)
	if databaseURL == "" {
		return NewMemory(), func() {}, nil
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}
	return NewPostgres(pool), pool.Close, nil
}
