package kafka

import "github.com/VictoriaMetrics/metrics"

var (
	kafkaDialsTotal      = metrics.NewCounter("kafka_dials_total")
	kafkaFetchesTotal    = metrics.NewCounter("kafka_fetches_total")
	kafkaMessagesTotal   = metrics.NewCounter("kafka_messages_total")
	kafkaBytesTotal      = metrics.NewCounter("kafka_bytes_total")
	kafkaRebalancesTotal = metrics.NewCounter("kafka_rebalances_total")
	kafkaTimeoutsTotal   = metrics.NewCounter("kafka_timeouts_total")
	kafkaErrorsTotal     = metrics.NewCounter("kafka_errors_total")
)
