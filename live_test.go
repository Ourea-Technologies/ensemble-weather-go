// SPDX-License-Identifier: BSD-2-Clause

//go:build live && go1.23

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
	if f.Location.Slug != "mt-hood-meadows" || f.Units != ensembleweather.UnitsMetric || len(f.Series.Timeseries) == 0 {
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

func TestLivePlumeAndCurrentPass(t *testing.T) {
	ctx := liveContext(t)
	c := liveClient(t, false)
	f, err := c.Forecast(ctx, "or/mt-hood-meadows", &ensembleweather.ForecastParams{Members: ensembleweather.MembersAsPlume, MaxHours: 12})
	if err != nil {
		t.Fatal(err)
	}
	if f.Series.Plume == nil || len(f.Series.Plume.Headline.ValidTime) == 0 || f.MembersView != ensembleweather.MembersPool {
		t.Errorf("plume view: plume %v, members_view %q", f.Series.Plume != nil, f.MembersView)
	}
	now, err := c.Current(ctx, "wa/snoqualmie-pass-i-90", nil)
	if err != nil {
		t.Fatal(err)
	}
	if now.Pass == nil || now.Summary == nil {
		t.Errorf("pass current conditions: pass %v, summary %v", now.Pass != nil, now.Summary != nil)
	}
}

func TestLiveNewRoutes(t *testing.T) {
	ctx := liveContext(t)
	c := liveClient(t, false)
	var items int
	for _, err := range c.AllExplore(ctx, ensembleweather.ExploreSnow5d, &ensembleweather.ExploreParams{Limit: 10, State: "OR"}) {
		if err != nil {
			t.Fatal(err)
		}
		items++
	}
	if items == 0 {
		t.Error("explore: no items")
	}
	hits, err := c.Search(ctx, "hood", &ensembleweather.SearchParams{Sort: ensembleweather.ExploreSnow5d, Near: &ensembleweather.Point{Lat: 45.33, Lon: -121.66}})
	if err != nil || hits.Count == 0 {
		t.Fatalf("sorted search: %v", err)
	}
	if _, err := c.Season(ctx, "or/mt-hood-meadows", &ensembleweather.SeasonParams{Include: []string{"daily"}}); err != nil {
		t.Fatal(err)
	}
	if comp, err := c.Composition(ctx); err != nil || comp.ModelCount == 0 {
		t.Fatalf("composition: %v", err)
	}
	if sp, err := c.Snowpack(ctx, "or/mt-hood-meadows", nil); err != nil || !sp.Beta {
		t.Fatalf("snowpack: %v", err)
	}
	if _, err := c.SnowpackPlume(ctx, "or/mt-hood-meadows", nil); err != nil {
		t.Fatal(err)
	}
	if out, err := c.AlertOutlines(ctx, &ensembleweather.AlertOutlinesParams{Classes: []ensembleweather.AlertClass{ensembleweather.AlertWinter}}); err != nil || out.Type != "FeatureCollection" {
		t.Fatalf("alert outlines: %v", err)
	}
	if zones, err := c.AvalancheZones(ctx); err != nil || len(zones.Features) == 0 {
		t.Fatalf("avalanche zones: %v", err)
	}
	if _, err := c.SnowTrend(ctx, "or/mt-hood-meadows", nil); !errors.Is(err, ensembleweather.ErrPermission) {
		t.Errorf("anonymous snow trend: err = %v, want PLUS_REQUIRED", err)
	}
}
