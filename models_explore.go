// SPDX-License-Identifier: BSD-2-Clause

// Rankings, the blend's composition, the season to date and the snow trend.

package ensembleweather

import (
	"time"
)

// ExploreResponse is one page of a ranking.
type ExploreResponse struct {
	ResponseMeta `json:"-"`

	Metric ExploreMetric `json:"metric"`
	Order  string        `json:"order"`
	Units  Units         `json:"units"`
	// Unit is the unit of every `value`, `p10` and `p90`.
	Unit  string        `json:"unit"`
	Items []ExploreItem `json:"items"`
	// Items on this page.
	Count int `json:"count"`
	// Pass as `?cursor=` (same other parameters) for the next page; null at the end.
	NextCursor *string `json:"next_cursor"`
}

// ExploreItem is one ranked point of a place.
type ExploreItem struct {
	// 1-based, across pages.
	Rank     int         `json:"rank"`
	Location LocationRef `json:"location"`
	// ElevationName is the named elevation; `null` = the place's own point.
	ElevationName *string `json:"elevation_name"`
	// What the elevation `elevation_name` is called, as `ElevationSummary.display_name`; absent where
	// it has no name.
	ElevationDisplayName *string `json:"elevation_display_name,omitempty"`
	// Elevation is the point's elevation, m or ft by `units`.
	Elevation float64 `json:"elevation"`
	// Value is the metric's value as the place's page serves it, in `unit`.
	Value float64 `json:"value"`
	// Present only on `snow_today` (the day's served `total_snowfall.p10`, where it has a
	// `snow_amount`) and `precip_today` (`total_precipitation_if_wet.p10`, where the day is not
	// `none`).
	P10 *float64 `json:"p10,omitempty"`
	// As `p10`.
	P90 *float64 `json:"p90,omitempty"`
	// When the value was taken: the forecast windows when recorded from the served run, `base_depth`
	// its hour, `new_snow_24h` the snowpack run's first hour, the season its newest computed day.
	AsOf time.Time `json:"as_of"`
	// Forecast windows only, the served run's `model_run.initialized_at`.
	InitializedAt *time.Time `json:"initialized_at"`
	Source        string     `json:"source"`
	// With `near` only, from it (0.1 km).
	DistanceKm *float64 `json:"distance_km,omitempty"`
	// Path is the place's forecast route; add `?elevations=<elevation_name>` for a named point.
	Path string `json:"path"`
}

// ExploreValue is a search hit's Explore value; the fields mean what an `ExploreItem`'s do.
type ExploreValue struct {
	Metric        ExploreMetric `json:"metric"`
	Value         float64       `json:"value"`
	Unit          string        `json:"unit"`
	P10           *float64      `json:"p10,omitempty"`
	P90           *float64      `json:"p90,omitempty"`
	AsOf          time.Time     `json:"as_of"`
	InitializedAt *time.Time    `json:"initialized_at"`
	Source        string        `json:"source"`
}

// CompositionResponse is the body of `GET /api/v1/composition`.
type CompositionResponse struct {
	ResponseMeta `json:"-"`

	// InitializedAt is the reference init: the newest contributing cycle, a forecast's
	// `model_run.initialized_at`.
	InitializedAt time.Time `json:"initialized_at"`
	// Each contributing model, its own cycle and its members (`members` absent only when the cycle's
	// member count is not known).
	Models []ModelEntry `json:"models"`
	// Entries in `models`.
	ModelCount int `json:"model_count"`
	// TotalMembers is the sum of `models[].members`: the members the blend draws on.
	TotalMembers int `json:"total_members"`
}

// SeasonResponse is a location's observed season-to-date precipitation and snowfall ([Client.Season]).
type SeasonResponse struct {
	ResponseMeta `json:"-"`

	Location LocationRef `json:"location"`
	// WaterYear is the water year by its ending year.
	WaterYear int `json:"water_year"`
	// Starts is the first local day of the water year.
	Starts string `json:"starts"`
	// Through is the last local day with an observed row (the day before `starts` when there is
	// none).
	Through string `json:"through"`
	// When the newest day row was computed; null with no rows.
	AsOf  *time.Time `json:"as_of"`
	Units Units      `json:"units"`
	// Null until the location has an observed day.
	Precipitation *SeasonTotal `json:"precipitation"`
	// Always null for now; season snowfall is not served yet.
	Snowfall *SeasonTotal `json:"snowfall"`
	// Station is the nearest SNOTEL or SCAN gauge within 25 km and 300 m with rows this water year,
	// or null.
	Station *SeasonStation `json:"station"`
	// Only with `include=daily`.
	Daily       []SeasonDay         `json:"daily,omitempty"`
	Attribution ResponseAttribution `json:"attribution"`
}

// SeasonTotal is a season-to-date total with its coverage and normal.
type SeasonTotal struct {
	// Amount (JSON `total`): Sum over the covered days, in `unit`.
	Amount float64 `json:"total"`
	Unit   string  `json:"unit"`
	// `mrms_multisensor`: the gridded analysis at the location; `station_gauge`: the location is a
	// gauge site and this is the gauge's accumulator.
	Source      string `json:"source"`
	DaysCovered int    `json:"days_covered"`
	// Days since `starts` with no complete observation; not counted as dry.
	DaysMissing int `json:"days_missing"`
	// Covered days still inside the 7-day restate window.
	ProvisionalDays int `json:"provisional_days"`
	// PRISM 1991-2020 over the covered days, in `unit`; null where no normal is loaded.
	NormalToDate    *float64 `json:"normal_to_date"`
	PercentOfNormal *int     `json:"percent_of_normal"`
	NormalSource    *string  `json:"normal_source"`
}

// SeasonStation is the nearest SNOTEL or SCAN gauge's own season to date.
type SeasonStation struct {
	// ID is the NRCS station triplet.
	ID         string  `json:"id"`
	Network    string  `json:"network"`
	Name       string  `json:"name"`
	DistanceKm float64 `json:"distance_km"`
	ElevationM float64 `json:"elevation_m"`
	// PrecipitationTotal is the gauge's water-year accumulator on `through` (mm or in).
	PrecipitationTotal           *float64 `json:"precipitation_total"`
	PrecipitationPercentOfMedian *int     `json:"precipitation_percent_of_median"`
	// Snow water equivalent on `through` (mm or in).
	SWE                *float64 `json:"swe"`
	SWEPercentOfMedian *int     `json:"swe_percent_of_median"`
	// Snow depth on `through` (cm or in).
	SnowDepth *float64 `json:"snow_depth"`
	// Through is the gauge's newest day.
	Through string `json:"through"`
}

// SeasonDay is one observed local day of the season.
type SeasonDay struct {
	Date string `json:"date"`
	// Precipitation is the day's total; null when an hour was not observed.
	Precipitation *float64 `json:"precipitation"`
	// Always null for now.
	Snowfall    *float64 `json:"snowfall"`
	Provisional bool     `json:"provisional"`
}

// SnowTrendResponse is each recent run's daily snowfall at a location's point, by source ([Client.SnowTrend]).
type SnowTrendResponse struct {
	ResponseMeta `json:"-"`

	Location LocationRef `json:"location"`
	Units    Units       `json:"units"`
	// Elevation is the point the grids are for.
	Elevation SnowTrendElevation `json:"elevation"`
	// How many days back runs are kept.
	RetentionDays int `json:"retention_days"`
	// Dates is the columns, location-local dates ascending.
	Dates []string `json:"dates"`
	// `ourea` first; with the `model_data` entitlement then `aifs`, `aigefs`, `gefs`, `ifs_ens` and
	// `rrfs`.
	Grids []SnowTrendGrid `json:"grids"`
}

// SnowTrendElevation is the point the grids are for.
type SnowTrendElevation struct {
	// Name is the named elevation; `null` for a location with one point.
	Name *string `json:"name"`
	// What that elevation is called, as `ElevationSummary.display_name`; absent where it has no name.
	DisplayName *string `json:"display_name,omitempty"`
	// Height (JSON `elevation`): Its elevation, in m (metric) or ft (imperial).
	Height float64 `json:"elevation"`
}

// SnowTrendGrid is one source's runs, oldest first.
type SnowTrendGrid struct {
	// `ourea`, or a model id.
	Source string `json:"source"`
	// Oldest first; the last is the newest.
	Runs []SnowTrendRun `json:"runs"`
}

// SnowTrendRun is one run's daily snowfall, one entry per date.
type SnowTrendRun struct {
	// `true`: re-rendered after the fact from the pinned run (`?run=<initialized_at>`) by the post-
	// deploy backfill -- today's calibration and no live nowcast, so it can differ slightly from what
	// was served then.
	Backfilled bool `json:"backfilled"`
	// `ourea`: the served body's `model_run.initialized_at`.
	InitializedAt time.Time `json:"initialized_at"`
	// One entry per `dates` entry, in order; `null` where the run has no value.
	Days []*SnowTrendDay `json:"days"`
}

// SnowTrendDay is one run's snowfall on one local day.
type SnowTrendDay struct {
	// Date is the `dates` entry this day is for.
	Date string `json:"date"`
	// Amount (JSON `snow`) is the day's snowfall in mm (metric) or inches (imperial), at the served
	// amount precision.
	Amount *float64 `json:"snow"`
	// Partial is the run covers only part of this local day.
	Partial bool `json:"partial"`
	// P10 is the range's low end.
	P10 *float64 `json:"p10,omitempty"`
	// P90 is the range's high end, as `p10`.
	P90 *float64 `json:"p90,omitempty"`
	// `ourea` only: the day's `precipitation_display.kind`.
	Kind *string `json:"kind,omitempty"`
}
