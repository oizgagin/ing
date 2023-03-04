//go:build e2e

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
	"github.com/stretchr/testify/require"

	dbpkg "github.com/oizgagin/ing/pkg/db"
	"github.com/oizgagin/ing/pkg/db/postgres"
	"github.com/oizgagin/ing/pkg/rsvps"
)

func TestDB_SaveRSVP(t *testing.T) {

	var (
		maxTestDuration = time.Minute
	)

	ctx, cancel := context.WithTimeout(context.Background(), maxTestDuration)
	defer cancel()

	db, conn, tearDown := setUp(t, ctx)
	defer tearDown()

	rsvp := rsvps.RSVP{
		ID:         1001,
		Mtime:      1002,
		Guests:     1,
		Visibility: "public",
		Response:   "yes",
		Venue:      rsvps.Venue{ID: 2001, Name: "venue_name1", Lat: 21, Lon: 22},
		Member:     rsvps.Member{ID: 3001, Name: "member_name1", Photo: "member_photo1"},
		Event:      rsvps.Event{ID: "event_id1", Name: "event_name1", URL: "event_url1", Time: 4001},
		Group: rsvps.Group{
			ID:      5001,
			Name:    "group_name1",
			Country: "US",
			State:   "",
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

	err := db.SaveRSVP(ctx, rsvp)
	require.NoError(t, err)

	require.Equal(t, rsvp.Venue, selectVenue(t, ctx, conn, rsvp.Venue.ID))
	require.Equal(t, rsvp.Member, selectMember(t, ctx, conn, rsvp.Member.ID))
	require.Equal(t, rsvp.Event, selectEvent(t, ctx, conn, rsvp.Event.ID))
	require.Equal(t, rsvp.Group, selectGroup(t, ctx, conn, rsvp.Group.ID))

	wantRsvp := dbRsvp{
		ID:         rsvp.ID,
		Mtime:      rsvp.Mtime,
		Guests:     rsvp.Guests,
		Response:   rsvp.Response,
		Visibility: rsvp.Visibility,
		EventID:    rsvp.Event.ID,
		VenueID:    rsvp.Venue.ID,
		GroupID:    rsvp.Group.ID,
		MemberID:   rsvp.Member.ID,
	}
	require.Equal(t, wantRsvp, selectRsvp(t, ctx, conn, rsvp.ID))
}

func TestDB_TopkEvents(t *testing.T) {

	var (
		maxTestDuration = time.Minute
	)

	ctx, cancel := context.WithTimeout(context.Background(), maxTestDuration)
	defer cancel()

	db, _, tearDown := setUp(t, ctx)
	defer tearDown()

	rsvp := rsvps.RSVP{
		Guests:     1,
		Visibility: "public",
		Response:   "yes",
		Venue:      rsvps.Venue{ID: 2001, Name: "venue_name1", Lat: 21, Lon: 22},
		Member:     rsvps.Member{ID: 3001, Name: "member_name1", Photo: "member_photo1"},
		Group: rsvps.Group{
			ID:      5001,
			Name:    "group_name1",
			Country: "US",
			State:   "",
			City:    "group_city1",
			Lat:     51,
			Lon:     52,
			Urlname: "group_urlname1",
			Topics: []rsvps.GroupTopic{
				{Urlkey: "group_urlkey1", TopicName: "group_topicname1"},
			},
		},
	}

	eventCounters := map[string]uint{
		"event_id1": 10,
		"event_id2": 20,
		"event_id3": 21,
		"event_id4": 30,
		"event_id5": 31,
		"event_id6": 40,
		"event_id7": 41,
	}

	day1 := time.Date(2023, 3, 4, 12, 30, 50, 0, time.UTC)
	day2 := time.Date(2023, 3, 5, 12, 30, 50, 0, time.UTC)

	rsvpDates := map[string]time.Time{
		"event_id1": day1,
		"event_id2": day2,
		"event_id3": day1,
		"event_id4": day2,
		"event_id5": day1,
		"event_id6": day2,
		"event_id7": day1,
	}

	for i := 1; i <= 7; i++ {
		eventID := fmt.Sprintf("event_id%d", i)

		for j := 0; j < int(eventCounters[eventID]); j++ {
			currRsvp := rsvp

			currRsvp.ID = int64(i*1e7 + j)
			currRsvp.Mtime = rsvpDates[eventID].UnixMilli()
			currRsvp.Event = rsvps.Event{
				ID:   eventID,
				Name: fmt.Sprintf("event_name%d", i),
				URL:  fmt.Sprintf("event_url%d", i),
				Time: rsvpDates[eventID].UnixMilli(),
			}

			err := db.SaveRSVP(ctx, currRsvp)
			require.NoError(t, err)
		}
	}

	topks1, err := db.TopkEvents(ctx, day1, 2)
	require.NoError(t, err)
	require.Equal(t, []dbpkg.TopkEvent{
		{
			Event: rsvps.Event{ID: "event_id7", Name: "event_name7", URL: "event_url7", Time: day1.UnixMilli()},
			RSVPs: eventCounters["event_id7"],
		},
		{
			Event: rsvps.Event{ID: "event_id5", Name: "event_name5", URL: "event_url5", Time: day1.UnixMilli()},
			RSVPs: eventCounters["event_id5"],
		},
	}, topks1)

	topks2, err := db.TopkEvents(ctx, day2, 2)
	require.NoError(t, err)
	require.Equal(t, []dbpkg.TopkEvent{
		{
			Event: rsvps.Event{ID: "event_id6", Name: "event_name6", URL: "event_url6", Time: day2.UnixMilli()},
			RSVPs: eventCounters["event_id6"],
		},
		{
			Event: rsvps.Event{ID: "event_id4", Name: "event_name4", URL: "event_url4", Time: day2.UnixMilli()},
			RSVPs: eventCounters["event_id4"],
		},
	}, topks2)
}

func setUp(t *testing.T, ctx context.Context) (*postgres.DB, *pgx.Conn, func()) {
	t.Helper()

	postgresAddr := os.Getenv("ING_E2E_POSTGRES_ADDR")
	require.NotEmpty(t, postgresAddr)

	postgresUser := os.Getenv("ING_E2E_POSTGRES_USER")
	require.NotEmpty(t, postgresUser)

	postgresPass := os.Getenv("ING_E2E_POSTGRES_PASS")
	require.NotEmpty(t, postgresPass)

	postgresDB := os.Getenv("ING_E2E_POSTGRES_DB")
	require.NotEmpty(t, postgresDB)

	cfg := postgres.Config{
		Addr:   postgresAddr,
		User:   postgresUser,
		Pass:   postgresPass,
		DBName: postgresDB,
	}

	db, err := postgres.NewDB(cfg)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, cfg.URL())
	require.NoError(t, err)

	return db, conn, func() {
		db.Close()
		conn.Close(ctx)
	}
}

func selectVenue(t *testing.T, ctx context.Context, conn *pgx.Conn, venueID int64) (venue rsvps.Venue) {
	err := conn.QueryRow(ctx, `
		SELECT
			id, name, lat, lon
		FROM venues
			WHERE id = $1
	`, venueID).Scan(&venue.ID, &venue.Name, &venue.Lat, &venue.Lon)
	require.NoError(t, err)
	return
}

func selectMember(t *testing.T, ctx context.Context, conn *pgx.Conn, memberID int64) (member rsvps.Member) {
	err := conn.QueryRow(ctx, `
		SELECT
			id, name, photo
		FROM members
			WHERE id = $1
	`, memberID).Scan(&member.ID, &member.Name, &member.Photo)
	require.NoError(t, err)
	return
}

func selectEvent(t *testing.T, ctx context.Context, conn *pgx.Conn, eventID string) (event rsvps.Event) {
	var eventTime time.Time
	err := conn.QueryRow(ctx, `
		SELECT
			id, name, time, url
		FROM events
			WHERE id = $1
	`, eventID).Scan(&event.ID, &event.Name, &eventTime, &event.URL)
	require.NoError(t, err)

	event.Time = eventTime.UnixMilli()
	return
}

func selectGroup(t *testing.T, ctx context.Context, conn *pgx.Conn, groupID int64) (group rsvps.Group) {
	var groupState zeronull.Text

	err := conn.QueryRow(ctx, `
		SELECT
			id, country, state, city, name, lat, lon, urlname, topics
		FROM groups
			WHERE id = $1
	`, groupID).Scan(
		&group.ID,
		&group.Country,
		&groupState,
		&group.City,
		&group.Name,
		&group.Lat,
		&group.Lon,
		&group.Urlname,
		&group.Topics,
	)

	require.NoError(t, err)
	group.State = string(groupState)
	return
}

type dbRsvp struct {
	ID         int64
	Mtime      int64
	Guests     uint
	Response   string
	Visibility string
	EventID    string
	VenueID    int64
	GroupID    int64
	MemberID   int64
}

func selectRsvp(t *testing.T, ctx context.Context, conn *pgx.Conn, rsvpID int64) (rsvp dbRsvp) {
	var (
		rsvpResponse bool
		rsvpMtime    time.Time
	)

	err := conn.QueryRow(ctx, `
		SELECT
			id, mtime, guests, response, visibility, event_id, venue_id, group_id, member_id
		FROM rsvps
			WHERE id = $1
	`, rsvpID).Scan(
		&rsvp.ID,
		&rsvpMtime,
		&rsvp.Guests,
		&rsvpResponse,
		&rsvp.Visibility,
		&rsvp.EventID,
		&rsvp.VenueID,
		&rsvp.GroupID,
		&rsvp.MemberID,
	)

	require.NoError(t, err)

	rsvp.Mtime = rsvpMtime.UnixMilli()
	if rsvpResponse {
		rsvp.Response = "yes"
	} else {
		rsvp.Response = "no"
	}

	return
}

func TestConfig(t *testing.T) {
	c := postgres.Config{
		Addr:   "localhost:5432",
		User:   "user",
		Pass:   "pass",
		DBName: "dbname",
	}

	require.Equal(t, "postgres://user:pass@localhost:5432/dbname", c.URL())
}
