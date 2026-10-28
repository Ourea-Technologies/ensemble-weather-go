// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"context"
	"strings"
	"time"
)

var (
	opLocationForecast = operation{id: "getLocationForecast"}
	opPointForecast    = operation{id: "getPointForecast"}
	opCurrent          = operation{id: "getCurrentConditions"}
	opCurrentAt        = operation{id: "getCurrentByCoordinates"}
	opAlerts           = operation{id: "getLocationAlerts"}
	opAlertsAt         = operation{id: "getAlertsByCoordinates"}
	opNearby           = operation{id: "getNearbyLocations"}
	opNearbyAt         = operation{id: "getNearbyByCoordinates"}
	opSearch           = operation{id: "searchLocations", open: true}
	opAllLocations     = operation{id: "getAllLocations"}
	opStateLocations   = operation{id: "getLocations"}
	opCatalogLocations = operation{id: "listCatalogLocations"}
	opCatalogLocation  = operation{id: "getCatalogLocation", open: true}
	opCatalogStations  = operation{id: "listCatalogStations"}
	opCatalogStation   = operation{id: "getCatalogStation", open: true}
	opRuns             = operation{id: "getAvailableRuns"}
	opAttribution      = operation{id: "getAttribution", open: true}
)

// Point is a coordinate, decimal degrees (WGS84, west negative). The SDK
// sends it exactly as given.
type Point struct {
	Lat, Lon float64
	// ElevationM is the point's elevation in metres, whatever the units.
	// nil lets the server use its 10 m terrain model. Only the forecast and
	// current-conditions routes take it.
	ElevationM *float64
}

// splitLocation splits "state/slug" into its two path segments and
// lowercases a two-letter state code. A state name passes through unchanged
// (the server redirects it to the code).
func splitLocation(op, location string) ([]string, error) {
	state, slug, ok := strings.Cut(location, "/")
	if !ok || state == "" || slug == "" || strings.Contains(slug, "/") {
		return nil, usageError(op, "location %q must be \"state/slug\", for example \"or/mt-hood-meadows\"", location)
	}
	return []string{"locations", normalizeState(state), slug}, nil
}

func normalizeState(s string) string {
	if len(s) == 2 {
		return strings.ToLower(s)
	}
	return s
}

func (c *Client) units(u Units) Units {
	if u == "" {
		return c.cfg.units
	}
	return u
}

// ForecastParams are the options of a forecast. The zero value asks for the
// API defaults: 24 hours, daily summaries, nowcast blended, the client's units.
type ForecastParams struct {
	// Units overrides the client's unit system for this request.
	Units Units
	// MaxHours is the number of hours past the first (6..360; default 24),
	// clamped to the tier's outlook (Free 240, Pro 360).
	MaxHours int
	// MaxDays is the number of daily rows (1..15; default the full horizon).
	MaxDays int
	// IncludePast serves the hourly series from the model run's init.
	IncludePast bool
	// Models adds per-model breakdowns (Pro).
	Models bool
	// Members adds per-member values (Free: the plume view).
	Members bool
	// Daily, Hourly and Nowcast default to true; set one to Ptr(false) to drop it.
	Daily, Hourly, Nowcast *bool
	// Elevations adds elevation forecasts: "all" or elevation names. Only
	// location forecasts take it.
	Elevations []string
	// Run pins a historical model run by its init time (Pro).
	Run time.Time
	// Sources and ExcludeSources select observation sources for the nowcast.
	Sources, ExcludeSources []string
	// Chances are probability terms, answered in Probabilities in this order.
	Chances []Chance
	// Totals are window totals, answered in WindowTotals in this order.
	Totals []Total
	// Diagnostics opts in to serve-path diagnostics: "provenance".
	Diagnostics []string
	// Include opts in to extra forecast variables: "mslp" and
	// "skin_temperature" (EnsembleStats.MSLP and SkinTemperature).
	Include []string
	// FeelsLike selects feels-like indices (default "nws", the only one).
	FeelsLike []string
}

type forecastPlan struct {
	q       query
	units   Units
	chances chanceQuery
	totals  totalsQuery
}

func (c *Client) planForecast(op string, p *ForecastParams) (*forecastPlan, error) {
	if p == nil {
		p = &ForecastParams{}
	}
	plan := &forecastPlan{units: c.units(p.Units)}
	var err error
	if plan.chances, err = buildChances(op, p.Chances, plan.units); err != nil {
		return nil, err
	}
	if plan.totals, err = buildTotals(op, p.Totals); err != nil {
		return nil, err
	}
	q := &plan.q
	q.units(plan.units)
	q.intDefault("max_hours", p.MaxHours, 24)
	q.intDefault("max_days", p.MaxDays, 0)
	q.flag("include_past", p.IncludePast)
	q.flag("models", p.Models)
	q.flag("members", p.Members)
	q.flagDefault("daily", p.Daily, true)
	q.flagDefault("hourly", p.Hourly, true)
	q.flagDefault("nowcast", p.Nowcast, true)
	q.list("elevations", p.Elevations)
	q.time("run", p.Run)
	q.list("sources", p.Sources)
	q.list("exclude_sources", p.ExcludeSources)
	q.literal("prob", plan.chances.value)
	q.literal("totals", plan.totals.value)
	q.list("diagnostics", p.Diagnostics)
	q.sortedList("include", p.Include)
	if len(p.FeelsLike) != 1 || p.FeelsLike[0] != "nws" {
		q.list("feels_like", p.FeelsLike)
	}
	return plan, nil
}

func (c *Client) forecast(ctx context.Context, op operation, segments []string, plan *forecastPlan) (*ForecastResponse, error) {
	var f ForecastResponse
	if err := c.get(ctx, op, segments, &plan.q, &f); err != nil {
		return nil, err
	}
	var err error
	if f.Probabilities, err = mapChances(op.id, plan.chances, f.Probabilities, f.Units); err != nil {
		return nil, err
	}
	if f.WindowTotals, err = mapTotals(op.id, plan.totals, f.WindowTotals, f.Units); err != nil {
		return nil, err
	}
	return &f, nil
}

// Forecast returns the forecast for a named location, given as "state/slug"
// (for example "or/mt-hood-meadows"). It needs no API key.
func (c *Client) Forecast(ctx context.Context, location string, params *ForecastParams) (*ForecastResponse, error) {
	segs, err := splitLocation(opLocationForecast.id, location)
	if err != nil {
		return nil, err
	}
	plan, err := c.planForecast(opLocationForecast.id, params)
	if err != nil {
		return nil, err
	}
	return c.forecast(ctx, opLocationForecast, segs, plan)
}

// ForecastAt returns the forecast at a point. It needs a Pro API key.
func (c *Client) ForecastAt(ctx context.Context, at Point, params *ForecastParams) (*ForecastResponse, error) {
	if params != nil && len(params.Elevations) > 0 {
		return nil, usageError(opPointForecast.id, "Elevations applies only to location forecasts; set Point.ElevationM instead")
	}
	plan, err := c.planForecast(opPointForecast.id, params)
	if err != nil {
		return nil, err
	}
	plan.q.num("lat", at.Lat)
	plan.q.num("lon", at.Lon)
	plan.q.numPtr("elevation_m", at.ElevationM)
	return c.forecast(ctx, opPointForecast, []string{"forecast"}, plan)
}

// CurrentParams are the options of current conditions.
type CurrentParams struct {
	// Units overrides the client's unit system for this request.
	Units Units
	// Sources and ExcludeSources select observation sources.
	Sources, ExcludeSources []string
	// IncludeNowcastHours adds the observation-blended hours (NowcastHours).
	IncludeNowcastHours bool
}

func (c *Client) currentQuery(p *CurrentParams) *query {
	if p == nil {
		p = &CurrentParams{}
	}
	q := &query{}
	q.units(c.units(p.Units))
	q.list("sources", p.Sources)
	q.list("exclude_sources", p.ExcludeSources)
	if p.IncludeNowcastHours {
		q.set("include", "nowcast_hours")
	}
	return q
}

// Current returns current conditions and active alerts at a named location.
func (c *Client) Current(ctx context.Context, location string, params *CurrentParams) (*CurrentResponse, error) {
	segs, err := splitLocation(opCurrent.id, location)
	if err != nil {
		return nil, err
	}
	var r CurrentResponse
	if err := c.get(ctx, opCurrent, append(segs, "current"), c.currentQuery(params), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CurrentAt returns current conditions at a point. It needs an API key.
func (c *Client) CurrentAt(ctx context.Context, at Point, params *CurrentParams) (*CurrentResponse, error) {
	q := c.currentQuery(params)
	q.num("lat", at.Lat)
	q.num("lon", at.Lon)
	q.numPtr("elevation_m", at.ElevationM)
	var r CurrentResponse
	if err := c.get(ctx, opCurrentAt, []string{"current"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Alerts returns the weather alerts in effect at a named location.
func (c *Client) Alerts(ctx context.Context, location string) (*AlertsResponse, error) {
	segs, err := splitLocation(opAlerts.id, location)
	if err != nil {
		return nil, err
	}
	var r AlertsResponse
	if err := c.get(ctx, opAlerts, append(segs, "alerts"), nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// AlertsAt returns the weather alerts in effect at a point. It needs an API key.
func (c *Client) AlertsAt(ctx context.Context, at Point) (*AlertsResponse, error) {
	if at.ElevationM != nil {
		return nil, usageError(opAlertsAt.id, "alerts take no elevation")
	}
	q := &query{}
	q.num("lat", at.Lat)
	q.num("lon", at.Lon)
	var r AlertsResponse
	if err := c.get(ctx, opAlertsAt, []string{"alerts"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// NearbyParams are the options of a nearby-locations search.
type NearbyParams struct {
	// Units sets the unit of the rows' elevations.
	Units Units
	// Limit is the number of neighbours (1..20; default 3).
	Limit int
	// RadiusKm is the search radius (up to 100; default 50).
	RadiusKm float64
}

func (c *Client) nearbyQuery(p *NearbyParams) *query {
	if p == nil {
		p = &NearbyParams{}
	}
	q := &query{}
	q.units(c.units(p.Units))
	q.intDefault("limit", p.Limit, 3)
	if p.RadiusKm != 0 && p.RadiusKm != 50 {
		q.num("radius_km", p.RadiusKm)
	}
	return q
}

// Nearby returns the locations near a named location.
func (c *Client) Nearby(ctx context.Context, location string, params *NearbyParams) (*NearbyLocationsResponse, error) {
	segs, err := splitLocation(opNearby.id, location)
	if err != nil {
		return nil, err
	}
	var r NearbyLocationsResponse
	if err := c.get(ctx, opNearby, append(segs, "nearby"), c.nearbyQuery(params), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// NearbyAt returns the locations near a point.
func (c *Client) NearbyAt(ctx context.Context, at Point, params *NearbyParams) (*NearbyPointResponse, error) {
	if at.ElevationM != nil {
		return nil, usageError(opNearbyAt.id, "nearby takes no elevation")
	}
	q := c.nearbyQuery(params)
	q.num("lat", at.Lat)
	q.num("lon", at.Lon)
	var r NearbyPointResponse
	if err := c.get(ctx, opNearbyAt, []string{"nearby"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SearchParams are the filters of a location search.
type SearchParams struct {
	// State filters by two-letter code or state name.
	State string
	// County filters by county name.
	County string
	// Type filters by location type.
	Type LocationType
	// Limit is the maximum number of hits (1..100; default 20).
	Limit int
}

// Search finds locations by free text (at least 2 characters). It is sent
// without the API key, so results are shared at the edge cache.
func (c *Client) Search(ctx context.Context, text string, params *SearchParams) (*SearchLocationsResponse, error) {
	if params == nil {
		params = &SearchParams{}
	}
	q := &query{}
	q.set("q", text)
	q.str("state", params.State)
	q.str("county", params.County)
	q.str("type", string(params.Type))
	q.intDefault("limit", params.Limit, 20)
	var r SearchLocationsResponse
	if err := c.get(ctx, opSearch, []string{"search", "locations"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// LocationsParams select a page of named locations.
type LocationsParams struct {
	// State limits the list to one state; empty lists every state.
	State string
	// Type filters by location type.
	Type LocationType
	// Units sets the unit of the rows' elevations.
	Units Units
	// Limit is the page size (1..5000).
	Limit int
	// Cursor is the previous page's NextCursor.
	Cursor string
}

// Locations returns one page of named locations: a state's, or every
// state's when params.State is empty. Follow NextCursor for the next page.
func (c *Client) Locations(ctx context.Context, params *LocationsParams) (*LocationsResponse, error) {
	if params == nil {
		params = &LocationsParams{}
	}
	op, segs := opAllLocations, []string{"locations"}
	if params.State != "" {
		op, segs = opStateLocations, []string{"locations", normalizeState(params.State)}
	}
	q := &query{}
	q.str("type", string(params.Type))
	q.units(c.units(params.Units))
	q.intDefault("limit", params.Limit, 0)
	q.str("cursor", params.Cursor)
	var r LocationsResponse
	if err := c.get(ctx, op, segs, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CatalogLocationsParams filter the location catalog.
type CatalogLocationsParams struct {
	// Admin1 is an ISO 3166-2 code, such as "US-OR".
	Admin1 string
	// Country is an ISO 3166-1 alpha-2 code, such as "US".
	Country string
	// Type is a location type, such as LocationSkiArea.
	Type LocationType
	// Template is a template variant, such as TemplateMountain.
	Template TemplateVariant
	// Limit is the page size (1..100; default 20).
	Limit int
	// Cursor is the previous page's NextCursor.
	Cursor string
}

// CatalogLocations returns one page of the location catalog.
func (c *Client) CatalogLocations(ctx context.Context, params *CatalogLocationsParams) (*CatalogLocationPage, error) {
	if params == nil {
		params = &CatalogLocationsParams{}
	}
	q := &query{}
	q.str("admin1", params.Admin1)
	q.str("country", params.Country)
	q.str("type", string(params.Type))
	q.str("template", string(params.Template))
	q.intDefault("limit", params.Limit, 20)
	q.str("cursor", params.Cursor)
	var r CatalogLocationPage
	if err := c.get(ctx, opCatalogLocations, []string{"catalog", "locations"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CatalogDetailParams select the extras of a catalog record.
type CatalogDetailParams struct {
	// Include opts in to "content" and "template".
	Include []string
	// Locale is the BCP-47 locale for content (default "en").
	Locale string
}

func catalogDetailQuery(p *CatalogDetailParams) *query {
	q := &query{}
	if p != nil {
		q.list("include", p.Include)
		q.strDefault("locale", p.Locale, "en")
	}
	return q
}

// CatalogLocation returns a location's full catalog record. It is sent
// without the API key, so the response is shared at the edge cache.
func (c *Client) CatalogLocation(ctx context.Context, slug string, params *CatalogDetailParams) (*CatalogLocationDetail, error) {
	var r CatalogLocationDetail
	if err := c.get(ctx, opCatalogLocation, []string{"catalog", "locations", slug}, catalogDetailQuery(params), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CatalogStationsParams filter the station catalog.
type CatalogStationsParams struct {
	// Admin1 is an ISO 3166-2 code, such as "US-OR".
	Admin1 string
	// Country is an ISO 3166-1 alpha-2 code, such as "US".
	Country string
	// Networks filters by network, such as NetworkSNOTEL (any of them).
	Networks []StationNetwork
	// Capabilities lists capabilities a station must all have, such as "swe".
	Capabilities []string
	// Status is a station status, such as StationActive.
	Status StationStatus
	// Limit is the page size (1..100; default 20).
	Limit int
	// Cursor is the previous page's NextCursor.
	Cursor string
}

// CatalogStations returns one page of the station catalog.
func (c *Client) CatalogStations(ctx context.Context, params *CatalogStationsParams) (*CatalogStationPage, error) {
	if params == nil {
		params = &CatalogStationsParams{}
	}
	q := &query{}
	q.str("admin1", params.Admin1)
	q.str("country", params.Country)
	networks := make([]string, len(params.Networks))
	for i, n := range params.Networks {
		networks[i] = string(n)
	}
	q.sortedList("network", networks)
	q.sortedList("capability", params.Capabilities)
	q.str("status", string(params.Status))
	q.intDefault("limit", params.Limit, 20)
	q.str("cursor", params.Cursor)
	var r CatalogStationPage
	if err := c.get(ctx, opCatalogStations, []string{"catalog", "stations"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CatalogStation returns a station's full catalog record. It is sent without
// the API key, so the response is shared at the edge cache.
func (c *Client) CatalogStation(ctx context.Context, slug string, params *CatalogDetailParams) (*CatalogStationDetail, error) {
	var r CatalogStationDetail
	if err := c.get(ctx, opCatalogStation, []string{"catalog", "stations", slug}, catalogDetailQuery(params), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Runs returns the available model runs, newest first. limit is the maximum
// number (up to 100); 0 means the default, 10.
func (c *Client) Runs(ctx context.Context, limit int) (*RunsResponse, error) {
	q := &query{}
	q.intDefault("limit", limit, 10)
	var r RunsResponse
	if err := c.get(ctx, opRuns, []string{"runs"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Attribution returns every data source with its licence, terms and required
// notices. It is sent without the API key.
func (c *Client) Attribution(ctx context.Context) (*AttributionRegistry, error) {
	var r AttributionRegistry
	if err := c.get(ctx, opAttribution, []string{"attribution"}, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
