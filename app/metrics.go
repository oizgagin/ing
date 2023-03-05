package app

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/oizgagin/ing/app/utils"
	"go.uber.org/zap"
)

type metricsServer struct {
	l   *zap.Logger
	srv *http.Server
}

func newMetricsServer(l *zap.Logger, addr string) *metricsServer {
	server := metricsServer{
		srv: &http.Server{Addr: addr},
	}

	server.srv.Handler = &server

	go func() {
		if err := server.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("metrics serve error", zap.Error(err))
			utils.StopApp()
		}
	}()

	return &server
}

func (s *metricsServer) Close() error {
	return s.srv.Close()
}

func (s *metricsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/metrics" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	metrics.WritePrometheus(w, true)
}
