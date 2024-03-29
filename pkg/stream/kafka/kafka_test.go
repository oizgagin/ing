//go:build e2e

package kafka_test

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	segmentiokafka "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	configtypes "github.com/oizgagin/ing/pkg/config/types"
	"github.com/oizgagin/ing/pkg/rsvps"
	"github.com/oizgagin/ing/pkg/stream/kafka"
)

func TestStream(t *testing.T) {
	const (
		topic         = "ing_e2e_kafka_test_topic"
		consumerGroup = "ing_e2e_kafka_test_consumergroup"
	)

	var (
		maxTestDuration = time.Minute
	)

	ctx, cancel := context.WithTimeout(context.Background(), maxTestDuration)
	defer cancel()

	brokerAddr, producer, logger, rawRsvps, tearDown := setUp(t, ctx, topic)
	defer tearDown(t)

	stream := kafka.NewStream(kafka.Config{
		Brokers:            []string{brokerAddr},
		Topic:              topic,
		ConsumerGroup:      consumerGroup,
		SessionTimeout:     configtypes.Duration{Duration: time.Minute},
		AutocommitInterval: configtypes.Duration{Duration: time.Second},
	}, logger)
	defer stream.Close()

	err := producer.produceMsgs(ctx, rawRsvps)
	require.NoError(t, err)

	gotRsvps := readStream(ctx, stream, 5*time.Second) // TODO: get rid of hardcoded 5s window, wait for all messages explicitly

	wantRsvps := parseRsvps(rawRsvps)

	cmpRsvp := func(rsvp1, rsvp2 rsvps.RSVP) bool {
		return rsvp1.ID < rsvp2.ID
	}

	sort.Slice(gotRsvps, func(i, j int) bool { return cmpRsvp(gotRsvps[i], gotRsvps[j]) })
	sort.Slice(wantRsvps, func(i, j int) bool { return cmpRsvp(wantRsvps[i], wantRsvps[j]) })

	require.Equal(t, wantRsvps, gotRsvps)
}

func setUp(t *testing.T, ctx context.Context, topic string) (string, *producer, *zap.Logger, []string, func(t *testing.T)) {
	t.Helper()

	brokerAddr := os.Getenv("ING_E2E_KAFKA_BROKER_ADDR")
	require.NotEmpty(t, brokerAddr)

	err := createTopic(ctx, brokerAddr, topic, 10)
	require.NoError(t, err)

	producer := newProducer(brokerAddr, topic)

	rawRsvps, err := getTestRsvps("testdata/meetups.json.gz")
	require.NoError(t, err)
	require.NotEmpty(t, rawRsvps)

	logger := zaptest.NewLogger(t)

	return brokerAddr, producer, logger, rawRsvps, func(t *testing.T) {
		assert.NoError(t, logger.Sync())
		assert.NoError(t, producer.close())
	}
}

func createTopic(ctx context.Context, brokerAddr, topic string, maxRetries int) error {

	create := func() error {
		conn, err := segmentiokafka.DialContext(ctx, "tcp", brokerAddr)
		if err != nil {
			return err
		}
		defer conn.Close()

		return conn.CreateTopics(segmentiokafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	var err error
	for i := 0; i < maxRetries; i++ {
		err = create()
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("could not create topic %v at %v: timeout", topic, brokerAddr)
		case <-time.After(time.Second):
		}
	}

	return fmt.Errorf("could not create topic %v at %v: %v", topic, brokerAddr, err)
}

func getTestRsvps(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open %v: %v", filename, err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("invalid gzip file %v: %v", filename, err)
	}
	defer gzr.Close()

	content, err := io.ReadAll(gzr)
	if err != nil {
		return nil, fmt.Errorf("could not read file %v: %v", filename, err)
	}

	return strings.Split(string(content), "\n"), nil
}

func parseRsvps(raws []string) (valid []rsvps.RSVP) {
	for _, raw := range raws {
		var rsvp rsvps.RSVP
		if err := json.Unmarshal([]byte(raw), &rsvp); err == nil {
			valid = append(valid, rsvp)
		}
	}
	return
}

// readStream waits till the first message appears at stream and returns all messages
// received during the `duration` time after first message
func readStream(ctx context.Context, stream *kafka.Stream, duration time.Duration) (rsvps []rsvps.RSVP) {
	var (
		readCh   = make(chan struct{})
		readOnce sync.Once
	)

	go func() {
		for rsvp := range stream.RSVPS() {
			readOnce.Do(func() { close(readCh) })
			rsvps = append(rsvps, rsvp)
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-readCh:
	}

	select {
	case <-ctx.Done():
	case <-time.After(duration):
	}

	return rsvps
}

type producer struct {
	w *segmentiokafka.Writer
}

func newProducer(brokerAddr, topic string) *producer {
	return &producer{
		w: &segmentiokafka.Writer{
			Addr:  segmentiokafka.TCP(brokerAddr),
			Topic: topic,
			Async: true,
		},
	}
}

func (p *producer) close() error {
	return p.w.Close()
}

func (p *producer) produceMsgs(ctx context.Context, msgs []string) error {
	for _, msg := range msgs {
		if err := p.w.WriteMessages(ctx, segmentiokafka.Message{Value: []byte(msg)}); err != nil {
			return fmt.Errorf("could not produce kafka message: %v", err)
		}
	}
	return nil
}
