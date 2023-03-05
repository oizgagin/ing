//go:build e2e

package redisring_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	cachepkg "github.com/oizgagin/ing/pkg/cache"
	"github.com/oizgagin/ing/pkg/cache/redisring"
	configtypes "github.com/oizgagin/ing/pkg/config/types"
	"github.com/oizgagin/ing/pkg/rsvps"
)

func TestCache_GetSet(t *testing.T) {

	var (
		maxTestDuration = time.Minute
	)

	ctx, cancel := context.WithTimeout(context.Background(), maxTestDuration)
	defer cancel()

	cache, _, tearDown := setUp(t, ctx)
	defer tearDown(t)

	eventID := "event_id1"
	eventTTL := time.Second

	eventInfo := rsvps.EventInfo{
		Venue: rsvps.Venue{
			ID:   1001,
			Name: "venue_name1",
			Lat:  11,
			Lon:  12,
		},
		Group: rsvps.Group{
			ID:      2001,
			Name:    "group_name1",
			Country: "US",
			State:   "AR",
			City:    "group_city1",
			Lat:     51,
			Lon:     52,
			Urlname: "group_urlname1",
			Topics: []rsvps.GroupTopic{
				{Urlkey: "group_urlkey1", TopicName: "group_topicname1"},
				{Urlkey: "group_urlkey2", TopicName: "group_topicname2"},
				{Urlkey: "group_urlkey3", TopicName: "group_topicname3"},
			},
		},
	}

	err := cache.Set(ctx, eventID, eventInfo, eventTTL)
	require.NoError(t, err)

	got1, err := cache.Get(ctx, eventID)
	require.NoError(t, err)
	require.Equal(t, eventInfo, got1)

	time.Sleep(2 * eventTTL)

	got2, err := cache.Get(ctx, eventID)
	require.Equal(t, got2, rsvps.EventInfo{})
	require.Equal(t, err, cachepkg.ErrNoCachedEventInfo)
}

func setUp(t *testing.T, ctx context.Context) (*redisring.Cache, *redis.Ring, func(t *testing.T)) {
	t.Helper()

	redisAddrs := strings.Split(os.Getenv("ING_E2E_REDIS_ADDRS"), ",")
	require.True(t, len(redisAddrs) > 1)
	for _, addr := range redisAddrs {
		require.NotEmpty(t, addr)
	}

	redisUser := os.Getenv("ING_E2E_REDIS_USER")
	require.NotEmpty(t, redisUser)

	redisPass := os.Getenv("ING_E2E_REDIS_PASS")
	require.NotEmpty(t, redisPass)

	cfg := redisring.Config{
		Addrs: redisAddrs,
		User:  redisUser,
		Pass:  redisPass,
		DB:    0,

		DialTimeout:  configtypes.Duration{Duration: time.Second},
		ReadTimeout:  configtypes.Duration{Duration: time.Second},
		WriteTimeout: configtypes.Duration{Duration: time.Second},

		PoolSize:     2,
		MinIdleConns: 1,
		MaxIdleConns: 1,
	}

	cache, err := redisring.NewCache(cfg)
	require.NoError(t, err)

	ring := redisring.NewRing(cfg)

	err = ring.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		return shard.FlushAll(ctx).Err()
	})
	require.NoError(t, err)

	return cache, ring, func(t *testing.T) {
		require.NoError(t, cache.Close())
	}
}
