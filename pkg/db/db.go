package db

import (
	"context"
	"time"

	"github.com/oizgagin/ing/pkg/rsvps"
)

type DB interface {
	SaveRSVP(ctx context.Context, rsvp rsvps.RSVP) error
	TopkEvents(ctx context.Context, date time.Time, k uint) ([]rsvps.Event, error)
	GetEventInfo(ctx context.Context, eventID string) (EventInfo, error)
}

type EventInfo struct {
	Group rsvps.Group
	Venue rsvps.Venue
}
