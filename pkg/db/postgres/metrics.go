package postgres

import "github.com/VictoriaMetrics/metrics"

var (
	pgpoolAcquireCount            = metrics.NewCounter("pgpool_acquire_count")
	pgpoolAcquiredConns           = metrics.NewCounter("pgpool_acquired_conns")
	pgpoolCanceledAcquireCount    = metrics.NewCounter("pgpool_canceled_acquire_count")
	pgpoolConstructingConns       = metrics.NewCounter("pgpool_constructing_conns")
	pgpoolEmptyAcquireCount       = metrics.NewCounter("pgpool_empty_acquire_count")
	pgpoolIdleConns               = metrics.NewCounter("pgpool_idle_conns")
	pgpoolMaxConns                = metrics.NewCounter("pgpool_max_conns")
	pgpoolTotalConns              = metrics.NewCounter("pgpool_total_conns")
	pgpoolNewConnsCount           = metrics.NewCounter("pgpool_new_conns_count")
	pgpoolMaxLifetimeDestroyCount = metrics.NewCounter("pgpool_max_lifetime_destroy_count")
	pgpoolMaxIdleDestroyCount     = metrics.NewCounter("pgpool_max_idle_destroy_count")
)
