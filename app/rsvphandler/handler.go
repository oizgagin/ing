package rsvphandler

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	configtypes "github.com/oizgagin/ing/pkg/config/types"
	"github.com/oizgagin/ing/pkg/db"
	"github.com/oizgagin/ing/pkg/rsvps"
	"github.com/oizgagin/ing/pkg/stream"
)

type Config struct {
	Workers     int                  `toml:"workers"`
	SaveTimeout configtypes.Duration `toml:"save_timeout"`
}

type Handler struct {
	l *zap.Logger

	ctxCancel func()
	wg        *sync.WaitGroup

	stream stream.Stream
	db     db.DB

	saveTimeout time.Duration
}

func NewHandler(cfg Config, l *zap.Logger, stream stream.Stream, db db.DB) *Handler {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	handler := &Handler{
		l:           l,
		ctxCancel:   cancel,
		wg:          wg,
		stream:      stream,
		db:          db,
		saveTimeout: cfg.SaveTimeout.Duration,
	}

	wg.Add(cfg.Workers)
	for i := 0; i < cfg.Workers; i++ {
		go handler.loop(ctx)
	}

	return handler
}

func (h *Handler) Stop() {
	h.ctxCancel()
	h.wg.Wait()
}

func (h *Handler) loop(ctx context.Context) {
	defer h.wg.Done()

	for {
		select {
		case rsvp := <-h.stream.RSVPS():
			if err := h.saveRsvp(ctx, rsvp); err != nil {
				h.l.Error("rsvp save error", zap.Int64("rsvp_id", rsvp.ID), zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}

func (h *Handler) saveRsvp(ctx context.Context, rsvp rsvps.RSVP) error {
	saveCtx, saveCancel := context.WithTimeout(ctx, h.saveTimeout)
	defer saveCancel()

	return h.db.SaveRSVP(saveCtx, rsvp)
}
