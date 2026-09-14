package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kxs888/kxs/backend/internal/audit"
	"github.com/kxs888/kxs/backend/internal/config"
	"github.com/kxs888/kxs/backend/internal/event"
	"github.com/kxs888/kxs/backend/internal/handler"
	httpx "github.com/kxs888/kxs/backend/internal/http"
	"github.com/kxs888/kxs/backend/internal/obs"
	"github.com/kxs888/kxs/backend/internal/repo"
	"github.com/kxs888/kxs/backend/internal/scheduler"
	"github.com/kxs888/kxs/backend/internal/service"
	"github.com/kxs888/kxs/backend/internal/stream"
)

type App struct {
	cfg  *config.Config
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	obs.Configure(os.Stdout, obs.ParseLevel(cfg.LogLevel))
	if _, _, err := obs.SetupTracer(ctx, cfg.OTELEndpoint, cfg.ServiceName); err != nil {
		return nil, err
	}
	pool, err := repo.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := repo.Migrate(cfg.DatabaseURL); err != nil {
		pool.Close()
		return nil, err
	}
	return &App{cfg: cfg, pool: pool}, nil
}

func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}

func (a *App) Handler() http.Handler {
	users := repo.NewUserRepo(a.pool)
	pings := repo.NewPingWriteRepo(a.pool)
	audStore := repo.NewAuditRepo(a.pool)
	outboxStore := repo.NewOutboxRepo(a.pool)
	idemp := repo.NewIdempotencyRepo(a.pool)
	tickets := repo.NewStreamRepo(a.pool)

	aud := audit.New(audStore)
	pub := event.New(outboxStore)
	authSvc := service.NewAuth(users, aud, a.cfg.JWTSecret, a.cfg.JWTTTL)
	pingSvc := service.NewPing(pings, aud, pub)
	streamSvc := stream.New(tickets)
	api := handler.New(a.pool, authSvc, pingSvc, streamSvc)
	return httpx.NewRouter(httpx.Deps{
		Cfg:         a.cfg,
		API:         api,
		Idempotency: idemp,
	})
}

func (a *App) EnsureBootstrap(ctx context.Context) error {
	users := repo.NewUserRepo(a.pool)
	authSvc := service.NewAuth(users, audit.New(repo.NewAuditRepo(a.pool)), a.cfg.JWTSecret, a.cfg.JWTTTL)
	return authSvc.EnsureBootstrap(ctx, a.cfg.BootstrapUser, a.cfg.BootstrapPass)
}

func (a *App) Run(ctx context.Context) error {
	if err := a.EnsureBootstrap(ctx); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	var sched *scheduler.Runner
	if a.cfg.RunWorker() {
		pub := event.New(repo.NewOutboxRepo(a.pool))
		sched = scheduler.New(pub, a.cfg.OutboxInterval)
		go sched.Run(ctx)
	}

	if !a.cfg.ServeHTTP() {
		obs.App().Info("role.worker_only")
		<-ctx.Done()
		return nil
	}

	srv := &http.Server{
		Addr:              a.cfg.HTTPAddr,
		Handler:           a.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		obs.App().Info("http.listen", "addr", a.cfg.HTTPAddr, "role", a.cfg.AppRole, "sm_crypto", a.cfg.SMCryptoEnabled)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}
