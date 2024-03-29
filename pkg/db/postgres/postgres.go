package postgres

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
	"github.com/jackc/pgx/v5/pgxpool"

	dbpkg "github.com/oizgagin/ing/pkg/db"
	"github.com/oizgagin/ing/pkg/rsvps"
)

type Config struct {
	Addr   string `toml:"addr"`
	User   string `toml:"user"`
	Pass   string `toml:"pass"`
	DBName string `toml:"dbname"`
}

func (c Config) URL() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Pass),
		Host:   c.Addr,
		Path:   c.DBName,
	}
	return u.String()
}

type DB struct {
	pool      *pgxpool.Pool
	ctxCancel func()
}

func NewDB(cfg Config) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("could not create pg pool: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("could not ping pg: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	db := &DB{pool: pool, ctxCancel: cancel}

	go db.metrics(ctx)

	return db, nil
}

func (db *DB) SaveRSVP(ctx context.Context, rsvp rsvps.RSVP) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO
			venues (id, name, lat, lon)
		VALUES
			($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, rsvp.Venue.ID, rsvp.Venue.Name, rsvp.Venue.Lat, rsvp.Venue.Lon)

	if err != nil {
		return fmt.Errorf("could not insert venue: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO
			groups (id, country, state, city, name, lat, lon, urlname, topics)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`,
		rsvp.Group.ID,
		rsvp.Group.Country,
		zeronull.Text(rsvp.Group.State),
		rsvp.Group.City,
		rsvp.Group.Name,
		rsvp.Group.Lat,
		rsvp.Group.Lon,
		rsvp.Group.Urlname,
		rsvp.Group.Topics,
	)
	if err != nil {
		return fmt.Errorf("could not insert group: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO
			members (id, name, photo)
		VALUES
			($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, rsvp.Member.ID, rsvp.Member.Name, rsvp.Member.Photo)

	if err != nil {
		return fmt.Errorf("could not insert member: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO
			events (id, name, time, url, venue_id, group_id, member_id)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`,
		rsvp.Event.ID,
		rsvp.Event.Name,
		time.UnixMilli(rsvp.Event.Time).UTC(),
		rsvp.Event.URL,
		rsvp.Venue.ID,
		rsvp.Group.ID,
		rsvp.Member.ID,
	)
	if err != nil {
		return fmt.Errorf("could not insert event: %w", err)
	}

	rsvpResponse := rsvp.Response == "yes"

	_, err = tx.Exec(ctx, `
		INSERT INTO
			rsvps (id, mtime, guests, response, visibility, event_id)
		VALUES
			($1, $2, $3, $4, $5, $6)
	`,
		rsvp.ID,
		time.UnixMilli(rsvp.Mtime).UTC(),
		rsvp.Guests,
		rsvpResponse,
		rsvp.Visibility,
		rsvp.Event.ID,
	)

	if err != nil {
		return fmt.Errorf("could not insert rsvp: %w", err)
	}

	if rsvpResponse {
		_, err = tx.Exec(ctx, `
			INSERT INTO
				event_counters (rsvp_date, event_id, confirmed_rsvps)
			VALUES
				($1, $2, 1)
			ON CONFLICT (rsvp_date, event_id) DO UPDATE
				SET confirmed_rsvps = event_counters.confirmed_rsvps + 1
		`, time.UnixMilli(rsvp.Mtime).UTC().Truncate(24*time.Hour), rsvp.Event.ID)

		if err != nil {
			return fmt.Errorf("could not update counters: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not commit rsvp: %w", err)
	}

	return nil
}

func (db *DB) TopkEvents(ctx context.Context, date time.Time, k uint) ([]dbpkg.TopkEvent, error) {
	date = date.UTC().Truncate(24 * time.Hour)

	rows, err := db.pool.Query(ctx, `
		WITH topk AS (
			SELECT event_id, confirmed_rsvps FROM event_counters WHERE rsvp_date = $1 ORDER BY confirmed_rsvps DESC LIMIT $2
		)
		SELECT
			events.id, events.name, events.time, events.url, topk.confirmed_rsvps
		FROM
			events INNER JOIN topk ON events.id = topk.event_id
		ORDER BY
			topk.confirmed_rsvps DESC
	`, date, k)

	if err != nil {
		return nil, fmt.Errorf("could not query topk events: %w", err)
	}

	var (
		topks []dbpkg.TopkEvent

		topk     dbpkg.TopkEvent
		topkTime time.Time
	)
	_, err = pgx.ForEachRow(rows, []any{&topk.Event.ID, &topk.Event.Name, &topkTime, &topk.Event.URL, &topk.ConfirmedRSVPs}, func() error {
		topk.Event.Time = topkTime.UnixMilli()
		topks = append(topks, topk)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not query topk events: %w", err)
	}

	return topks, nil
}

func (db *DB) GetEventInfo(ctx context.Context, eventID string) (rsvps.EventInfo, error) {
	var (
		eventInfo  rsvps.EventInfo
		groupState zeronull.Text
	)

	err := db.pool.QueryRow(ctx, `
		SELECT
			groups.id,
			groups.country,
			groups.state,
			groups.city,
			groups.name,
			groups.lat,
			groups.lon,
			groups.urlname,
			groups.topics,
			venues.id,
			venues.name,
			venues.lat,
			venues.lon
		FROM
			events
				INNER JOIN groups ON events.group_id = groups.id
				INNER JOIN venues ON events.venue_id = venues.id
		WHERE
			events.id = $1
	`, eventID).Scan(
		&eventInfo.Group.ID,
		&eventInfo.Group.Country,
		&groupState,
		&eventInfo.Group.City,
		&eventInfo.Group.Name,
		&eventInfo.Group.Lat,
		&eventInfo.Group.Lon,
		&eventInfo.Group.Urlname,
		&eventInfo.Group.Topics,
		&eventInfo.Venue.ID,
		&eventInfo.Venue.Name,
		&eventInfo.Venue.Lat,
		&eventInfo.Venue.Lon,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return rsvps.EventInfo{}, dbpkg.ErrNoEvents
		}
		return rsvps.EventInfo{}, fmt.Errorf("could not query event info: %w", err)
	}

	eventInfo.Group.State = string(groupState)

	return eventInfo, nil
}

func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

func (db *DB) metrics(ctx context.Context) {
	for {
		stats := db.pool.Stat()

		pgpoolAcquireCount.Set(uint64(stats.AcquireCount()))
		pgpoolAcquiredConns.Set(uint64(stats.AcquiredConns()))
		pgpoolCanceledAcquireCount.Set(uint64(stats.CanceledAcquireCount()))
		pgpoolConstructingConns.Set(uint64(stats.ConstructingConns()))
		pgpoolEmptyAcquireCount.Set(uint64(stats.EmptyAcquireCount()))
		pgpoolIdleConns.Set(uint64(stats.IdleConns()))
		pgpoolMaxConns.Set(uint64(stats.MaxConns()))
		pgpoolTotalConns.Set(uint64(stats.TotalConns()))
		pgpoolNewConnsCount.Set(uint64(stats.NewConnsCount()))
		pgpoolMaxLifetimeDestroyCount.Set(uint64(stats.MaxLifetimeDestroyCount()))
		pgpoolMaxIdleDestroyCount.Set(uint64(stats.MaxIdleDestroyCount()))

		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
	}
}
