package scheduler

import (
	"context"
	"time"

	"github.com/kxs888/kxs/backend/internal/event"
	"github.com/kxs888/kxs/backend/internal/obs"
)

type Runner struct {
	pub      *event.Publisher
	interval time.Duration
}

func New(pub *event.Publisher, interval time.Duration) *Runner {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &Runner{pub: pub, interval: interval}
}

func (r *Runner) Run(ctx context.Context) {
	if r == nil || r.pub == nil {
		return
	}
	t := time.NewTicker(r.interval)
	defer t.Stop()
	obs.App().InfoContext(ctx, "scheduler.started", "interval", r.interval.String())
	for {
		select {
		case <-ctx.Done():
			obs.App().InfoContext(ctx, "scheduler.stopped")
			return
		case <-t.C:
			n, err := r.pub.Drain(ctx, 50)
			if err != nil {
				obs.App().ErrorContext(ctx, "scheduler.drain", "err", err.Error())
				continue
			}
			if n > 0 {
				obs.App().InfoContext(ctx, "scheduler.drained", "count", n)
			}
		}
	}
}
