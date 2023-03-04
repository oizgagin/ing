package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	configtypes "github.com/oizgagin/ing/pkg/config/types"
	"github.com/oizgagin/ing/pkg/rsvps"
)

type Config struct {
	Brokers            []string             `toml:"brokers"`
	Topic              string               `toml:"topic"`
	ConsumerGroup      string               `toml:"consumer_group"`
	SessionTimeout     configtypes.Duration `toml:"session_timeout"`
	AutocommitInterval configtypes.Duration `toml:"autocommit_interval"`
}

type Stream struct {
	r *kafka.Reader

	l *zap.Logger

	ch     chan rsvps.RSVP
	chOnce sync.Once

	ctxCancel func()

	stats struct {
		totalMsgs   uint64 // atomic
		invalidMsgs uint64 // atomic
	}
}

func NewStream(cfg Config, logger *zap.Logger) *Stream {
	kafkaLogger := newKafkaLogger(logger)

	ctx, cancel := context.WithCancel(context.Background())

	stream := &Stream{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        cfg.ConsumerGroup,
			Topic:          cfg.Topic,
			SessionTimeout: cfg.SessionTimeout.Duration,
			CommitInterval: cfg.AutocommitInterval.Duration,
			Logger:         kafka.LoggerFunc(kafkaLogger.log),
			ErrorLogger:    kafka.LoggerFunc(kafkaLogger.errorLog),
		}),
		l:         logger.With(zap.String("logger", "kafka_stream")),
		ch:        make(chan rsvps.RSVP),
		ctxCancel: cancel,
	}

	go stream.loop(ctx)
	go stream.metrics(ctx)

	return stream
}

func (stream *Stream) RSVPS() <-chan rsvps.RSVP {
	return stream.ch
}

func (stream *Stream) Close() error {
	err := stream.r.Close()
	stream.ctxCancel()
	stream.chOnce.Do(func() { close(stream.ch) })
	return err
}

func (stream *Stream) TotalMsgs() uint64 {
	return atomic.LoadUint64(&stream.stats.totalMsgs)
}

func (stream *Stream) InvalidMsgs() uint64 {
	return atomic.LoadUint64(&stream.stats.invalidMsgs)
}

func (stream *Stream) loop(ctx context.Context) {
	for {
		m, err := stream.r.ReadMessage(ctx)
		if err != nil {
			return
		}

		atomic.AddUint64(&stream.stats.totalMsgs, 1)

		l := stream.l.With(
			zap.Int("partition", m.Partition),
			zap.Int64("offset", m.Offset),
		)

		l.Debug("received kafka message")

		var rsvp rsvps.RSVP
		if err := json.Unmarshal(m.Value, &rsvp); err != nil {
			atomic.AddUint64(&stream.stats.invalidMsgs, 1)
			l.Error("invalid kafka message", zap.Error(err))
			continue
		}

		select {
		case stream.ch <- rsvp:
		case <-ctx.Done():
			return
		}
	}
}

func (stream *Stream) metrics(ctx context.Context) {
	for {
		stats := stream.r.Stats()

		kafkaDialsTotal.Set(uint64(stats.Dials))
		kafkaFetchesTotal.Set(uint64(stats.Fetches))
		kafkaMessagesTotal.Set(uint64(stats.Messages))
		kafkaBytesTotal.Set(uint64(stats.Bytes))
		kafkaRebalancesTotal.Set(uint64(stats.Rebalances))
		kafkaTimeoutsTotal.Set(uint64(stats.Timeouts))
		kafkaErrorsTotal.Set(uint64(stats.Errors))

		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
	}
}

type kafkaLogger struct{ l *zap.Logger }

func newKafkaLogger(l *zap.Logger) *kafkaLogger {
	return &kafkaLogger{l: l.With(zap.String("logger", "kafka"))}
}

func (l *kafkaLogger) log(msg string, args ...interface{}) {
	l.l.Info(fmt.Sprintf(msg, args...))
}

func (l *kafkaLogger) errorLog(msg string, args ...interface{}) {
	l.l.Error(fmt.Sprintf(msg, args...))
}
