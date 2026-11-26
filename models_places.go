// SPDX-License-Identifier: BSD-2-Clause

// A place's summary and a mountain pass's forecast (on Current).

package ensembleweather

import (
	"time"
)

// PlaceSummary is a short plain-language summary of what is coming at a place, written by the
// server from templates filled with served values (never free text from a model).
type PlaceSummary struct {
	// One sentence: the most important thing at the place, by priority (an NWS alert, precipitation
	// starting or stopping soon, snow coming, wind or a large temperature change, then the day's
	// headline).
	Short string `json:"short"`
	// One to three sentences for the location page, each from a different kind of event.
	Long []string `json:"long"`
	// Events is the event behind each sentence of `long`, in the same order (`events[i]` wrote
	// `long[i]`), so a client can style a sentence (an alert chip, a snow icon) without parsing text.
	Events []SummaryEvent `json:"events"`
	// ElevationName is the served elevation the sentences describe: its display name where the
	// catalog has one, else its name.
	ElevationName *string `json:"elevation_name,omitempty"`
	// Elevation is the height the sentences describe, in the response's length units (m, ft).
	Elevation float64 `json:"elevation"`
	// Run is the init of the forecast run the sentences were written from.
	Run time.Time `json:"run"`
	// When the sentences were written (the request's now); the "updated" stamp.
	GeneratedAt time.Time `json:"generated_at"`
}

// SummaryEvent is one event a place summary is built from.
type SummaryEvent struct {
	// What the sentence is about.
	Kind string `json:"kind"`
	// Template is the phrasing used, `<template>.<variant>` (`snow_coming.v1.3`), stable across
	// releases so sentences can be counted, reviewed and later translated.
	Template string `json:"template"`
	// At is the hour the event begins (precipitation starting or stopping, an alert's onset, a pass
	// event).
	At *time.Time `json:"at,omitempty"`
	// When the event ends, where the sentence says so (an alert's end, a snow window's end).
	Until *time.Time `json:"until,omitempty"`
	// AlertID is the `alerts[].id` an `alert` event states.
	AlertID *string `json:"alert_id,omitempty"`
	// Values is the served numbers the sentence states, unrounded, in the response's units.
	Values *SummaryValues `json:"values,omitempty"`
}

// SummaryValues is each number a summary sentence states, as served.
type SummaryValues struct {
	// Snowfall is the window's headline snowfall total (mm, in): for `snow_coming` the forecast's
	// `snow_outlook.<snowfall_window>`; for `pass_snow_starts` the window's
	// `pass.snow_windows[].total`.
	Snowfall *float64 `json:"snowfall,omitempty"`
	// SnowfallP10 is the window's served p10 (`snow_outlook.quantiles.<window>.p10`, or the pass
	// window's `p10`).
	SnowfallP10 *float64 `json:"snowfall_p10,omitempty"`
	// SnowfallP90 is the window's served p90.
	SnowfallP90 *float64 `json:"snowfall_p90,omitempty"`
	// Which window `snowfall` totals.
	SnowfallWindow *string `json:"snowfall_window,omitempty"`
	// Temperature is a day's served high (`daily[].high.mean`) a `day_headline` or
	// `temperature_change` states.
	Temperature *float64 `json:"temperature,omitempty"`
	// WindGust is the served hourly gust mean (`ensemble.wind_gust.mean`) a `wind` sentence states.
	WindGust *float64 `json:"wind_gust,omitempty"`
	// PrecipitationProbability is the served chance (0-1) behind the sentence's likelihood word.
	PrecipitationProbability *float64 `json:"precipitation_probability,omitempty"`
}

// PassForecast answers "is the pass getting weather?" at the pass's own point and elevation: forecast
// quantities only, in the response's units.
type PassForecast struct {
	// Elevation is the pass's elevation, in the response's length units (m, ft).
	Elevation float64 `json:"elevation"`
	// ElevationM is the pass's elevation in metres, whatever `units` is.
	ElevationM float64 `json:"elevation_m"`
	// From is the first hour of `hours` (the hour containing now); the 24 h and 48 h windows open
	// here.
	From    time.Time   `json:"from"`
	Snow24h PassSnow24h `json:"snow_24h"`
	Chances PassChances `json:"chances"`
	// Each spell of snow at the pass in the next 48 h, in time order; empty when none.
	SnowWindows []PassSnowWindow `json:"snow_windows"`
	// Hours is the next 48 hours at the pass, from `from`, one per hour.
	Hours          []PassHour          `json:"hours"`
	Windows        PassTravelWindows   `json:"windows"`
	OfficialStatus *PassOfficialStatus `json:"official_status,omitempty"`
	// RoadSensors is the state DOT road-weather stations nearest the pass, nearest first: at most 3,
	// within 25 km, each with its newest reading in the last 3 h.
	RoadSensors []PassRoadSensor `json:"road_sensors,omitempty"`
	// Cameras is the DOT's own cameras at the pass, in the order the pass's curation lists them.
	Cameras []PassCamera `json:"cameras,omitempty"`
}

// PassSnow24h is the next 24 h of snow at the pass: exactly the default forecast's own
// `snow_outlook.next_24h` and `snow_outlook.quantiles.next_24h` (unpinned, the nowcast on, the
// same cycle), so the pass page, the passes board and search state one number.
type PassSnow24h struct {
	// `snow_outlook.next_24h` (mm, in); null where the forecast does not reach 24 h.
	Total *float64 `json:"total"`
	// `snow_outlook.quantiles.next_24h.p10`; null without members covering the window.
	P10 *float64 `json:"p10"`
	// `snow_outlook.quantiles.next_24h.p90`.
	P90 *float64 `json:"p90"`
	// StartsAt is the start of the first snow spell (`snow_windows[0].start`) when it begins inside
	// the 24 h; null otherwise.
	StartsAt *time.Time `json:"starts_at"`
}

// PassChances are the chances at the pass over the next `window_hours` hours, each the share of the
// forecast's members meeting the event (the `?prob=` machinery), except the gust chance (see its
// `method`).
type PassChances struct {
	// WindowHours is the window's length, 24.
	WindowHours int       `json:"window_hours"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	// Snowfall total over the window of 1 in (25.4 mm) or more: `prob=snowfall:ge:<1 in>:24h`.
	PSnowGe1in *PassChance `json:"p_snow_ge_1in"`
	// Snowfall total over the window of 6 in (152.4 mm) or more.
	PSnowGe6in *PassChance `json:"p_snow_ge_6in"`
	// Some hour of the window in which the member is at or below freezing (`threshold`: 0 C, 32 F)
	// with precipitation of at least `precipitation_threshold` (0.254 mm, 0.01 in) in the member's
	// value ending at that hour: a wet road at or below freezing.
	PWetBelowFreezing *PassChance `json:"p_wet_below_freezing"`
	// PGustGt40mph is a gust over 40 mph (17.9 m/s; threshold fixed, does not change with units) in
	// the window.
	PGustGt40mph *PassChance `json:"p_gust_gt_40mph"`
}

// PassChance is one chance at a pass: a probability with its threshold and members, never a bare number.
type PassChance struct {
	// 0-1 (0.001); null when it cannot be answered (`null_reason`).
	Probability *float64 `json:"probability"`
	// Threshold is the event's threshold, in the response's units.
	Threshold float64 `json:"threshold"`
	// For `p_wet_below_freezing`, the precipitation that makes an hour wet, in the response's units.
	PrecipitationThreshold *float64 `json:"precipitation_threshold,omitempty"`
	// Members is the members that voted (0 for `hourly_gust_distribution`).
	Members    int         `json:"members"`
	Method     string      `json:"method"`
	NullReason *NullReason `json:"null_reason,omitempty"`
}

// PassSnowWindow is a span of snowfall at the pass.
type PassSnowWindow struct {
	// When the first snowing hour begins (its `valid_time` minus 1 h).
	Start time.Time `json:"start"`
	// When the last snowing hour ends (its `valid_time`).
	End time.Time `json:"end"`
	// Total is the sum of the hourly snowfall means over the spell (mm, in), as `snow_outlook` sums
	// its windows.
	Total float64 `json:"total"`
	// P10 is the p10 of the members' own totals over the spell (`?totals=` machinery); omitted
	// without enough members.
	P10 *float64 `json:"p10,omitempty"`
	// P90 is the p90 of the members' own totals over the spell.
	P90 *float64 `json:"p90,omitempty"`
	// PGe1in is the share of members whose own total over the spell is 1 in (25.4 mm) or more;
	// threshold fixed, does not change with units.
	PGe1in *float64 `json:"p_ge_1in,omitempty"`
}

// PassHour is one hour at the pass's own point and elevation.
type PassHour struct {
	// Time is the hour's `valid_time`; amounts are over the hour ending then.
	Time time.Time `json:"time"`
	// `ensemble.temperature.mean`.
	Temperature float64 `json:"temperature"`
	// `ensemble.precipitation.mean` (mm, in).
	Precipitation float64 `json:"precipitation"`
	// `ensemble.precipitation.probability`, 0-1.
	PrecipitationProbability *float64 `json:"precipitation_probability,omitempty"`
	// PrecipitationType is the hour's served `precipitation_type.type`; omitted on a dry hour.
	PrecipitationType *string `json:"precipitation_type,omitempty"`
	// `ensemble.snowfall.mean` (mm, in); 0 without snow.
	Snowfall float64 `json:"snowfall"`
	// `ensemble.snow_level.mean` (m, ft); omitted where not served.
	SnowLevel *float64 `json:"snow_level,omitempty"`
	// SnowLevelMargin (JSON `snow_level_above_pass`): `snow_level` minus the pass's elevation (m,
	// ft): above 0 the snow level is above the pass (rain at the pass when wet), below 0 it is under
	// it.
	SnowLevelMargin *float64 `json:"snow_level_above_pass,omitempty"`
	// FreezingOnWetRoad is the hour is at or below freezing (temperature mean at or below 0 C, 32 F)
	// and wet: precipitation of at least 0.254 mm (0.01 in) in this hour or one of the two before it.
	FreezingOnWetRoad bool `json:"freezing_on_wet_road"`
	// `ensemble.wind_speed.mean`.
	WindSpeed float64 `json:"wind_speed"`
	// `ensemble.wind_gust.mean`; omitted where not served.
	WindGust *float64 `json:"wind_gust,omitempty"`
	// `wind_gust` is over 40 mph (17.9 m/s); threshold fixed, does not change with units.
	GustOver40mph bool `json:"gust_over_40mph"`
	// Gusts of 25 mph (11.2 m/s) or more at or below freezing with loose snow to move: snow falling
	// this hour, 1 in (25.4 mm) or more fallen in the 24 served hours before it, or snow on the
	// ground at the pass (`snow_on_ground`).
	BlowingSnow bool `json:"blowing_snow"`
}

// PassTravelWindows is the quietest and the busiest stretch of weather at the pass in the next
// `horizon_hours`, by one fixed rule on the hours above: an hour has weather when it is snowing
// (`snowfall` at least 0.04 in, 1 mm), `freezing_on_wet_road`, or `gust_over_40mph`.
type PassTravelWindows struct {
	// HorizonHours is the horizon's length in hours, 24.
	HorizonHours int `json:"horizon_hours"`
	// Null when every hour has weather.
	LeastWeather *PassTravelWindow `json:"least_weather"`
	// Null when no hour has weather.
	MostWeather *PassTravelWindow `json:"most_weather"`
}

// PassTravelWindow is a travel window and the weather in it.
type PassTravelWindow struct {
	// When the run's first hour begins.
	Start time.Time `json:"start"`
	// When its last hour ends.
	End time.Time `json:"end"`
	// Snowfall is the sum of the hourly snowfall means over the run (mm, in).
	Snowfall               float64 `json:"snowfall"`
	FreezingOnWetRoadHours int     `json:"freezing_on_wet_road_hours"`
	// MaxWindGust is the highest hourly gust mean in the run; omitted where no gust is served.
	MaxWindGust *float64 `json:"max_wind_gust,omitempty"`
}

// PassOfficialStatus is the DOT's own status for the road over the pass -- chain controls and
// closures as the DOT publishes them, its wording verbatim.
type PassOfficialStatus struct {
	// Source is the DOT feed.
	Source string `json:"source"`
	// Who to credit, as the client shows it.
	Attribution string `json:"attribution"`
	// URL is the DOT's own page for this road, to link out to.
	URL string `json:"url"`
	// How fresh the DOT's feed behind these segments was at Ourea's last read: the time the DOT wrote
	// the file Ourea last read (Caltrans rewrites it every minute; Ourea reads it every 5), and the
	// older of two when the segments span two DOT districts.
	CheckedAt time.Time `json:"checked_at"`
	// `checked_at` is more than 30 minutes before `as_of`.
	Stale bool            `json:"stale"`
	Level PassStatusLevel `json:"level"`
	// Segments is the DOT's chain-control and closure points on the road over the pass, each
	// direction, in the pass's curated order.
	Segments []PassStatusSegment `json:"segments"`
}

// PassStatusSegment is one DOT chain-control or closure point, as the DOT publishes it.
type PassStatusSegment struct {
	// ID is the DOT's own id for the point.
	ID    string `json:"id"`
	Route string `json:"route"`
	// Direction is the direction the point controls, as the DOT names it.
	Direction string `json:"direction"`
	// Location is the DOT's name for the point (its trailing field padding removed).
	Location string `json:"location"`
	// State (JSON `status`) is the DOT's status code, verbatim.
	State string          `json:"status"`
	Level PassStatusLevel `json:"level"`
	// Text (JSON `status_text`) is the DOT's own description of the status, verbatim.
	Text string `json:"status_text"`
	// When the DOT set this status.
	UpdatedAt time.Time `json:"updated_at"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	// RoadCondition is the DOT's own report of the road surface there, verbatim, where it gives one
	// beside the restriction (WSDOT, ODOT).
	RoadCondition *string `json:"road_condition,omitempty"`
	// WeatherCondition is the DOT's own report of the weather there, verbatim (WSDOT, ODOT).
	WeatherCondition *string `json:"weather_condition,omitempty"`
}

// PassRoadSensor is one state DOT road-weather station's newest reading, in the response's units.
type PassRoadSensor struct {
	// StationID is the station's id in its network.
	StationID string `json:"station_id"`
	// Name is the station's registry name, else its id.
	Name string `json:"name"`
	// Source is the feed: `rwis` (Caltrans CWWP2 RWIS) or `wsdot`.
	Source string `json:"source"`
	// Who measured it, as the client credits it.
	Attribution string `json:"attribution"`
	// Straight-line distance from the pass, km.
	DistanceKm float64 `json:"distance_km"`
	// Elevation is the station's registry elevation in the response's length units (m, ft).
	Elevation *float64 `json:"elevation,omitempty"`
	// ElevationM is the station's registry elevation in metres.
	ElevationM *float64  `json:"elevation_m,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
	// `observed_at` is more than 60 minutes before `as_of`.
	Stale bool `json:"stale"`
	// PavementCondition is the pavement state the sensor reports (NTCIP 1204 essSurfaceStatus as
	// `/current` sensors serve it: dry, trace_moisture, wet, chemically_wet, ice_warning, ice_watch,
	// snow_warning, snow_watch, absorption, dew, frost, absorption_at_dewpoint).
	PavementCondition *string `json:"pavement_condition,omitempty"`
	// Pavement surface temperature: °C, °F (imperial) or K (kelvin).
	PavementTemp *float64 `json:"pavement_temp,omitempty"`
	// As `pavement_temp`.
	SubsurfaceTemp *float64 `json:"subsurface_temp,omitempty"`
	// Air temperature, as `pavement_temp`.
	Temperature *float64 `json:"temperature,omitempty"`
	// m/s or mph (imperial).
	WindSpeed *float64 `json:"wind_speed,omitempty"`
	// As `wind_speed`.
	WindGust *float64 `json:"wind_gust,omitempty"`
	// Degrees clockwise from north.
	WindDirection *float64 `json:"wind_direction,omitempty"`
	// m, or statute miles (imperial).
	Visibility *float64 `json:"visibility,omitempty"`
}

// PassCamera is one DOT camera, a still image the DOT refreshes about once a minute.
type PassCamera struct {
	ID string `json:"id"`
	// Title is the DOT's name for the camera.
	Title string `json:"title"`
	// ImageURL is the DOT's current still image.
	ImageURL string `json:"image_url"`
	// URL is the DOT's own page for the camera, to link out to.
	URL string `json:"url"`
	// CapturedAt is the image's time (its Last-Modified) when Ourea last checked it, every 15
	// minutes; the image at `image_url` may be newer.
	CapturedAt *time.Time `json:"captured_at,omitempty"`
	// `captured_at` is unknown or more than 60 minutes before `as_of`: the camera may be frozen.
	Stale       bool   `json:"stale"`
	Source      string `json:"source"`
	Attribution string `json:"attribution"`
}
