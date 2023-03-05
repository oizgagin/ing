package cache

import (
	"context"
	"errors"
	"time"

	"github.com/oizgagin/ing/pkg/rsvps"
)

//go:generate mockery --name EventInfoCache
type EventInfoCache interface {
	Get(ctx context.Context, eventID string) (rsvps.EventInfo, error)
	Set(ctx context.Context, eventID string, info rsvps.EventInfo, ttl time.Duration) error
}

var ErrNoCachedEventInfo = errors.New("no cached event info")
