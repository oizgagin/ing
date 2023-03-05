package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/oizgagin/ing/app/utils"
	"github.com/oizgagin/ing/pkg/cache"
	configtypes "github.com/oizgagin/ing/pkg/config/types"
	"github.com/oizgagin/ing/pkg/db"
)

type Config struct {
	Addr            string               `toml:"addr"`
	ReadTimeout     configtypes.Duration `toml:"read_timeout"`
	WriteTimeout    configtypes.Duration `toml:"write_timeout"`
	ShutdownTimeout configtypes.Duration `toml:"shutdown_timeout"`
	CacheTTL        configtypes.Duration `toml:"cache_ttl"`
	CacheSetTimeout configtypes.Duration `toml:"cache_set_timeout"`
}

type Server struct {
	l          *zap.Logger
	db         db.DB
	eventCache cache.EventInfoCache

	ln              net.Listener
	srv             *http.Server
	shutdownTimeout time.Duration
	cacheTTL        time.Duration
	cacheSetTimeout time.Duration
}

func NewServer(cfg Config, l *zap.Logger, db db.DB, eventCache cache.EventInfoCache) (*Server, error) {
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("could not start listening on %v: %w", cfg.Addr, err)
	}

	srv := &http.Server{
		ReadTimeout:  cfg.ReadTimeout.Duration,
		WriteTimeout: cfg.WriteTimeout.Duration,
	}

	server := Server{
		l:               l,
		db:              db,
		eventCache:      eventCache,
		srv:             srv,
		shutdownTimeout: cfg.ShutdownTimeout.Duration,
		cacheTTL:        cfg.CacheTTL.Duration,
		cacheSetTimeout: cfg.CacheSetTimeout.Duration,
	}

	srv.Handler = &server

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			l.Error("http serve error", zap.Error(err))
			utils.StopApp()
		}
	}()

	return &server, nil
}

func (s *Server) Close() error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer shutdownCancel()

	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("could not shutdown server: %w", err)
	}

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/api/v1/events/topk":
		s.handleEventsTopk(w, r)
	case "/api/v1/events/info":
		s.handleEventsInfo(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (s *Server) handleEventsTopk(w http.ResponseWriter, r *http.Request) {
	l := s.l.With(zap.String("handler", "handleEventsTopk"))

	date, err := time.Parse("2006-01-02", r.URL.Query()["date"][0])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	k, err := strconv.ParseUint(r.URL.Query()["k"][0], 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	l = l.With(zap.Time("date", date), zap.Uint("k", uint(k)))

	topk, err := s.db.TopkEvents(r.Context(), date, uint(k))
	if err != nil {
		l.Error("could not get topk events", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(topk); err != nil {
		l.Error("could not write response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleEventsInfo(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query()["event_id"][0]

	l := s.l.With(zap.String("handler", "handleEventsInfo"), zap.String("event_id", eventID))

	info, err := s.eventCache.Get(r.Context(), eventID)
	if err != nil && err != cache.ErrNoCachedEventInfo {
		l.Error("could not get cached event info", zap.String("event_id", eventID), zap.Error(err))
		return
	}

	if err == cache.ErrNoCachedEventInfo {
		info, err = s.db.GetEventInfo(r.Context(), eventID)
		if err != nil && err != db.ErrNoEvents {
			l.Error("could not get event info from db", zap.String("event_id", eventID), zap.Error(err))
			return
		}
		if err == db.ErrNoEvents {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// TODO: maybe set this in goroutine
		setCtx, setCancel := context.WithTimeout(context.Background(), s.cacheSetTimeout)
		defer setCancel()

		if err := s.eventCache.Set(setCtx, eventID, info, s.cacheTTL); err != nil {
			l.Error("could not cache event info", zap.Error(err))
		}
	}

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(info); err != nil {
		l.Error("could not write response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}
