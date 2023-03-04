package db

import (
	"context"
	"time"

	"github.com/oizgagin/ing/pkg/rsvps"
)

type DB interface {
	SaveRSVP(ctx context.Context, rsvp rsvps.RSVP) error
	TopkEvents(ctx context.Context, date time.Time, k uint) ([]TopkEvent, error)
	GetEventInfo(ctx context.Context, eventID string) (rsvps.EventInfo, error)
}

type TopkEvent struct {
	Event          rsvps.Event
	ConfirmedRSVPs int
}
