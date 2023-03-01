package rsvps

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	configtypes "github.com/oizgagin/ing/config/types"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	Brokers            []string             `toml:"brokers"`
	Topic              string               `toml:"topic"`
	ConsumerGroup      string               `toml:"consumer_group"`
	SessionTimeout     configtypes.Duration `toml:"session_timeout"`
	AutocommitInterval configtypes.Duration `toml:"autocommit_interval"`
}

type KafkaStream struct {
	r         *kafka.Reader
	l         *zap.Logger
	ch        chan RSVP
	ctxCancel func()

	stats struct {
		totalMsgs   uint64 // atomic
		invalidMsgs uint64 // atomic
	}
}

func NewKafkaStream(cfg KafkaConfig, logger *zap.Logger) *KafkaStream {
	kafkaLogger := newKafkaLogger(logger)

	ctx, cancel := context.WithCancel(context.Background())

	stream := &KafkaStream{
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
		ch:        make(chan RSVP),
		ctxCancel: cancel,
	}

	go stream.loop(ctx)

	return stream
}

func (stream *KafkaStream) RSVPS() <-chan RSVP {
	return stream.ch
}

func (stream *KafkaStream) Close() error {
	err := stream.r.Close()
	stream.ctxCancel()
	close(stream.ch)
	return err
}

func (stream *KafkaStream) TotalMsgs() uint64 {
	return atomic.LoadUint64(&stream.stats.totalMsgs)
}

func (stream *KafkaStream) InvalidMsgs() uint64 {
	return atomic.LoadUint64(&stream.stats.invalidMsgs)
}

func (stream *KafkaStream) loop(ctx context.Context) {
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

		var rsvp RSVP
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
