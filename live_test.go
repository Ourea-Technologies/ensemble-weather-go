// SPDX-License-Identifier: BSD-2-Clause

//go:build live

package ensembleweather_test

// The nightly contract run (.github/workflows/live.yml) against the live API.
// It runs only with -tags live and a read-only key in OUREA_LIVE_API_KEY:
//
//	OUREA_LIVE_API_KEY=ow_... go test -tags live -run '^TestLive' -v .

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	ensembleweather "oureawx.com/go/ensemble-weather"
)

func liveClient(t *testing.T, keyed bool) *ensembleweather.Client {
	t.Helper()
	key := os.Getenv("OUREA_LIVE_API_KEY")
	if key == "" {
		t.Skip("OUREA_LIVE_API_KEY is not set")
	}
	if !keyed {
		key = ""
	}
	return ensembleweather.NewClient(ensembleweather.WithAPIKey(key), ensembleweather.WithBaseURL(ensembleweather.DefaultBaseURL),
		ensembleweather.WithUserAgent("ensemble-weather-go-ci"))
}

func liveContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func TestLiveAnonymousForecast(t *testing.T) {
	ctx := liveContext(t)
	f, err := liveClient(t, false).Forecast(ctx, "or/mt-hood-meadows", &ensembleweather.ForecastParams{
		Chances: []ensembleweather.Chance{{Variable: ensembleweather.Snowfall, Op: ensembleweather.AtLeast, Threshold: ensembleweather.Inches(6)}},
		Totals:  []ensembleweather.Total{{Variable: ensembleweather.Snowfall, Window: ensembleweather.Hours(48)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if f.Location.Slug != "mt-hood-meadows" || f.Units != ensembleweather.UnitsMetric || len(f.Forecast.Timeseries) == 0 {
		t.Errorf("unexpected forecast: %+v", f.Location)
	}
	if len(f.Probabilities) != 1 || len(f.WindowTotals) != 1 {
		t.Errorf("probabilities %d, totals %d; want 1 each", len(f.Probabilities), len(f.WindowTotals))
	}
}

func TestLiveKeyedForecastAndRateLimit(t *testing.T) {
	ctx := liveContext(t)
	f, err := liveClient(t, true).Forecast(ctx, "or/mt-hood-meadows", &ensembleweather.ForecastParams{Units: ensembleweather.UnitsImperial})
	if err != nil {
		t.Fatal(err)
	}
	if f.Units != ensembleweather.UnitsImperial {
		t.Errorf("units = %s", f.Units)
	}
	if f.RateLimit == nil {
		t.Error("a keyed response carries no rate-limit state")
	}
}

func TestLiveOpenRoutes(t *testing.T) {
	ctx := liveContext(t)
	c := liveClient(t, true)
	hits, err := c.Search(ctx, "hood meadows", nil)
	if err != nil || hits.Count == 0 {
		t.Fatalf("search: %v (%d hits)", err, hits.Count)
	}
	if _, err := c.Attribution(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.CatalogLocations(ctx, &ensembleweather.CatalogLocationsParams{Limit: 2}); err != nil {
		t.Fatal(err)
	}
}

func TestLiveErrors(t *testing.T) {
	ctx := liveContext(t)
	_, err := liveClient(t, false).Forecast(ctx, "or/no-such-location-ci", nil)
	if !errors.Is(err, ensembleweather.ErrNotFound) {
		t.Errorf("err = %v, want not found", err)
	}
}
