package redisring

import "github.com/VictoriaMetrics/metrics"

var (
	redispoolHits       = metrics.NewCounter("redis_pool_dials_total")
	redispoolMisses     = metrics.NewCounter("redis_pool_misses_total")
	redispoolTimeouts   = metrics.NewCounter("redis_pool_timeouts_total")
	redispoolTotalConns = metrics.NewCounter("redis_pool_total_conns")
	redispoolIdleConns  = metrics.NewCounter("redis_pool_idle_conns")
	redispoolStaleConns = metrics.NewCounter("redis_pool_stale_conns")
)
