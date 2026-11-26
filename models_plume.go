// SPDX-License-Identifier: BSD-2-Clause

// The plume view (Members: MembersAsPlume) and the forecast's added shapes.

package ensembleweather

import (
	"time"
)

// ForecastLocation is the forecast's `location` -- the one `LocationRef` (every field where it
// was) plus what a forecast client labels and picks with, so the forecast body alone is enough to
// draw a plume or a meteogram at a named elevation (no `/nearby` call, no client-side unit
// conversion).
type ForecastLocation struct {
	LocationRef
	// `elevation_m` in the request's length units (m for `metric`/`kelvin`, ft for `imperial`), to
	// 0.1 -- the elevation `forecast.timeseries` (and its members) is served at.
	Elevation float64 `json:"elevation"`
	// Elevations is the location's named elevations -- exactly the list its `/nearby` and list rows
	// serve (`LocationSummary.elevations`): the primary first and its ONE `default: true`, then each
	// other elevation (ski-area base / mid / summit), in the request's length units.
	Elevations []ElevationSummary `json:"elevations,omitempty"`
}

// ElevationSummary is one of a location's elevation points, as `/nearby`, the list rows and the
// forecast's `location.elevations` serve it.
type ElevationSummary struct {
	// Name is the elevation's id: the name to pass in `?elevations=` (case, `-`, `_` and spaces
	// ignored).
	Name string `json:"name"`
	// What the place is called, for display ("K-1 Lodge", "FourRunner Quad top").
	DisplayName *string `json:"display_name,omitempty"`
	// Height (JSON `elevation`): In the caller's length units (m, ft), to 0.1.
	Height float64 `json:"elevation"`
	// Present (true) on the primary only.
	Default *bool `json:"default,omitempty"`
}

// MemberQuantiles is a distribution over the ensemble MEMBERS (`members_cumulative`,
// `members_spread`, `members_6h`): the `Quantiles` shape.
type MemberQuantiles struct {
	Mean float64 `json:"mean"`
	P10  float64 `json:"p10"`
	P25  float64 `json:"p25"`
	// P50 is the members' median.
	P50 *float64 `json:"p50,omitempty"`
	P75 float64  `json:"p75"`
	P90 float64  `json:"p90"`
	Min float64  `json:"min"`
	Max float64  `json:"max"`
	// As `Quantiles.probability`, where the use says.
	Probability *float64 `json:"probability,omitempty"`
}

// MembersSpread is the members' spread at this hour for the plume's instantaneous variables -- the
// Quantiles of every member's own `temperature` (temperature unit) and `snow_level` (length unit)
// here, over the members as served.
type MembersSpread struct {
	Temperature *MemberQuantiles `json:"temperature,omitempty"`
	SnowLevel   *MemberQuantiles `json:"snow_level,omitempty"`
	// SnowLevelWetBulb is the Quantiles of the members' own `snow_level_wet_bulb` (the snow level on
	// the NWS scale), over the members `snow_level` is over.
	SnowLevelWetBulb *MemberQuantiles `json:"snow_level_wet_bulb,omitempty"`
	// SLR is the snow-to-liquid ratio over the members SNOWING at this hour (their own `snowfall`
	// above 0), at 00/06/12/18Z only, in both member views.
	SLR *MemberQuantiles `json:"slr,omitempty"`
	// WindSpeed is the members' own wind speed at this hour, in the wind unit.
	WindSpeed *MemberQuantiles `json:"wind_speed,omitempty"`
}

// PrecipitationByType is this hour's headline precipitation (`ensemble.precipitation.mean`) split by
// phase, in the precipitation units -- `rain + snow + rain_snow_mix + freezing_rain` is the hour's
// mean (to the served precision), so a stacked bar is the headline bar.
type PrecipitationByType struct {
	Rain float64 `json:"rain"`
	// Snow is the liquid that falls as snow (not snow depth; that is `ensemble.snowfall`).
	Snow        float64 `json:"snow"`
	RainSnowMix float64 `json:"rain_snow_mix"`
	// -- liquid that freezes on contact (counted in `rain` before).
	FreezingRain float64 `json:"freezing_rain"`
}

// SnowLevelRange is a day's snow level over its wet hours (m, or ft with `units=imperial`): the
// lowest and highest hourly ensemble-mean snow level among them, and their precipitation-weighted
// mean.
type SnowLevelRange struct {
	// Low is the lowest hourly snow level among the wet hours.
	Low float64 `json:"low"`
	// High is the highest hourly snow level among the wet hours.
	High float64 `json:"high"`
	// Mean is the wet hours' snow level, weighted by their precipitation.
	Mean float64 `json:"mean"`
}

// DayNormal is the day's climate normal at the forecast point: the mean, over the day of year
// +/-15 days of every year of the station's record, of the very quantities this row states -- the
// local CALENDAR date's maximum hourly temperature (`high`, the quantity `high` is), its minimum
// (`low`) and its precipitation TOTAL (`precipitation`, dry days included; a daily total, never an
// hourly amount).
type DayNormal struct {
	// Mean local-calendar-day maximum temperature.
	High *float64 `json:"high,omitempty"`
	// Mean local-calendar-day minimum temperature.
	Low *float64 `json:"low,omitempty"`
	// Mean local-calendar-day precipitation total, dry days included.
	Precipitation *float64 `json:"precipitation,omitempty"`
	// Source is a climatology station's own record, or the ERA5 cell of the point.
	Source string `json:"source"`
	// Great-circle distance from the point to the station (or the cell centre), km.
	DistanceKm float64 `json:"distance_km"`
}

// DayDeparture is the served headline minus the normal, in the request's units: `high.p50 -
// normal.high` and `low.p50 - normal.low`.
type DayDeparture struct {
	High *float64 `json:"high,omitempty"`
	Low  *float64 `json:"low,omitempty"`
}

// DayPercentile is how unusual the day is: where the served headline (`p50`) falls in the day's
// climatological distribution, as a percentile 0-100, `Pr{climate <= forecast}` -- 95 for
// `precipitation` reads "wetter than 95% of this date's days".
type DayPercentile struct {
	High          *float64 `json:"high,omitempty"`
	Low           *float64 `json:"low,omitempty"`
	Precipitation *float64 `json:"precipitation,omitempty"`
}

// PrecipOutlook holds rolling liquid-equivalent totals over the same windows as `snow_outlook`, for
// places where rain is the story (the favorites cards' "0.6 in rain" with its range).
type PrecipOutlook struct {
	// Sum of the hourly precipitation means over the 24 h from the first served hour; `null` when not
	// covered.
	Next24h *float64 `json:"next_24h"`
	// Next48h is the same over 48 h.
	Next48h *float64 `json:"next_48h"`
	// Next72h is the same over 72 h.
	Next72h *float64 `json:"next_72h"`
	// Next7d is the same over 168 h.
	Next7d    *float64              `json:"next_7d"`
	Quantiles *SnowOutlookQuantiles `json:"quantiles,omitempty"`
}

// SnowOnGround is what the snowpack model states is on the ground at a spot at one hour.
type SnowOnGround struct {
	// Always `true` while the product is a beta; fields may be added.
	Beta bool `json:"beta"`
	// Snow depth: cm, or in (imperial).
	Depth float64 `json:"depth"`
	// `measured`: the run started from a snow-depth sensor at the spot.
	Basis string `json:"basis"`
	// Surface is the surface class at the hour; `/snowpack`'s `definitions.surface` carries the text
	// of each value (an open set).
	Surface *string `json:"surface,omitempty"`
	// ValidTime is the hour the depth is for.
	ValidTime time.Time `json:"valid_time"`
	// When the snowpack run that produced it finished.
	GeneratedAt time.Time `json:"generated_at"`
}

// AvalancheForecast is the avalanche center's danger rating for the forecast zone the location's
// point is in, from avalanche.org's public map layer.
type AvalancheForecast struct {
	// CenterID is the avalanche center's short id.
	CenterID   string `json:"center_id"`
	CenterName string `json:"center_name"`
	// avalanche.org's id for the forecast zone; for a CAIC forecast area, an id at or above 2^52
	// derived from CAIC's area id (stable for that grouping of CAIC's polygons).
	ZoneID   int64  `json:"zone_id"`
	ZoneName string `json:"zone_name"`
	// DangerLevel is the North American Public Avalanche Danger Scale level, 1-5; present only when
	// `status` is `rated`.
	DangerLevel *int `json:"danger_level"`
	// Display text: the scale's English name for `danger_level` (Low, Moderate, Considerable, High,
	// Extreme), derived from the level, never the center's own wording; `null` exactly when
	// `danger_level` is.
	DangerLabel *string `json:"danger_label"`
	// `rated`: a current rating.
	Status string `json:"status"`
	// When the center's current product took effect (UTC), when stated.
	ValidFrom *time.Time `json:"valid_from"`
	// When the center's current product expires (UTC), when stated.
	ValidUntil *time.Time `json:"valid_until"`
	// URL (JSON `forecast_url`) is the center's full forecast for the zone.
	URL       string `json:"forecast_url"`
	CenterURL string `json:"center_url"`
	// Attribution is the credit to show with the block: "Avalanche forecast by <center> via
	// avalanche.org", or "Avalanche forecast by Colorado Avalanche Information Center via
	// avalanche.state.co.us" for a CAIC area.
	Attribution string `json:"attribution"`
}

// MemberPlume is one served series' members in columns (`members=plume`).
type MemberPlume struct {
	Headline PlumeHeadline `json:"headline"`
	Bands    PlumeBands    `json:"bands"`
	// One entry per model with members in the served window, in the order aifs, aigefs, gefs,
	// ifs_ens, rrfs.
	Models []ModelColumns `json:"models"`
	// Pool (JSON `members`).
	Pool *PooledColumns `json:"members,omitempty"`
}

// PlumeHeadline is the lines a plume draws over its members, at every served hour (`valid_time`):
// the blend's `temperature` and `snow_level` means (`ensemble.*.mean`) and its running totals
// (`cumulative_precipitation`, `cumulative_snowfall`: the "Ourea forecast" line), and that
// snowfall line at a 10:1 snow-to-liquid ratio (`cumulative_snowfall_10to1`), and its `wind_speed`
// mean and its snowfall in each 6 hours ...
type PlumeHeadline struct {
	ValidTime   []time.Time `json:"valid_time"`
	Temperature []*float64  `json:"temperature"`
	SnowLevel   []*float64  `json:"snow_level"`
	// SnowLevelWetBulb is the blend's snow level on the NWS scale
	// (`ensemble.snow_level_wet_bulb.mean`, the height of the 0.5 C wet-bulb); `snow_level` is the
	// rain line.
	SnowLevelWetBulb        []*float64 `json:"snow_level_wet_bulb"`
	CumulativePrecipitation []*float64 `json:"cumulative_precipitation"`
	// CumulativeSnowfall is the hourly series' `cumulative_snowfall`: the running sum of the day
	// cards' snowfall (`daily[].precipitation_display.snow_amount`), spread inside each local date by
	// the blend's hourly mean snowfall, so at the end of each date served whole it equals the running
	// sum of the cards.
	CumulativeSnowfall []*float64 `json:"cumulative_snowfall"`
	// CumulativeSnowfall10to1 is the headline's running snowfall at a fixed 10:1 snow-to-liquid
	// ratio, in the snowfall units (in / mm), at the amount precision: each hour's served snowfall
	// with its SLR (`ensemble.slr.mean`) replaced by 10 -- the liquid that falls as snow, times 10
	// (0.1 in of it is 1 in; 1 mm is 1 cm) -- summed ...
	CumulativeSnowfall10to1 []*float64 `json:"cumulative_snowfall_10to1"`
	// WindSpeed is the forecast's hourly wind speed, the hourly series' `ensemble.wind_speed.mean`,
	// in the wind unit (mph / km/h / m/s as `units` says) at the wind precision (0.1), at every
	// served hour.
	WindSpeed []*float64 `json:"wind_speed"`
	// Snowfall6h is the forecast's snowfall in the 6 hours ending at each 00/06/12/18Z hour, in the
	// snowfall units (in / mm) at the amount precision: the rise of `cumulative_snowfall` over
	// `(valid_time - 6 h, valid_time]`, computed here so no client differences the line.
	Snowfall6h []*float64 `json:"snowfall_6h"`
}

// PlumeBands is the members' distributions, at the served hours that carry at least one
// (`valid_time`; the synoptic hours in practice): `cumulative_precipitation` /
// `cumulative_snowfall` are the row form's `members_cumulative` (the spread of the members'
// running totals, never a sum of hourly percentiles), `temperature` / `snow_level` its
// `members_spread`, `precipitation_6h` / `snowfall_6h` its `members_6h` ...
type PlumeBands struct {
	ValidTime               []time.Time `json:"valid_time"`
	CumulativePrecipitation BandColumns `json:"cumulative_precipitation"`
	CumulativeSnowfall      BandColumns `json:"cumulative_snowfall"`
	// CumulativeSnowfall10to1 is the members' running snowfall at 10:1 -- each member's snowfall with
	// its SLR replaced by 10 (its snowing liquid x 10, every phase gate of its `snowfall` kept, the
	// `snowfall_total_10to1` map's member rule), totalled as `cumulative_snowfall` is -- by
	// `cumulative_snowfall`'s rule: one fixed set of ...
	CumulativeSnowfall10to1 *BandColumns `json:"cumulative_snowfall_10to1"`
	Temperature             BandColumns  `json:"temperature"`
	SnowLevel               BandColumns  `json:"snow_level"`
	// SnowLevelWetBulb is the row form's `members_spread.snow_level_wet_bulb` -- the members' snow
	// level on the NWS scale, at the hours `snow_level` carries a band.
	SnowLevelWetBulb *BandColumns `json:"snow_level_wet_bulb"`
	Precipitation6h  BandColumns  `json:"precipitation_6h"`
	Snowfall6h       BandColumns  `json:"snowfall_6h"`
	SLR              *BandColumns `json:"slr,omitempty"`
	// `members_spread.wind_speed`.
	WindSpeed *BandColumns `json:"wind_speed,omitempty"`
}

// BandColumns is one band's statistics (Hyndman-Fan type 7 over the served member values), each an
// array aligned to `PlumeBands.valid_time`, `null` where the band is absent at that hour.
type BandColumns struct {
	Mean []*float64 `json:"mean"`
	Min  []*float64 `json:"min"`
	P10  []*float64 `json:"p10"`
	P25  []*float64 `json:"p25"`
	P50  []*float64 `json:"p50,omitempty"`
	P75  []*float64 `json:"p75"`
	P90  []*float64 `json:"p90"`
	Max  []*float64 `json:"max"`
}

// PooledColumns is the members as ONE unlabelled pool, for a caller without the plume's paid
// detail (Free and anonymous; `members_view: pool`).
type PooledColumns struct {
	ValidTime     []time.Time  `json:"valid_time"`
	Temperature   [][]*float64 `json:"temperature"`
	Precipitation [][]*float64 `json:"precipitation"`
	// `[member][step]`: the hours each `precipitation` and `snowfall` value covers: the grid step
	// (6), less where a member's first value covers less; `null` where the value is absent.
	PrecipitationHours      [][]*int     `json:"precipitation_hours"`
	CumulativePrecipitation [][]*float64 `json:"cumulative_precipitation"`
	SnowLevel               [][]*float64 `json:"snow_level"`
	SnowLevelWetBulb        [][]*float64 `json:"snow_level_wet_bulb"`
	Snowfall                [][]*float64 `json:"snowfall"`
	CumulativeSnowfall      [][]*float64 `json:"cumulative_snowfall"`
	SLR                     [][]*float64 `json:"slr"`
	// CumulativePrecipitationBandMembers is the rows (indices into every column) in
	// `bands.cumulative_precipitation`'s fixed population, ascending: that band's `mean` is the mean
	// of these rows' running totals.
	CumulativePrecipitationBandMembers []int `json:"cumulative_precipitation_band_members"`
	// Likewise for `bands.cumulative_snowfall`.
	CumulativeSnowfallBandMembers []int `json:"cumulative_snowfall_band_members"`
	// Likewise for `bands.cumulative_snowfall_10to1`.
	CumulativeSnowfall10to1BandMembers []int `json:"cumulative_snowfall_10to1_band_members"`
}

// ModelColumns is one model's members over its NATIVE steps inside the served window
// (`valid_time`: the hours at which a member of the model carries a value -- every hour for RRFS,
// every sixth for the global ensembles; never padded to the hourly series).
type ModelColumns struct {
	ID        string      `json:"id"`
	ValidTime []time.Time `json:"valid_time"`
	// MemberIDs is the members, ascending (`MemberForecast.member_id`).
	MemberIDs     []int        `json:"member_ids"`
	Temperature   [][]*float64 `json:"temperature"`
	Precipitation [][]*float64 `json:"precipitation"`
	// `[member][step]`: the hours each `precipitation` value covers; `null` where it is absent or its
	// span unknown.
	PrecipitationHours      [][]*int     `json:"precipitation_hours"`
	CumulativePrecipitation [][]*float64 `json:"cumulative_precipitation"`
	SnowLevel               [][]*float64 `json:"snow_level"`
	// `[Member][step]`, each member's `MemberForecast.snow_level_wet_bulb` (the snow level on the NWS
	// scale); `null` exactly where its `snow_level` is.
	SnowLevelWetBulb   [][]*float64 `json:"snow_level_wet_bulb"`
	Snowfall           [][]*float64 `json:"snowfall"`
	CumulativeSnowfall [][]*float64 `json:"cumulative_snowfall"`
	// One value per `valid_time` step (not per member): the median of the model's members'
	// `cumulative_snowfall` there (Hyndman-Fan type 7, as `bands`: the middle value, or the midpoint
	// of the two middle ones; a single-member model's own total), at the amount precision.
	CumulativeSnowfallP50 []*float64 `json:"cumulative_snowfall_p50"`
	// As `cumulative_snowfall_p50`, over the members' running snowfall at 10:1 (see
	// `PlumeBands.cumulative_snowfall_10to1`).
	CumulativeSnowfall10to1P50 []*float64 `json:"cumulative_snowfall_10to1_p50"`
	// TemperatureMean is the model's mean lines.
	TemperatureMean []*float64 `json:"temperature_mean"`
	// PrecipitationMean is the mean of `precipitation` (see `temperature_mean`).
	PrecipitationMean []*float64 `json:"precipitation_mean"`
	// CumulativePrecipitationMean is the mean of `cumulative_precipitation` (see `temperature_mean`).
	CumulativePrecipitationMean []*float64 `json:"cumulative_precipitation_mean"`
	// SnowLevelMean is the mean of `snow_level` (see `temperature_mean`).
	SnowLevelMean []*float64 `json:"snow_level_mean"`
	// SnowLevelWetBulbMean is the mean of `snow_level_wet_bulb` (see `temperature_mean`).
	SnowLevelWetBulbMean []*float64 `json:"snow_level_wet_bulb_mean"`
	// SnowfallMean is the mean of `snowfall` (see `temperature_mean`).
	SnowfallMean []*float64 `json:"snowfall_mean"`
	// CumulativeSnowfallMean is the mean of `cumulative_snowfall` (see `temperature_mean`), beside
	// its median `cumulative_snowfall_p50`.
	CumulativeSnowfallMean []*float64 `json:"cumulative_snowfall_mean"`
	// CumulativeSnowfall10to1Mean is the mean of the members' running snowfall at 10:1, beside
	// `cumulative_snowfall_10to1_p50` (the members' 10:1 column itself is not served).
	CumulativeSnowfall10to1Mean []*float64 `json:"cumulative_snowfall_10to1_mean"`
	// WindSpeedMean is the mean of `wind_speed` (see `temperature_mean`).
	WindSpeedMean []*float64 `json:"wind_speed_mean,omitempty"`
	// Which of this model's `member_ids` are in `bands.cumulative_precipitation`, ascending.
	CumulativePrecipitationBandMemberIDs []int `json:"cumulative_precipitation_band_member_ids"`
	// As `cumulative_precipitation_band_member_ids`, for `bands.cumulative_snowfall`.
	CumulativeSnowfallBandMemberIDs []int `json:"cumulative_snowfall_band_member_ids"`
	// As `cumulative_precipitation_band_member_ids`, for `bands.cumulative_snowfall_10to1`.
	CumulativeSnowfall10to1BandMemberIDs []int `json:"cumulative_snowfall_10to1_band_member_ids"`
	// Pro (`members_view: full`) only; absent in the plume view.
	WindSpeed [][]*float64 `json:"wind_speed,omitempty"`
	// Pro only; absent in the plume view.
	WindDirection [][]*float64 `json:"wind_direction,omitempty"`
	// In both views.
	SLR [][]*float64 `json:"slr,omitempty"`
}
