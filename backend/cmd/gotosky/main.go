package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/config"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/handler"
	"github.com/gotosky/gotosky/internal/logger"
	"github.com/gotosky/gotosky/internal/scheduler"
	"github.com/gotosky/gotosky/internal/seed"
	"github.com/gotosky/gotosky/internal/service"
	"github.com/gotosky/gotosky/internal/store"
	"github.com/gotosky/gotosky/internal/telescope"
	"github.com/gotosky/gotosky/internal/weather"
	"github.com/gotosky/gotosky/internal/ws"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.LogLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.L().Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := store.Migrate(ctx, db); err != nil {
		logger.L().Error("migrate", "err", err)
		os.Exit(1)
	}
	st := &store.Store{DB: db}

	var inner weather.Provider
	if cfg.WeatherIsMock() {
		inner = weather.NewMock("MIXED")
	} else {
		inner = weather.NewOpenMeteo()
	}
	guard := weather.NewGuard(inner, cfg.WeatherQuota, func(p, ep string, _ any, lat time.Duration, code int, hit bool) {
		st.LogAPI(context.Background(), p, ep, lat, code, hit)
	})

	if cfg.SeedDemo {
		if err := seed.Run(ctx, st); err != nil {
			logger.L().Error("seed", "err", err)
			os.Exit(1)
		}
	}

	scorer := service.NewScorer(st, guard)
	hub := ws.NewHub(cfg.CORSOrigins, cfg.PublicHost)

	newDrv := func() telescope.Driver {
		if cfg.TelescopeIsMock() {
			return telescope.NewMockDriver()
		}
		return telescope.NewAlpaca(cfg.AlpacaBaseURL)
	}
	reg := telescope.NewRegistry(func(s domain.Session) *telescope.SessionActor {
		return telescope.NewActor(s, newDrv(), st, hub)
	})

	// Recovery scan: mark interrupted, do not silently drop.
	if incs, err := st.IncompleteSessions(ctx); err == nil {
		for _, s := range incs {
			from := s.State
			s.State = "ERROR"
			s.LastError = "INTERRUPTED"
			now := time.Now()
			s.EndedAt = &now
			_ = st.SaveSession(ctx, s)
			_ = st.AppendEvent(ctx, domain.SessionEvent{
				ID: uuid.New(), SessionID: s.ID, FromState: from, ToState: "ERROR", Class: "TRANSIENT",
				Context: []byte(`{"reason":"INTERRUPTED"}`),
			})
		}
	}

	api := &handler.API{Cfg: cfg, Store: st, Scorer: scorer, Weather: guard, Guard: guard, Hub: hub, Reg: reg, NewDrv: newDrv}

	go (&scheduler.Job{Store: st, Scorer: scorer, Every: 30 * time.Minute}).Run(ctx)
	go func() {
		time.Sleep(2 * time.Second)
		sites, _ := st.ListSites(ctx)
		tgts, _ := st.ListTargets(ctx, "", "")
		if len(tgts) > 24 {
			tgts = tgts[:24]
		}
		for _, s := range sites {
			_ = scorer.Recompute(ctx, s, tgts, 7)
		}
	}()

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: api.Routes(), ReadHeaderTimeout: 8 * time.Second}
	go func() {
		logger.L().Info("listen", "addr", cfg.HTTPAddr, "weather", cfg.WeatherProvider, "driver", cfg.TelescopeDriver)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Error("http", "err", err)
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	shctx, c2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer c2()
	_ = srv.Shutdown(shctx)
}
