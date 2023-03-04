package redisring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	configtypes "github.com/oizgagin/ing/pkg/config/types"
	"github.com/oizgagin/ing/pkg/rsvps"
)

type Config struct {
	Addrs []string `toml:"addrs"`
	User  string   `toml:"user"`
	Pass  string   `toml:"pass"`
	DB    int      `toml:"db"`

	TTL configtypes.Duration `toml:"ttl"`

	DialTimeout  configtypes.Duration `toml:"dial_timeout"`
	ReadTimeout  configtypes.Duration `toml:"read_timeout"`
	WriteTimeout configtypes.Duration `toml:"write_timeout"`

	PoolSize     int `toml:"pool_size"`
	MinIdleConns int `toml:"min_idle_conns"`
	MaxIdleConns int `toml:"max_idle_conns"`
}

type Cache struct {
	ring *redis.Ring
	ttl  time.Duration
}

func NewCache(cfg Config) (*Cache, error) {
	addrs := make(map[string]string)
	for i, addr := range cfg.Addrs {
		addrs[fmt.Sprintf("shard%d", i)] = addr
	}

	ring := redis.NewRing(&redis.RingOptions{
		Addrs:    addrs,
		Username: cfg.User,
		Password: cfg.Pass,
		DB:       cfg.DB,

		DialTimeout:  cfg.DialTimeout.Duration,
		ReadTimeout:  cfg.ReadTimeout.Duration,
		WriteTimeout: cfg.WriteTimeout.Duration,

		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxIdleConns: cfg.MaxIdleConns,
	})

	pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Second)
	defer pingCancel()

	if err := ring.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("could not ping redis ring: %w", err)
	}

	return &Cache{ring: ring, ttl: cfg.TTL.Duration}, nil
}

func (c *Cache) Get(ctx context.Context, eventID string) (rsvps.EventInfo, error) {
	j, err := c.ring.Get(ctx, eventID).Bytes()
	if err != nil {
		return rsvps.EventInfo{}, fmt.Errorf("could not get cached event: %w", err)
	}

	var info rsvps.EventInfo
	if err := json.Unmarshal(j, &info); err != nil {
		return rsvps.EventInfo{}, fmt.Errorf("could not unmarshal cached event: %w", err)
	}

	return info, nil
}

func (c *Cache) Set(ctx context.Context, eventID string, info rsvps.EventInfo) error {
	b, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("could not marshal event %v: %w", eventID, err)
	}
	if err := c.ring.Set(ctx, eventID, string(b), c.ttl).Err(); err != nil {
		return fmt.Errorf("could not cache event %v: %w", eventID, err)
	}
	return nil
}
