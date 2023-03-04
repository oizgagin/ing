package cache

import (
	"context"
	"time"

	"github.com/oizgagin/ing/pkg/rsvps"
)

type EventInfoCache interface {
	Get(ctx context.Context, eventID string) (rsvps.EventInfo, error)
	Set(ctx context.Context, eventID string, info rsvps.EventInfo, ttl time.Duration) error
}
