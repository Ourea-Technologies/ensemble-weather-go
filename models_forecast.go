// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"strconv"
	"time"
)

// Quantiles is the API's one distribution shape: an ensemble statistic in the
// request's units, with percentiles over the ensemble members. Every field is
// optional; a nil field is absent (never 0).
//
// The SDK never sums, averages or interpolates quantiles. Ask the server
// instead: [Chance] for a probability, [Total] for a window total.
//
// Wind direction is not a Quantiles: it is a [CircularStats].
type Quantiles struct {
	Mean *float64 `json:"mean,omitempty"`
	P10  *float64 `json:"p10,omitempty"`
	P25  *float64 `json:"p25,omitempty"`
	P50  *float64 `json:"p50,omitempty"`
	P75  *float64 `json:"p75,omitempty"`
	P90  *float64 `json:"p90,omitempty"`
	Min  *float64 `json:"min,omitempty"`
	Max  *float64 `json:"max,omitempty"`
	// Probability is the chance of occurrence (hourly precipitation and
	// snowfall only).
	Probability *float64 `json:"probability,omitempty"`
}

// Interval is a pair of percentiles the server served.
type Interval struct {
	Low, High float64
}

// String formats the interval as "low–high".
func (i Interval) String() string {
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
	return f(i.Low) + "–" + f(i.High)
}

// Median returns the ensemble median (p50). It is the median, not "the
// forecast": the forecast is the whole distribution.
func (q Quantiles) Median() (float64, bool) {
	if q.P50 == nil {
		return 0, false
	}
	return *q.P50, true
}

// Range80 returns the central 80 % range, p10 to p90, as the server sent
// it. ok is false when either end is absent.
//
// Only the served pairs (p10–p90 and p25–p75) have accessors: any other
// range would need interpolation, which the SDK never does.
func (q Quantiles) Range80() (r Interval, ok bool) { return pair(q.P10, q.P90) }

// Range50 returns the central 50 % range, p25 to p75, as the server sent
// it. ok is false when either end is absent.
func (q Quantiles) Range50() (r Interval, ok bool) { return pair(q.P25, q.P75) }

func pair(lo, hi *float64) (Interval, bool) {
	if lo == nil || hi == nil {
		return Interval{}, false
	}
	return Interval{*lo, *hi}, true
}

// ForecastResponse is a forecast for a named location or a point.
type ForecastResponse struct {
	ResponseMeta `json:"-"`

	// Thresholds are the server's fixed thresholds behind the daily probabilities.
	Thresholds *ProbabilityThresholds `json:"thresholds,omitempty"`
	// Probabilities answer [ForecastParams.Chances], in the caller's order
	// (duplicates share one answer). It is nil when none were asked for or
	// the server sent none.
	Probabilities []ProbabilityResult `json:"probabilities,omitempty"`
	// WindowTotals answer [ForecastParams.Totals], in the caller's order;
	// nil when none were asked for or the server sent none.
	WindowTotals []WindowTotal `json:"window_totals,omitempty"`
	// Applied reports options the server clamped to the caller's tier.
	Applied  *AppliedOptions `json:"applied,omitempty"`
	Location LocationRef     `json:"location"`
	ModelRun ModelRunInfo    `json:"model_run"`
	// Units is the unit system of every value in this response.
	Units              Units               `json:"units"`
	Forecast           ForecastData        `json:"forecast"`
	ElevationForecasts []ElevationForecast `json:"elevation_forecasts,omitempty"`
	// NowcastApplied reports whether the live nowcast layer is baked in.
	NowcastApplied bool `json:"nowcast_applied"`
	// NowcastUnavailable lists live inputs this response was built without.
	NowcastUnavailable []NowcastSource `json:"nowcast_unavailable,omitempty"`
	// MembersView is present exactly when per-member values are.
	MembersView    MembersView        `json:"members_view,omitempty"`
	GapFlow        *GapFlow           `json:"gap_flow,omitempty"`
	NearestSensors []NearestSensor    `json:"nearest_sensors,omitempty"`
	Point          *PointInfo         `json:"point,omitempty"`
	Current        *CurrentConditions `json:"current,omitempty"`
	// Attribution names the data sources; carry Credit with redistributed output.
	Attribution ResponseAttribution `json:"attribution"`
}

// ProbabilityResult is one probability term, answered, with its reference
// class: the variable, comparison, threshold, window and member count.
type ProbabilityResult struct {
	// Term is the canonical term the SDK sent, echoed by the server; the
	// SDK matches each answer to the caller's [Chance] by it.
	Term     string   `json:"term"`
	Variable Variable `json:"variable"`
	Op       Op       `json:"op"`
	// Threshold is in the request's units at the served precision.
	Threshold float64 `json:"threshold"`
	// ThresholdUnit is Threshold's unit (set by the SDK from the response's units).
	ThresholdUnit    Unit        `json:"-"`
	WindowStartHours int         `json:"window_start_hours"`
	WindowEndHours   int         `json:"window_end_hours"`
	Aggregation      Aggregation `json:"aggregation"`
	// From is the window's first instant; nil when not covered.
	From *time.Time `json:"from"`
	// To is the window's (exclusive) end; nil when not covered.
	To *time.Time `json:"to"`
	// Probability is the share of members meeting the term, 0..1. It is nil
	// when the server could not answer (see NullReason): absent, never 0.
	Probability *float64 `json:"probability"`
	// Members is the number of members that voted.
	Members int `json:"members"`
	// NullReason says why Probability is nil; nil when it is set.
	NullReason *NullReason `json:"null_reason"`
}

// WindowTotal is the distribution of each member's total over a window.
type WindowTotal struct {
	// Term is the canonical term the SDK sent, echoed by the server; the
	// SDK matches each answer to the caller's [Total] by it.
	Term     string   `json:"term"`
	Variable Variable `json:"variable"`
	// Unit is the quantiles' unit (set by the SDK from the response's units).
	Unit             Unit `json:"-"`
	WindowStartHours int  `json:"window_start_hours"`
	WindowEndHours   int  `json:"window_end_hours"`
	// Aggregation is how each member's hours were combined ("total" for
	// precipitation and snowfall).
	Aggregation Aggregation `json:"aggregation"`
	From        *time.Time  `json:"from"`
	To          *time.Time  `json:"to"`
	// Quantiles is nil when the server could not answer (see NullReason).
	Quantiles *Quantiles `json:"quantiles"`
	Members   int        `json:"members"`
	// NullReason says why Quantiles is nil; nil when it is set.
	NullReason *NullReason `json:"null_reason"`
}

// PrecipDisplay is a day's precipitation headline for display.
type PrecipDisplay struct {
	Kind   PrecipKind `json:"kind"`
	Chance float64    `json:"chance"`
	Amount *float64   `json:"amount,omitempty"`
}

// ProbabilityThresholds are the fixed thresholds behind the hourly and daily
// probabilities, in the request's units.
type ProbabilityThresholds struct {
	HourlyPrecipitation                     float64 `json:"hourly_precipitation"`
	HourlySnowfall                          float64 `json:"hourly_snowfall"`
	DailyPrecipitationProbability           float64 `json:"daily_precipitation_probability"`
	DailyPrecipitationOccurrenceProbability float64 `json:"daily_precipitation_occurrence_probability"`
	DailySnowProbability                    float64 `json:"daily_snow_probability"`
}

// AppliedOptions reports the horizon actually served.
type AppliedOptions struct {
	MaxHours *int `json:"max_hours"`
	MaxDays  *int `json:"max_days"`
	// Clamped names the options the server reduced ("max_hours", "max_days").
	Clamped []string `json:"clamped"`
}

// GapFlow is the Columbia Gorge gap-flow status at the first served hour.
type GapFlow struct {
	Status          GapFlowStatus `json:"status,omitempty"`
	PressureDiffHpa *float64      `json:"pressure_diff_hpa,omitempty"`
	WindMultiplier  *float64      `json:"wind_multiplier,omitempty"`
	Direction       string        `json:"direction,omitempty"`
}

// PointInfo describes the grid point a coordinate forecast was served from.
type PointInfo struct {
	RequestedLat    float64         `json:"requested_lat"`
	RequestedLon    float64         `json:"requested_lon"`
	Lat             float64         `json:"lat"`
	Lon             float64         `json:"lon"`
	ResolutionDeg   float64         `json:"resolution_deg"`
	Elevation       float64         `json:"elevation"`
	ElevationSource ElevationSource `json:"elevation_source"`
	State           string          `json:"state"`
}

// NearestSensor is an observation sensor near the forecast point.
type NearestSensor struct {
	StationID           string    `json:"station_id"`
	Name                string    `json:"name"`
	Source              string    `json:"source"`
	DistanceKm          float64   `json:"distance_km"`
	ElevationDiff       float64   `json:"elevation_diff"`
	EffectiveDistanceKm float64   `json:"effective_distance_km"`
	ObservationTime     time.Time `json:"observation_time"`
}

// ModelRunInfo describes the model run behind a forecast.
type ModelRunInfo struct {
	InitializedAt time.Time    `json:"initialized_at"`
	TotalMembers  int          `json:"total_members"`
	Models        []ModelEntry `json:"models"`
}

// ModelEntry is one model contributing to a run.
type ModelEntry struct {
	ID      string     `json:"id"`
	Cycle   *time.Time `json:"cycle,omitempty"`
	Members *int       `json:"members,omitempty"`
}

// ForecastData holds the hourly series and daily summaries.
type ForecastData struct {
	MaxHours      int `json:"max_hours"`
	IntervalHours int `json:"interval_hours"`
	// Timeseries starts at the current hour (or the run's init with
	// IncludePast).
	Timeseries  []TimeSeriesPoint `json:"timeseries,omitempty"`
	Daily       []DailyForecast   `json:"daily,omitempty"`
	SnowOutlook *SnowOutlook      `json:"snow_outlook,omitempty"`
}

// SnowOutlook holds snow totals for common display windows, as the server
// computes them (sums of hourly snowfall means). For a distribution of a
// window's total, ask for a [Total].
type SnowOutlook struct {
	Next24h         *float64              `json:"next_24h"`
	Next48h         *float64              `json:"next_48h"`
	Next7d          *float64              `json:"next_7d"`
	Overnight4pm9am *float64              `json:"overnight_4pm_9am,omitempty"`
	Quantiles       *SnowOutlookQuantiles `json:"quantiles,omitempty"`
}

// SnowOutlookQuantiles are the distributions of each member's own window total.
type SnowOutlookQuantiles struct {
	Next24h *Quantiles `json:"next_24h,omitempty"`
	Next48h *Quantiles `json:"next_48h,omitempty"`
	Next7d  *Quantiles `json:"next_7d,omitempty"`
}

// TimeSeriesPoint is one forecast hour.
type TimeSeriesPoint struct {
	ValidTime         time.Time      `json:"valid_time"`
	LeadTimeHours     int            `json:"lead_time_hours"`
	IsNight           bool           `json:"is_night"`
	Condition         Condition      `json:"condition"`
	ConditionText     string         `json:"condition_text"`
	ConditionIcon     ConditionIcon  `json:"condition_icon"`
	GustNotable       *bool          `json:"gust_notable,omitempty"`
	PrecipitationType *PrecipType    `json:"precipitation_type,omitempty"`
	Inversion         *InversionInfo `json:"inversion,omitempty"`
	FreezingLevel     *FreezingLevel `json:"freezing_level,omitempty"`
	Ensemble          EnsembleStats  `json:"ensemble"`
	// Provenance is present only with Diagnostics "provenance".
	Provenance *HourProvenance `json:"provenance,omitempty"`
	ModelMeans []ModelMean     `json:"model_means,omitempty"`
	// Members is present only with Members: one entry per model.
	Members           []ModelMembers     `json:"members,omitempty"`
	MembersCumulative *MembersCumulative `json:"members_cumulative,omitempty"`
	// CumulativePrecipitation is the server's running total of the hourly
	// precipitation means from the first served hour.
	CumulativePrecipitation *float64 `json:"cumulative_precipitation,omitempty"`
	CumulativeSnowfall      *float64 `json:"cumulative_snowfall,omitempty"`
}

// HourProvenance names the post-processing rules that shaped each variable
// this hour. It is a research diagnostic, not for display.
type HourProvenance struct {
	Densified               *bool    `json:"densified,omitempty"`
	NGR                     []string `json:"ngr,omitempty"`
	Fallback                []string `json:"fallback,omitempty"`
	Borrowed                []string `json:"borrowed,omitempty"`
	SpreadFloored           []string `json:"spread_floored,omitempty"`
	FloorClamped            []string `json:"floor_clamped,omitempty"`
	BucketSplit             []string `json:"bucket_split,omitempty"`
	DistributionParameters  []string `json:"distribution_parameters,omitempty"`
	SpreadInterpolated      []string `json:"spread_interpolated,omitempty"`
	InterpolatedContributor []string `json:"interpolated_contributor,omitempty"`
	MomentMatchedFamily     []string `json:"moment_matched_family,omitempty"`
	LapseProfile            []string `json:"lapse_profile,omitempty"`
	LapseClimatology        []string `json:"lapse_climatology,omitempty"`
	LapsePRISM              []string `json:"lapse_prism,omitempty"`
	SpreadClampedLow        []string `json:"spread_clamped_low,omitempty"`
	SpreadClampedHigh       []string `json:"spread_clamped_high,omitempty"`
	GustDressed             []string `json:"gust_dressed,omitempty"`
}

// Sector is a clockwise arc of directions, in degrees.
type Sector struct {
	From float64 `json:"from"`
	To   float64 `json:"to"`
}

// CircularStats describes wind direction as a circular distribution, in
// degrees clockwise from north. Each sector is a clockwise arc, so its To
// may be numerically less than its From (the arc crosses north); the SDK
// does not "fix" this.
type CircularStats struct {
	// Mean is the members' resultant-vector mean direction, in [0, 360).
	Mean float64 `json:"mean"`
	// Stddev is the members' angular spread around Mean; nil on a
	// single-member hour.
	Stddev *float64 `json:"stddev,omitempty"`
	// Sector50 holds the central 50 % of the members (25th to 75th circular
	// percentile); Sector90 the central 90 % (5th to 95th).
	Sector50 Sector `json:"sector_50"`
	Sector90 Sector `json:"sector_90"`
}

// ModelMembers holds one model's member values at an hour.
type ModelMembers struct {
	ID      string           `json:"id"`
	Members []MemberForecast `json:"members"`
}

// MemberForecast is one ensemble member's values at an hour. The plume view
// (Free) withholds wind and SLR.
type MemberForecast struct {
	MemberID                int      `json:"member_id"`
	Temperature             *float64 `json:"temperature,omitempty"`
	Precipitation           *float64 `json:"precipitation,omitempty"`
	PrecipitationHours      *int     `json:"precipitation_hours,omitempty"`
	WindSpeed               *float64 `json:"wind_speed,omitempty"`
	WindDirection           *float64 `json:"wind_direction,omitempty"`
	SnowLevel               *float64 `json:"snow_level,omitempty"`
	SLR                     *float64 `json:"slr,omitempty"`
	Snowfall                *float64 `json:"snowfall,omitempty"`
	CumulativePrecipitation *float64 `json:"cumulative_precipitation,omitempty"`
	CumulativeSnowfall      *float64 `json:"cumulative_snowfall,omitempty"`
}

// MembersCumulative are the distributions of the members' running totals.
type MembersCumulative struct {
	Precipitation *Quantiles `json:"precipitation,omitempty"`
	Snowfall      *Quantiles `json:"snowfall,omitempty"`
}

// InversionInfo describes a detected temperature inversion or cold pool.
type InversionInfo struct {
	Kind     InversionKind `json:"kind"`
	Depth    float64       `json:"depth"`
	Strength float64       `json:"strength"`
}

// PrecipType is an hour's precipitation type with its probabilities.
type PrecipType struct {
	Type            PrecipKind `json:"type"`
	SnowProbability float64    `json:"snow_probability"`
	RainProbability float64    `json:"rain_probability"`
}

// EnsembleStats are an hour's ensemble distributions, in the request's units.
type EnsembleStats struct {
	Temperature   Quantiles  `json:"temperature"`
	FeelsLike     *Quantiles `json:"feels_like,omitempty"`
	Dewpoint      *Quantiles `json:"dewpoint,omitempty"`
	Precipitation Quantiles  `json:"precipitation"`
	WindSpeed     Quantiles  `json:"wind_speed"`
	WindGust      *Quantiles `json:"wind_gust,omitempty"`
	// WindDirection is the circular distribution of the members' directions.
	WindDirection CircularStats `json:"wind_direction"`
	SnowLevel     *Quantiles    `json:"snow_level,omitempty"`
	SLR           *Quantiles    `json:"slr,omitempty"`
	Snowfall      *Quantiles    `json:"snowfall,omitempty"`
	CloudCover    *Quantiles    `json:"cloud_cover,omitempty"`
	// SkinTemperature and MSLP are present only when asked for with
	// [ForecastParams.Include].
	SkinTemperature  *Quantiles `json:"skin_temperature,omitempty"`
	MSLP             *Quantiles `json:"mslp,omitempty"`
	CAPE             *Quantiles `json:"cape,omitempty"`
	RelativeHumidity *Quantiles `json:"relative_humidity,omitempty"`
}

// FreezingLevel is an hour's freezing level. Kind says which fields are set:
// Height (and AboveFreezingLayerAloft) for [FreezingHeight], Top for
// [FreezingAboveProfile], none for [FreezingBelowSurface].
type FreezingLevel struct {
	Kind                    FreezingLevelKind `json:"kind"`
	Height                  *float64          `json:"height,omitempty"`
	AboveFreezingLayerAloft *bool             `json:"above_freezing_layer_aloft,omitempty"`
	Top                     *float64          `json:"top,omitempty"`
}

// ModelMean is one model's statistics at an hour.
type ModelMean struct {
	ID            string         `json:"id"`
	Temperature   *Quantiles     `json:"temperature,omitempty"`
	Dewpoint      *Quantiles     `json:"dewpoint,omitempty"`
	Precipitation *Quantiles     `json:"precipitation,omitempty"`
	WindSpeed     *Quantiles     `json:"wind_speed,omitempty"`
	WindDirection *CircularStats `json:"wind_direction,omitempty"`
	SnowLevel     *Quantiles     `json:"snow_level,omitempty"`
	SLR           *Quantiles     `json:"slr,omitempty"`
	Snowfall      *Quantiles     `json:"snowfall,omitempty"`
	Interpolated  *bool          `json:"interpolated,omitempty"`
}

// DailyForecast summarises one local calendar date.
type DailyForecast struct {
	// Date is the local date, "YYYY-MM-DD", in the location's timezone.
	Date          string        `json:"date"`
	Condition     Condition     `json:"condition"`
	ConditionText string        `json:"condition_text"`
	ConditionIcon ConditionIcon `json:"condition_icon"`
	// PrecipitationType is nil when the day is unlikely to precipitate.
	PrecipitationType *PrecipKind `json:"precipitation_type,omitempty"`
	High              Quantiles   `json:"high"`
	Low               Quantiles   `json:"low"`
	FeelsLikeHigh     *Quantiles  `json:"feels_like_high,omitempty"`
	FeelsLikeLow      *Quantiles  `json:"feels_like_low,omitempty"`
	// TotalPrecipitation is the distribution of each member's own day total.
	TotalPrecipitation Quantiles  `json:"total_precipitation"`
	TotalSnowfall      *Quantiles `json:"total_snowfall,omitempty"`
	// PrecipitationProbability is P(day total >= 2.5 mm).
	PrecipitationProbability float64 `json:"precipitation_probability"`
	// PrecipitationOccurrenceProbability is P(day total >= 1.0 mm).
	PrecipitationOccurrenceProbability float64 `json:"precipitation_occurrence_probability"`
	// SnowProbability is P(day snowfall >= 2.5 mm).
	SnowProbability             float64           `json:"snow_probability"`
	SnowLevelMean               *float64          `json:"snow_level_mean"`
	SnowLevel                   *Quantiles        `json:"snow_level,omitempty"`
	TotalPrecipitationIfWet     *Quantiles        `json:"total_precipitation_if_wet,omitempty"`
	TotalSnowfallIfWet          *Quantiles        `json:"total_snowfall_if_wet,omitempty"`
	PrecipitationDisplay        PrecipDisplay     `json:"precipitation_display"`
	CloudCoverMean              *float64          `json:"cloud_cover_mean,omitempty"`
	Wind                        *DailyWind        `json:"wind,omitempty"`
	StormOrientation            *StormOrientation `json:"storm_orientation,omitempty"`
	StormOrientationDescription *string           `json:"storm_orientation_description,omitempty"`
	Astro                       DailyAstro        `json:"astro"`
	TemperatureClimatology      *DayClimatology   `json:"temperature_climatology,omitempty"`
}

// DailyWind summarises a day's wind (gusts when available).
type DailyWind struct {
	Status         WindStatus `json:"status,omitempty"`
	PeakSpeed      *float64   `json:"peak_speed,omitempty"`
	PeakSpeedP90   *float64   `json:"peak_speed_p90,omitempty"`
	IsGust         *bool      `json:"is_gust,omitempty"`
	SustainedSpeed *float64   `json:"sustained_speed,omitempty"`
}

// DailyAstro holds a day's sun and moon.
type DailyAstro struct {
	Sunrise            *time.Time `json:"sunrise,omitempty"`
	Sunset             *time.Time `json:"sunset,omitempty"`
	MoonPhase          MoonPhase  `json:"moon_phase"`
	MoonIllumination   float64    `json:"moon_illumination"`
	AstronomicalSeason Season     `json:"astronomical_season"`
}

// DayClimatology is the date's temperature climatology.
type DayClimatology struct {
	Mean float64 `json:"mean"`
	Std  float64 `json:"std"`
}

// ElevationForecast is a forecast at one of a location's extra elevations
// (ski areas: base, mid, summit).
type ElevationForecast struct {
	Name        string            `json:"name"`
	Elevation   float64           `json:"elevation"`
	Lat         float64           `json:"lat"`
	Lon         float64           `json:"lon"`
	Hourly      []TimeSeriesPoint `json:"hourly,omitempty"`
	Daily       []DailyForecast   `json:"daily,omitempty"`
	SnowOutlook *SnowOutlook      `json:"snow_outlook,omitempty"`
}

// ResponseAttribution names the data sources a response drew on.
type ResponseAttribution struct {
	Sources []string `json:"sources"`
	// Credit is the line to carry with redistributed output.
	Credit string `json:"credit"`
	URL    string `json:"url"`
}
