package rsvps

type RSVP struct {
	ID         int64  `json:"rsvp_id"`
	Mtime      int64  `json:"mtime"`
	Guests     uint   `json:"guests"`
	Visibility string `json:"visibility"`
	Response   string `json:"response"`

	Venue  Venue  `json:"venue"`
	Member Member `json:"member"`
	Event  Event  `json:"event"`
	Group  Group  `json:"group"`
}

type Venue struct {
	ID   int64   `json:"venue_id"`
	Name string  `json:"venue_name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

type Member struct {
	ID    int64  `json:"member_id"`
	Name  string `json:"member_name"`
	Photo string `json:"photo"`
}

type Event struct {
	ID   string `json:"event_id"`
	Name string `json:"event_name"`
	URL  string `json:"event_url"`
	Time int64  `json:"time"`
}

type Group struct {
	ID      int64        `json:"group_id"`
	Name    string       `json:"group_name"`
	Country string       `json:"group_country"`
	State   string       `json:"state"`
	City    string       `json:"city"`
	Lat     float64      `json:"group_lat"`
	Lon     float64      `json:"group_lon"`
	Urlname string       `json:"group_urlname"`
	Topics  []GroupTopic `json:"topics"`
}

type GroupTopic struct {
	Urlkey    string `json:"urlkey"`
	TopicName string `json:"topic_name"`
}
