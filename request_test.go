// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// recorder is a RoundTripper that answers every request with fn and keeps
// the requests.
type recorder struct {
	mu   sync.Mutex
	reqs []*http.Request
	fn   func(n int, r *http.Request) (*http.Response, error)
}

func (rt *recorder) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.mu.Lock()
	rt.reqs = append(rt.reqs, r)
	n := len(rt.reqs)
	rt.mu.Unlock()
	return rt.fn(n, r)
}

func (rt *recorder) last() *http.Request {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return rt.reqs[len(rt.reqs)-1]
}

func respond(status int, body string, kv ...string) *http.Response {
	h := http.Header{"Content-Type": {"application/json"}}
	for i := 0; i+1 < len(kv); i += 2 {
		h.Set(kv[i], kv[i+1])
	}
	return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

const minimalForecast = `{"location":{"slug":"mt-hood-meadows","state":"or","name":"Mt. Hood Meadows","lat":45.3,"lon":-121.6,"elevation_m":1661,"type":"ski_area","timezone":"America/Los_Angeles"},"model_run":{"initialized_at":"2026-11-15T00:00:00Z","total_members":131,"models":[]},"units":"metric","forecast":{"max_hours":24,"interval_hours":1},"nowcast_applied":true,"attribution":{"sources":[],"credit":"","url":""}}`

func testClient(t *testing.T, fn func(n int, r *http.Request) (*http.Response, error), opts ...Option) (*Client, *recorder, *[]float64) {
	t.Helper()
	t.Setenv(EnvAPIKey, "")
	t.Setenv(EnvBaseURL, "")
	rt := &recorder{fn: fn}
	var delays []float64
	all := append([]Option{WithHTTPClient(&http.Client{Transport: rt}), withTestSeams(&delays)}, opts...)
	return NewClient(all...), rt, &delays
}

func okForecast(int, *http.Request) (*http.Response, error) {
	return respond(200, minimalForecast), nil
}

func TestQueryCanonicalForm(t *testing.T) {
	tests := []struct {
		name   string
		params *ForecastParams
		want   string
	}{
		{"nil params", nil, ""},
		{"defaults omitted", &ForecastParams{Units: UnitsMetric, MaxHours: 24, Daily: Ptr(true), Hourly: Ptr(true), Nowcast: Ptr(true), FeelsLike: []string{"nws"}}, ""},
		{"false booleans sent", &ForecastParams{Daily: Ptr(false), Nowcast: Ptr(false)}, "daily=false&nowcast=false"},
		{"lists are percent-encoded", &ForecastParams{Elevations: []string{"base", "summit"}}, "elevations=base%2Csummit"},
		{"run in UTC RFC 3339", &ForecastParams{Run: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)}, "run=2026-10-14T12%3A00%3A00Z"},
		{"kelvin units", &ForecastParams{Units: UnitsKelvin, MaxDays: 3}, "max_days=3&units=kelvin"},
		{"totals literal", &ForecastParams{Totals: []Total{{Snowfall, Hours(72)}, {Precipitation, HoursBetween(24, 48)}, {Snowfall, Hours(72)}}}, "totals=precipitation:24-48h,snowfall:72h"},
		{"window default sent", &ForecastParams{Totals: []Total{{Variable: Snowfall}}}, "totals=snowfall:24h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rt, _ := testClient(t, okForecast)
			if _, err := c.Forecast(context.Background(), "or/mt-hood-meadows", tt.params); err != nil {
				t.Fatal(err)
			}
			if got := rt.last().URL.RawQuery; got != tt.want {
				t.Errorf("query = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEscape(t *testing.T) {
	tests := []struct{ in, keep, want string }{
		{"Mt. Hood & co/ñ", "", "Mt.%20Hood%20%26%20co%2F%C3%B1"},
		{"a+b=c?d#e", "", "a%2Bb%3Dc%3Fd%23e"},
		{"-._~AZaz09", "", "-._~AZaz09"},
		{"snowfall:ge:6.000:24h,x", ":,", "snowfall:ge:6.000:24h,x"},
		{"a:b,c", "", "a%3Ab%2Cc"},
	}
	for _, tt := range tests {
		if got := escape(tt.in, tt.keep); got != tt.want {
			t.Errorf("escape(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	tests := map[float64]string{45.3311: "45.3311", -121.6649: "-121.6649", 1661: "1661", 1e21: "1000000000000000000000", 0.0000001: "0.0000001"}
	for in, want := range tests {
		if got := formatNumber(in); got != want {
			t.Errorf("formatNumber(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestCatalogListsCanonical(t *testing.T) {
	c, rt, _ := testClient(t, func(int, *http.Request) (*http.Response, error) {
		return respond(200, `{"items":[],"next_cursor":null}`), nil
	})
	_, err := c.CatalogStations(context.Background(), &CatalogStationsParams{
		Capabilities: []string{"swe", "air_temp", "swe"}, Networks: []StationNetwork{NetworkSNOTEL, NetworkRAWS, NetworkSNOTEL}, Status: StationActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := rt.last().URL.RawQuery, "capability=air_temp%2Cswe&network=raws%2Csnotel&status=active"; got != want {
		t.Errorf("query = %q, want %q", got, want)
	}
}

func TestLocationParsing(t *testing.T) {
	tests := []struct {
		in, path string
		usage    bool
	}{
		{"or/mt-hood-meadows", "/api/v1/locations/or/mt-hood-meadows", false},
		{"OR/mt-hood-meadows", "/api/v1/locations/or/mt-hood-meadows", false},
		{"Oregon/mt-hood-meadows", "/api/v1/locations/Oregon/mt-hood-meadows", false},
		{"new york/x y", "/api/v1/locations/new%20york/x%20y", false},
		{"mt-hood-meadows", "", true},
		{"or/a/b", "", true},
		{"/x", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			c, rt, _ := testClient(t, okForecast)
			_, err := c.Forecast(context.Background(), tt.in, nil)
			if tt.usage {
				if !errors.Is(err, ErrUsage) || len(rt.reqs) != 0 {
					t.Fatalf("err = %v, requests = %d; want a usage error and no request", err, len(rt.reqs))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := rt.last().URL.EscapedPath(); got != tt.path {
				t.Errorf("path = %q, want %q", got, tt.path)
			}
		})
	}
}

func TestPointRoutes(t *testing.T) {
	c, rt, _ := testClient(t, okForecast, WithAPIKey("ow_pro"))
	ctx := context.Background()
	if _, err := c.GetPointForecast(ctx, Point{Lat: 45.33, Lon: -121.71}, &ForecastParams{Elevations: []string{"all"}}); !errors.Is(err, ErrUsage) {
		t.Errorf("Elevations on a point forecast: err = %v, want usage", err)
	}
	if _, err := c.GetPointAlerts(ctx, Point{Lat: 1, Lon: 2, ElevationM: Ptr(3.0)}); !errors.Is(err, ErrUsage) {
		t.Errorf("elevation on alerts: err = %v, want usage", err)
	}
	if len(rt.reqs) != 0 {
		t.Fatalf("usage errors sent %d requests", len(rt.reqs))
	}
	_, _ = c.GetPointCurrent(ctx, Point{Lat: 45.33, Lon: -121.71, ElevationM: Ptr(0.0)}, &CurrentParams{IncludeNowcastHours: true, Units: UnitsImperial})
	if got, want := rt.last().URL.RawQuery, "elevation_m=0&include=nowcast_hours&lat=45.33&lon=-121.71&units=imperial"; got != want {
		t.Errorf("current query = %q, want %q", got, want)
	}
	_, _ = c.GetPointNearby(ctx, Point{Lat: 45, Lon: -121}, &NearbyParams{Limit: 3, RadiusKm: 25})
	if got, want := rt.last().URL.RawQuery, "lat=45&lon=-121&radius_km=25"; got != want {
		t.Errorf("nearby query = %q, want %q", got, want)
	}
}

func TestClientUnitsDefault(t *testing.T) {
	c, rt, _ := testClient(t, okForecast, WithUnits(UnitsImperial))
	_, _ = c.Forecast(context.Background(), "or/x", nil)
	if got := rt.last().URL.RawQuery; got != "units=imperial" {
		t.Errorf("query = %q, want units=imperial", got)
	}
	_, _ = c.Forecast(context.Background(), "or/x", &ForecastParams{Units: UnitsMetric})
	if got := rt.last().URL.RawQuery; got != "" {
		t.Errorf("per-request metric: query = %q, want empty", got)
	}
}

func TestUserAgent(t *testing.T) {
	c, rt, _ := testClient(t, okForecast, WithUserAgent("my-app/1.2"))
	_, _ = c.Forecast(context.Background(), "or/x", nil)
	ua := rt.last().Header.Get("User-Agent")
	if !strings.HasPrefix(ua, "ensemble-weather-go/"+Version+" (go/") || !strings.HasSuffix(ua, ") my-app/1.2") {
		t.Errorf("User-Agent = %q", ua)
	}
}

func TestInvalidBaseURL(t *testing.T) {
	for _, base := range []string{"ftp://x", "not a url", "https://", "https://h/?q=1"} {
		c, rt, _ := testClient(t, okForecast, WithBaseURL(base))
		if _, err := c.Forecast(context.Background(), "or/x", nil); !errors.Is(err, ErrUsage) {
			t.Errorf("base %q: err = %v, want usage", base, err)
		}
		if len(rt.reqs) != 0 {
			t.Errorf("base %q: sent a request", base)
		}
	}
}

func TestBaseURLFromEnvironment(t *testing.T) {
	t.Setenv(EnvAPIKey, "")
	t.Setenv(EnvBaseURL, "http://127.0.0.1:9/prefix/")
	rt := &recorder{fn: okForecast}
	c := NewClient(WithHTTPClient(&http.Client{Transport: rt}))
	_, _ = c.Attribution(context.Background())
	if got := rt.last().URL.String(); got != "http://127.0.0.1:9/prefix/api/v1/attribution" {
		t.Errorf("URL = %q", got)
	}
}

func TestForecastIncludeCanonical(t *testing.T) {
	c, rt, _ := testClient(t, okForecast)
	_, err := c.Forecast(context.Background(), "or/x", &ForecastParams{Include: []string{"skin_temperature", "mslp", "mslp"}})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := rt.last().URL.RawQuery, "include=mslp%2Cskin_temperature"; got != want {
		t.Errorf("query = %q, want %q", got, want)
	}
}

func TestCoordinateLocationHasNoKey(t *testing.T) {
	var f ForecastResponse
	if err := json.Unmarshal([]byte(`{"location":{"slug":null,"type":null,"state":"or","name":"45.33, -121.71"}}`), &f); err != nil {
		t.Fatal(err)
	}
	if f.Location.Slug != "" || f.Location.Type != "" || f.Location.Key() != "" {
		t.Errorf("location = %+v, key %q", f.Location, f.Location.Key())
	}
	if got := (LocationRef{State: "or", Slug: "mt-hood-meadows"}).Key(); got != "or/mt-hood-meadows" {
		t.Errorf("Key() = %q", got)
	}
}
