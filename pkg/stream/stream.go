package stream

import "github.com/oizgagin/ing/pkg/rsvps"

type Stream interface {
	RSVPS() <-chan rsvps.RSVP
	Close() error
}
