//go:build e2e
// +build e2e

package kafka_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {

	const (
		topic         = "ing_e2e_kafka_test_topic"
		consumerGroup = "ing_e2e_kafka_test_consumergroup"
	)

	brokerAddr := setUp(t, topic)
	fmt.Println(brokerAddr)

}

func setUp(t *testing.T, topic string) string {
	t.Helper()

	brokerAddr := os.Getenv("ING_E2E_KAFKA_BROKER_ADDR")
	assert.NotEmpty(t, brokerAddr)

	retries := 10

	topicCreated := false
	for i := 0; i < retries; i++ {
		err := createTopic(brokerAddr, topic)
		if err != nil {
			time.Sleep(time.Second)
			continue
		} else {
			topicCreated = true
			break
		}
	}

	assert.True(t, topicCreated)

	return brokerAddr
}

func createTopic(brokerAddr, topic string) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}
