// Package event 提供 outbox 骨架。S0 不引入 Kafka；worker 仅把 pending 标记为 published。
package event

import (
	"context"

	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/obs"
)

type Store interface {
	Enqueue(ctx context.Context, topic string, payload map[string]any) (*domain.OutboxEvent, error)
	ClaimPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
}

type Publisher struct {
	store Store
}

func New(store Store) *Publisher {
	return &Publisher{store: store}
}

func (p *Publisher) Enqueue(ctx context.Context, topic string, payload map[string]any) error {
	if p == nil || p.store == nil {
		return nil
	}
	_, err := p.store.Enqueue(ctx, topic, payload)
	return err
}

func (p *Publisher) Drain(ctx context.Context, limit int) (int, error) {
	if p == nil || p.store == nil {
		return 0, nil
	}
	items, err := p.store.ClaimPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, ev := range items {
		obs.App().InfoContext(ctx, "outbox.publish",
			"topic", ev.Topic,
			"id", ev.ID.String(),
		)
		if err := p.store.MarkPublished(ctx, ev.ID); err != nil {
			obs.App().ErrorContext(ctx, "outbox.mark_failed", "id", ev.ID.String(), "err", err.Error())
			continue
		}
		n++
	}
	return n, nil
}
