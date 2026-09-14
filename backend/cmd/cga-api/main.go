package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kxs888/kxs/backend/internal/app"
	"github.com/kxs888/kxs/backend/internal/config"
	"github.com/kxs888/kxs/backend/internal/obs"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err.Error())
		os.Exit(1)
	}
	ctx := context.Background()
	a, err := app.New(ctx, cfg)
	if err != nil {
		obs.App().Error("startup", "err", err.Error())
		os.Exit(1)
	}
	defer a.Close()
	if err := a.Run(ctx); err != nil {
		obs.App().Error("run", "err", err.Error())
		os.Exit(1)
	}
}
