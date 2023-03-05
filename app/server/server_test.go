package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/oizgagin/ing/app/server"
	cachepkg "github.com/oizgagin/ing/pkg/cache"
	cachemocks "github.com/oizgagin/ing/pkg/cache/mocks"
	configtypes "github.com/oizgagin/ing/pkg/config/types"
	dbpkg "github.com/oizgagin/ing/pkg/db"
	dbmocks "github.com/oizgagin/ing/pkg/db/mocks"
	"github.com/oizgagin/ing/pkg/rsvps"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	topkEvents = []dbpkg.TopkEvent{
		{Event: rsvps.Event{ID: "event_id1", Name: "event_name1", URL: "event_url1", Time: 1001}, ConfirmedRSVPs: 10},
		{Event: rsvps.Event{ID: "event_id2", Name: "event_name2", URL: "event_url1", Time: 2001}, ConfirmedRSVPs: 20},
		{Event: rsvps.Event{ID: "event_id3", Name: "event_name3", URL: "event_url1", Time: 3001}, ConfirmedRSVPs: 30},
	}

	eventInfo1 = rsvps.EventInfo{
		Group: rsvps.Group{ID: 1002, Name: "group_name1", Country: "US", City: "group_city1"},
		Venue: rsvps.Venue{ID: 1003, Name: "venue_name1", Lat: 11, Lon: 12},
	}
)

func TestServer(t *testing.T) {

	t.Run("eventsTopk", func(t *testing.T) {
		dbMock, _, server, tearDown := setUp(t, time.Second)
		defer tearDown(t)

		dbMock.
			On("TopkEvents", mock.Anything, time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC), uint(3)).
			Return(topkEvents, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/events/topk?date=2023-03-05&k=3", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		var resp []dbpkg.TopkEvent
		err := json.NewDecoder(rec.Result().Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, topkEvents, resp)

		require.Equal(t, 200, rec.Result().StatusCode)
		require.Equal(t, "application/json", rec.Result().Header.Get("Content-Type"))

	})

	t.Run("eventsInfo", func(t *testing.T) {
		cacheTTL := time.Second

		dbMock, cacheMock, server, tearDown := setUp(t, cacheTTL)
		defer tearDown(t)

		dbMock.
			On("GetEventInfo", mock.Anything, "event_id1").
			Return(func(ctx context.Context, eventID string) (rsvps.EventInfo, error) {
				return eventInfo1, nil
			})

		cacheMock.
			On("Get", mock.Anything, "event_id1").
			Return(rsvps.EventInfo{}, cachepkg.ErrNoCachedEventInfo)

		cacheMock.
			On("Set", mock.Anything, "event_id1", eventInfo1, cacheTTL).
			Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/events/info?event_id=event_id1", nil)
		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, req)

		var resp rsvps.EventInfo
		err := json.NewDecoder(rec.Result().Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, eventInfo1, resp)

		require.Equal(t, 200, rec.Result().StatusCode)
		require.Equal(t, "application/json", rec.Result().Header.Get("Content-Type"))
	})

}

func setUp(t *testing.T, cacheTTL time.Duration) (*dbmocks.DB, *cachemocks.EventInfoCache, *server.Server, func(t *testing.T)) {
	t.Helper()

	db := dbmocks.NewDB(t)
	eventCache := cachemocks.NewEventInfoCache(t)

	logger := zaptest.NewLogger(t)

	cfg := server.Config{
		Addr:            ":0",
		CacheTTL:        configtypes.Duration{Duration: cacheTTL},
		CacheSetTimeout: configtypes.Duration{Duration: time.Second},
	}

	server, err := server.NewServer(cfg, logger, db, eventCache)
	require.NoError(t, err)

	return db, eventCache, server, func(t *testing.T) {
		require.NoError(t, server.Close())
		require.NoError(t, logger.Sync())
	}
}
