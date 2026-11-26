// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"context"
	"math"
	"strings"
	"time"
)

var (
	opLocationForecast = operation{id: "getLocationForecast"}
	opPointForecast    = operation{id: "getPointForecast"}
	opCurrent          = operation{id: "getCurrentConditions"}
	opPointCurrent     = operation{id: "getCurrentByCoordinates"}
	opAlerts           = operation{id: "getLocationAlerts"}
	opPointAlerts      = operation{id: "getAlertsByCoordinates"}
	opNearby           = operation{id: "getNearbyLocations"}
	opPointNearby      = operation{id: "getNearbyByCoordinates"}
	opSearch           = operation{id: "searchLocations", open: true}
	opAllLocations     = operation{id: "getAllLocations"}
	opStateLocations   = operation{id: "getLocations"}
	opCatalogLocations = operation{id: "listCatalogLocations"}
	opCatalogLocation  = operation{id: "getCatalogLocation", open: true}
	opCatalogStations  = operation{id: "listCatalogStations"}
	opCatalogStation   = operation{id: "getCatalogStation", open: true}
	opRuns             = operation{id: "getAvailableRuns"}
	opAttribution      = operation{id: "getAttribution", open: true}
	opExplore          = operation{id: "explore"}
	opComposition      = operation{id: "getComposition", open: true}
	opSeason           = operation{id: "getLocationSeason"}
	opSnowTrend        = operation{id: "getLocationSnowTrend"}
	opSnowpack         = operation{id: "getLocationSnowpack"}
	opSnowpackPlume    = operation{id: "getLocationSnowpackPlume"}
	opAlertOutlines    = operation{id: "getAlertOutlines"}
	opAvalancheZones   = operation{id: "getAvalancheZones"}
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
	// Members adds per-member values: [MembersOn] in each hour, or
	// [MembersAsPlume] as columns in Series.Plume. The zero value (or
	// [MembersOff]) adds none; any other value is refused.
	Members MembersMode
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

// MembersMode selects a forecast's per-member values (the members query
// parameter, which takes true, false or "plume").
type MembersMode string

// The member modes.
const (
	// MembersOff adds no per-member values (the default; the zero value too).
	MembersOff MembersMode = "false"
	// MembersOn adds each hour's member values (TimeSeriesPoint.Members).
	MembersOn MembersMode = "true"
	// MembersAsPlume serves the members as columns in Series.Plume, instead
	// of the hourly series and daily summaries.
	MembersAsPlume MembersMode = "plume"
)

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
	switch p.Members {
	case "", MembersOff:
	case MembersOn, MembersAsPlume:
		q.set("members", string(p.Members))
	default:
		return nil, usageError(op, "Members %q must be MembersOff, MembersOn or MembersAsPlume", string(p.Members))
	}
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
	q.list("include", p.Include)
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

// GetPointForecast returns the forecast at a point. It needs a Pro API key.
func (c *Client) GetPointForecast(ctx context.Context, at Point, params *ForecastParams) (*ForecastResponse, error) {
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

// GetPointCurrent returns current conditions at a point. It needs an API key.
func (c *Client) GetPointCurrent(ctx context.Context, at Point, params *CurrentParams) (*CurrentResponse, error) {
	q := c.currentQuery(params)
	q.num("lat", at.Lat)
	q.num("lon", at.Lon)
	q.numPtr("elevation_m", at.ElevationM)
	var r CurrentResponse
	if err := c.get(ctx, opPointCurrent, []string{"current"}, q, &r); err != nil {
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

// GetPointAlerts returns the weather alerts in effect at a point. It needs an API key.
func (c *Client) GetPointAlerts(ctx context.Context, at Point) (*AlertsResponse, error) {
	if at.ElevationM != nil {
		return nil, usageError(opPointAlerts.id, "alerts take no elevation")
	}
	q := &query{}
	q.num("lat", at.Lat)
	q.num("lon", at.Lon)
	var r AlertsResponse
	if err := c.get(ctx, opPointAlerts, []string{"alerts"}, q, &r); err != nil {
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
	q.numDefault("radius_km", p.RadiusKm, 50)
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

// GetPointNearby returns the locations near a point.
func (c *Client) GetPointNearby(ctx context.Context, at Point, params *NearbyParams) (*NearbyPointResponse, error) {
	if at.ElevationM != nil {
		return nil, usageError(opPointNearby.id, "nearby takes no elevation")
	}
	q := c.nearbyQuery(params)
	q.num("lat", at.Lat)
	q.num("lon", at.Lon)
	var r NearbyPointResponse
	if err := c.get(ctx, opPointNearby, []string{"nearby"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SearchParams are the filters and order of a location search.
type SearchParams struct {
	// State filters by two-letter code or state name.
	State string
	// County filters by county name.
	County string
	// Type filters by location type.
	Type LocationType
	// Limit is the maximum number of hits (1..100; default 20).
	Limit int
	// Sort orders the hits by an explore metric instead of by relevance (the
	// default, also the zero value); each hit then carries Explore.
	Sort ExploreMetric
	// Order is the metric order (default [OrderDesc]).
	Order Order
	// Near limits the hits to RadiusKm around a point (no elevation).
	Near *Point
	// RadiusKm is Near's radius (up to 1000; default 100).
	RadiusKm float64
	// Units overrides the client's unit system for the hits' values.
	Units Units
}

// Search finds locations by free text (at least 2 characters). It is sent
// without the API key, so results are shared at the edge cache.
func (c *Client) Search(ctx context.Context, text string, params *SearchParams) (*SearchLocationsResponse, error) {
	if params == nil {
		params = &SearchParams{}
	}
	q := &query{}
	q.set("q", text)
	q.str("state", normalizeState(params.State))
	q.str("county", params.County)
	q.str("type", string(params.Type))
	q.intDefault("limit", params.Limit, 20)
	q.strDefault("sort", string(params.Sort), "relevance")
	q.strDefault("order", string(params.Order), string(OrderDesc))
	if err := q.near(opSearch.id, params.Near); err != nil {
		return nil, err
	}
	q.numDefault("radius_km", params.RadiusKm, 100)
	q.units(c.units(params.Units))
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
	q.list("network", networks)
	q.list("capability", params.Capabilities)
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

// ExploreParams are the filters of a ranking. The zero value ranks every
// place's primary point, largest first, 25 to a page, in the client's units.
type ExploreParams struct {
	// Order is the direction (default [OrderDesc], the largest first).
	Order Order
	// Type filters by location type.
	Type LocationType
	// State filters by two-letter code (lowercased) or state name.
	State string
	// Near limits the ranking to RadiusKm around a point (no elevation).
	Near *Point
	// RadiusKm is Near's radius (up to 1000; default 100).
	RadiusKm float64
	// ElevationMin and ElevationMax bound the ranked points' elevation, in
	// the request's length units.
	ElevationMin, ElevationMax *float64
	// AllPoints ranks every named elevation of a place, not only its
	// primary point (points=all).
	AllPoints bool
	// Limit is the page size (1..100; default 25).
	Limit int
	// Cursor is the previous page's NextCursor.
	Cursor string
	// Units overrides the client's unit system for this request.
	Units Units
}

// Explore returns one page of places ranked by metric (snow, precipitation,
// snow depth or season totals). Follow NextCursor for the next page.
func (c *Client) Explore(ctx context.Context, metric ExploreMetric, params *ExploreParams) (*ExploreResponse, error) {
	if metric == "" {
		return nil, usageError(opExplore.id, "metric is required, for example ExploreSnow5d")
	}
	if params == nil {
		params = &ExploreParams{}
	}
	q := &query{}
	q.set("metric", string(metric))
	q.strDefault("order", string(params.Order), string(OrderDesc))
	q.str("type", string(params.Type))
	q.str("state", normalizeState(params.State))
	if err := q.near(opExplore.id, params.Near); err != nil {
		return nil, err
	}
	q.numDefault("radius_km", params.RadiusKm, 100)
	q.numPtr("elevation_min", params.ElevationMin)
	q.numPtr("elevation_max", params.ElevationMax)
	if params.AllPoints {
		q.set("points", "all")
	}
	q.intDefault("limit", params.Limit, 25)
	q.str("cursor", params.Cursor)
	q.units(c.units(params.Units))
	var r ExploreResponse
	if err := c.get(ctx, opExplore, []string{"explore"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Composition returns the blend's served models and member counts. It is
// sent without the API key, so the response is shared at the edge cache.
func (c *Client) Composition(ctx context.Context) (*CompositionResponse, error) {
	var r CompositionResponse
	if err := c.get(ctx, opComposition, []string{"composition"}, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SeasonParams are the options of a location's season to date.
type SeasonParams struct {
	// WaterYear is the water year by its ending year (2027 and later); 0
	// means the current one.
	WaterYear int
	// Include opts in to extras: "daily" (SeasonResponse.Daily).
	Include []string
	// Units overrides the client's unit system for this request.
	Units Units
}

// Season returns a named location's observed season-to-date precipitation
// and snowfall.
func (c *Client) Season(ctx context.Context, location string, params *SeasonParams) (*SeasonResponse, error) {
	segs, err := splitLocation(opSeason.id, location)
	if err != nil {
		return nil, err
	}
	if params == nil {
		params = &SeasonParams{}
	}
	q := &query{}
	q.intDefault("water_year", params.WaterYear, 0)
	q.list("include", params.Include)
	q.units(c.units(params.Units))
	var r SeasonResponse
	if err := c.get(ctx, opSeason, append(segs, "season"), q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SnowTrendParams are the options of a location's snow trend.
type SnowTrendParams struct {
	// Elevation is a named elevation of the location, such as "base"; empty
	// means its primary point.
	Elevation string
	// Units overrides the client's unit system for this request.
	Units Units
}

// SnowTrend returns each recent run's daily snowfall at a named location,
// by source. It needs a Plus or Pro API key: anonymous and Free callers get
// a [CategoryPermission] error with code [CodePlusRequired].
func (c *Client) SnowTrend(ctx context.Context, location string, params *SnowTrendParams) (*SnowTrendResponse, error) {
	segs, err := splitLocation(opSnowTrend.id, location)
	if err != nil {
		return nil, err
	}
	if params == nil {
		params = &SnowTrendParams{}
	}
	q := &query{}
	q.str("elevation", params.Elevation)
	q.units(c.units(params.Units))
	var r SnowTrendResponse
	if err := c.get(ctx, opSnowTrend, append(segs, "snow-trend"), q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SnowpackParams are the options of a location's snowpack forecast and
// snow-depth plume.
type SnowpackParams struct {
	// Units overrides the client's unit system for this request.
	Units Units
}

// Snowpack returns the snowpack forecast of a named location's spots. It is
// a beta: the body carries Beta true, and fields may be added.
func (c *Client) Snowpack(ctx context.Context, location string, params *SnowpackParams) (*SnowpackResponse, error) {
	segs, err := splitLocation(opSnowpack.id, location)
	if err != nil {
		return nil, err
	}
	var r SnowpackResponse
	if err := c.get(ctx, opSnowpack, append(segs, "snowpack"), c.snowpackQuery(params), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SnowpackPlume returns a named location's snow-depth plume. It is a beta:
// the body carries Beta true, and fields may be added.
func (c *Client) SnowpackPlume(ctx context.Context, location string, params *SnowpackParams) (*SnowpackPlumeResponse, error) {
	segs, err := splitLocation(opSnowpackPlume.id, location)
	if err != nil {
		return nil, err
	}
	var r SnowpackPlumeResponse
	if err := c.get(ctx, opSnowpackPlume, append(segs, "snowpack", "plume"), c.snowpackQuery(params), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) snowpackQuery(p *SnowpackParams) *query {
	q := &query{}
	var u Units
	if p != nil {
		u = p.Units
	}
	q.units(c.units(u))
	return q
}

// BBox is a bounding box in decimal degrees (WGS84, west negative).
type BBox struct {
	West, South, East, North float64
}

// AlertOutlinesParams filter the alert outlines.
type AlertOutlinesParams struct {
	// Classes keeps the outlines of these classes only; empty keeps all.
	Classes []AlertClass
	// BBox keeps the outlines that intersect a box; nil keeps all.
	BBox *BBox
}

// AlertOutlines returns the outlines of the active NWS winter and wind
// alerts as a GeoJSON FeatureCollection. Its keys are the API's, untouched,
// so it can go straight to a map library; geometries are as served.
func (c *Client) AlertOutlines(ctx context.Context, params *AlertOutlinesParams) (*AlertOutlineCollection, error) {
	q := &query{}
	if params != nil {
		classes := make([]string, len(params.Classes))
		for i, k := range params.Classes {
			classes[i] = string(k)
		}
		q.list("class", classes)
		if b := params.BBox; b != nil {
			if err := checkBBox(opAlertOutlines.id, *b); err != nil {
				return nil, err
			}
			q.numbers("bbox", b.West, b.South, b.East, b.North)
		}
	}
	var r AlertOutlineCollection
	if err := c.get(ctx, opAlertOutlines, []string{"alerts", "outlines"}, q, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// AvalancheZones returns the avalanche centers' forecast zones, each with
// today's published danger rating, as a GeoJSON FeatureCollection. Its keys
// are the API's, untouched (Properties.ForecastURL stays forecast_url).
func (c *Client) AvalancheZones(ctx context.Context) (*AvalancheZoneCollection, error) {
	var r AvalancheZoneCollection
	if err := c.get(ctx, opAvalancheZones, []string{"avalanche", "zones"}, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func finite(vs ...float64) bool {
	for _, v := range vs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

func checkLatLon(lat, lon float64) bool {
	return finite(lat, lon) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

func checkBBox(op string, b BBox) error {
	switch {
	case !checkLatLon(b.South, b.West) || !checkLatLon(b.North, b.East):
		return usageError(op, "bbox %v is outside -180..180 longitude, -90..90 latitude", b)
	case b.West > b.East:
		return usageError(op, "bbox west %v is east of east %v", b.West, b.East)
	case b.South > b.North:
		return usageError(op, "bbox south %v is north of north %v", b.South, b.North)
	}
	return nil
}
