package scheduler

import (
	"context"
	"time"

	"github.com/gotosky/gotosky/internal/logger"
	"github.com/gotosky/gotosky/internal/service"
	"github.com/gotosky/gotosky/internal/store"
)

type Job struct {
	Store  *store.Store
	Scorer *service.Scorer
	Every  time.Duration
}

func (j *Job) Run(ctx context.Context) {
	if j.Every <= 0 {
		j.Every = 30 * time.Minute
	}
	j.tick(ctx)
	t := time.NewTicker(j.Every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			j.tick(ctx)
		}
	}
}

func (j *Job) tick(ctx context.Context) {
	sites, err := j.Store.ListSites(ctx)
	if err != nil {
		logger.L().Error("scheduler sites", "err", err)
		return
	}
	tgts, err := j.Store.ListTargets(ctx, "", "")
	if err != nil {
		return
	}
	if len(tgts) > 36 {
		tgts = tgts[:36]
	}
	for _, s := range sites {
		if err := j.Scorer.Recompute(ctx, s, tgts, 7); err != nil {
			logger.L().Error("scheduler recompute", "site", s.Name, "err", err)
		}
	}
}
