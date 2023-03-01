package config

import (
	"os"
	"testing"
	"time"

	configtypes "github.com/oizgagin/ing/config/types"
	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {

	raw := `
		[kafka]
		brokers = [ "broker1:9092", "broker2:9092", "broker3:9092" ]
		topic = "rsvps"
		consumer_group = "rsvps_consumer_group"
		autocommit_interval = "30s"
		session_timeout = "30s"

		[redis]
		addrs = [ "redis1:7000", "redis2:7000", "redis3:7000" ]
		user = "rsvps_redis_user"
		pass = "{% ENV:REDIS_PASS %}"

		[postgres]
		addr = "postgres1:5432"
		user = "rsvps_postgres_user"
		pass = "{% ENV:POSTGRES_PASS %}"
		dbname = "rsvps"
	`

	t.Setenv("REDIS_PASS", "rsvps_redis_pass")
	t.Setenv("POSTGRES_PASS", "rsvps_postgres_pass")

	filename, tearDown := setUp(t, raw)
	defer tearDown()

	got, err := ParseFile(filename)
	assert.NoError(t, err)

	want := Config{}
	want.Kafka.Brokers = []string{"broker1:9092", "broker2:9092", "broker3:9092"}
	want.Kafka.Topic = "rsvps"
	want.Kafka.ConsumerGroup = "rsvps_consumer_group"
	want.Kafka.AutocommitInterval = configtypes.Duration{Duration: 30 * time.Second}
	want.Kafka.SessionTimeout = configtypes.Duration{Duration: 30 * time.Second}
	want.Redis.Addrs = []string{"redis1:7000", "redis2:7000", "redis3:7000"}
	want.Redis.User = "rsvps_redis_user"
	want.Redis.Pass = "rsvps_redis_pass"
	want.Postgres.Addr = "postgres1:5432"
	want.Postgres.User = "rsvps_postgres_user"
	want.Postgres.Pass = "rsvps_postgres_pass"
	want.Postgres.DBName = "rsvps"

	assert.Equal(t, want, got)
}

func setUp(t *testing.T, raw string) (string, func()) {
	t.Helper()

	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	if _, err := f.WriteString(raw); err != nil {
		t.Fatalf("could not write test config to temp file %v: %v", f.Name(), err)
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("could not flush test config file %v: %v", f.Name(), err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("could not close test config file %v: %v", f.Name(), err)
	}

	return f.Name(), func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Fatalf("could not cleanup test config file %v: %v", f.Name(), err)
		}
	}
}
