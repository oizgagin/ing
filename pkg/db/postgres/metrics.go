package postgres

import "github.com/VictoriaMetrics/metrics"

var (
	pgpoolAcquireCount            = metrics.NewCounter("pg_pool_acquire_total")
	pgpoolAcquiredConns           = metrics.NewCounter("pg_pool_acquired_conns")
	pgpoolCanceledAcquireCount    = metrics.NewCounter("pg_pool_canceled_acquire_total")
	pgpoolConstructingConns       = metrics.NewCounter("pg_pool_constructing_conns")
	pgpoolEmptyAcquireCount       = metrics.NewCounter("pg_pool_empty_acquire_total")
	pgpoolIdleConns               = metrics.NewCounter("pg_pool_idle_conns")
	pgpoolMaxConns                = metrics.NewCounter("pg_pool_max_conns")
	pgpoolTotalConns              = metrics.NewCounter("pg_pool_total_conns")
	pgpoolNewConnsCount           = metrics.NewCounter("pg_pool_new_conns_total")
	pgpoolMaxLifetimeDestroyCount = metrics.NewCounter("pg_pool_max_lifetime_destroy_total")
	pgpoolMaxIdleDestroyCount     = metrics.NewCounter("pg_pool_max_idle_destroy_total")
)
