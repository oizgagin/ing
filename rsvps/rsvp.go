package rsvps

type RSVP struct {
	ID uint64 `json:"rsvp_id"`

	Mtime int64 `json:"mtime"`

	Event struct {
		ID   string `json:"event_id"`
		Time int64  `json:"time"`
		URL  string `json:"event_url"`
	}

	Venue struct {
		Name string  `json:"venue_name"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
	}

	Group struct {
		Name string  `json:"group_name"`
		Lat  float64 `json:"group_lat"`
		Lon  float64 `json:"group_lon"`
	}
}

type Stream interface {
	RSVPS() <-chan RSVP
}
