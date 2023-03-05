package kafka

import "github.com/VictoriaMetrics/metrics"

var (
	kafkaDials           = metrics.NewCounter("kafka_dials_total")
	kafkaFetches         = metrics.NewCounter("kafka_fetches_total")
	kafkaMessages        = metrics.NewCounter("kafka_messages_total")
	kafkaInvalidMessages = metrics.NewCounter("kafka_invalid_messages_total")
	kafkaBytes           = metrics.NewCounter("kafka_bytes_total")
	kafkaRebalances      = metrics.NewCounter("kafka_rebalances_total")
	kafkaTimeouts        = metrics.NewCounter("kafka_timeouts_total")
	kafkaErrors          = metrics.NewCounter("kafka_errors_total")
)
