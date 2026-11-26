// SPDX-License-Identifier: BSD-2-Clause

// The snowpack forecast and its snow-depth plume (beta).

package ensembleweather

import (
	"time"
)

// SnowpackResponse is the snowpack forecast of a location's spots.
type SnowpackResponse struct {
	ResponseMeta `json:"-"`

	// Always `true` while the product is a beta; fields may be added.
	Beta     bool         `json:"beta"`
	Location LocationRef  `json:"location"`
	Units    Units        `json:"units"`
	ModelRun *SnowpackRun `json:"model_run,omitempty"`
	// Stale is the newest run's blend cycle is more than 24 h old.
	Stale bool `json:"stale"`
	// Spots is the location's own point first, then its named elevations in their display order.
	Spots       []SnowpackSpot      `json:"spots"`
	Definitions SnowpackDefinitions `json:"definitions"`
}

// SnowpackDefinitions maps each surface class and layer form code to its meaning. The keys are data
// (such as "buried_surface_hoar") and are kept as the API sends them.
type SnowpackDefinitions struct {
	// Each surface class and what it means.
	Surface map[string]string `json:"surface"`
	// Each layer form and what it is.
	LayerForm map[string]string `json:"layer_form"`
}

// SnowpackRun is the newest snowpack run among the spots served.
type SnowpackRun struct {
	// InitializedAt is the blend cycle the snowpack model was driven with.
	InitializedAt time.Time     `json:"initialized_at"`
	GeneratedAt   time.Time     `json:"generated_at"`
	Model         SnowpackModel `json:"model"`
}

// SnowpackModel names the snowpack model that ran, its authors and licence.
type SnowpackModel struct {
	Name    string `json:"name"`
	By      string `json:"by"`
	License string `json:"license"`
	// Version is the build that ran: the commits of Ourea's SNOWPACK and MeteoIO source it was built
	// from, as `snowpack <commit> meteoio <commit>`.
	Version *string `json:"version,omitempty"`
}

// SnowpackSpot is the snowpack forecast at one of a location's points.
type SnowpackSpot struct {
	// Name is the named elevation (`location.elevations[].name`); `null` is the location's own point.
	Name *string `json:"name"`
	// What the named elevation is called, as `ElevationSummary.display_name`; absent for the
	// location's own point and where the catalog has no name for the elevation.
	DisplayName *string `json:"display_name,omitempty"`
	// In the response's units (m or ft).
	Elevation  float64 `json:"elevation"`
	ElevationM float64 `json:"elevation_m"`
	// InitializedAt is the blend cycle this spot's forecast was driven with; older than
	// `model_run.initialized_at` when the newest run did not reach the spot.
	InitializedAt time.Time `json:"initialized_at"`
	// Anchor is the snow-depth sensor at the spot.
	Anchor *SnowpackAnchor `json:"anchor,omitempty"`
	Depth  SnowpackDepth   `json:"depth"`
	// DepthChange24h is the depth at `depth.start_time` less this spot's modelled depth 24 h before,
	// in the depth's units and rounding (cm or in; a loss is negative): the model's own change,
	// snowfall, settlement and melt together.
	DepthChange24h *float64 `json:"depth_change_24h,omitempty"`
	// Hourly snow water equivalent (mm or in), parallel to `depth.p50`.
	SWE []float64 `json:"swe"`
	// One entry every six hours.
	Steps []SnowpackStep `json:"steps"`
	// Buried layers at `depth.start_time`, at most three.
	Layers []SnowpackLayer `json:"layers"`
	// Aspects is the snow on a 38° slope at the spot facing each way, in the order north, east,
	// south, west.
	Aspects []SnowpackAspect `json:"aspects,omitempty"`
}

// SnowpackAnchor is the snow-depth sensor at the spot.
type SnowpackAnchor struct {
	// Station is the station's catalog slug.
	Station string `json:"station"`
	Name    string `json:"name"`
	Network string `json:"network"`
}

// SnowpackStep is the snowpack's surface and new snow at one step.
type SnowpackStep struct {
	ValidTime time.Time `json:"valid_time"`
	// Surface is a key of `definitions.surface`: today `none`, `powder`, `wind_affected`, `settled`,
	// `wet`, `crust` or `ice_crust`.
	Surface string `json:"surface"`
	// New snow of the 24 h to `valid_time`, as it lies (cm or in).
	NewSnow24h float64 `json:"new_snow_24h"`
	// Its density.
	NewSnowDensityKgM3 *float64 `json:"new_snow_density_kg_m3,omitempty"`
	// How far the snow on the ground 24 h earlier settled (cm or in).
	Settlement24h *float64 `json:"settlement_24h,omitempty"`
}

// SnowpackLayer is one layer of the modelled snowpack.
type SnowpackLayer struct {
	// Form is a key of `definitions.layer_form`: today `faceted_grains`, `depth_hoar`,
	// `buried_surface_hoar`, `melt_freeze_crust`, `ice_layer` or `rounding_facets`.
	Form string `json:"form"`
	// cm or in.
	Thickness float64 `json:"thickness"`
	// From the surface down to the layer's top (cm or in).
	Depth  float64 `json:"depth"`
	Formed string  `json:"formed"`
}

// SnowpackAspect is the snow on one slope at a spot, described as the spot's flat ground is: the
// model run on a slope of `slope_angle_deg` facing `aspect`, with the flat ground's snowfall and
// the slope's own sun.
type SnowpackAspect struct {
	// Aspect is the way the slope faces: today `north`, `east`, `south` or `west`.
	Aspect string `json:"aspect"`
	// How steep the slope is, degrees from horizontal.
	SlopeAngleDeg float64 `json:"slope_angle_deg"`
	// Hourly snow depth on the slope, measured vertically as the flat depth is.
	Depth *SnowpackDepth `json:"depth"`
	// As the spot's `depth_change_24h`, this slope's own (open loop on both sides), read from the
	// same earlier run as the spot's and from no other.
	DepthChange24h *float64 `json:"depth_change_24h,omitempty"`
	// One entry every six hours, at the spot's own `steps[].valid_time`.
	Steps []SnowpackStep `json:"steps"`
	// Buried layers on the slope at `depth.start_time`, at most three.
	Layers []SnowpackLayer `json:"layers"`
}

// SnowpackDepth is hourly snow depth from `start_time` (cm or in).
type SnowpackDepth struct {
	StartTime     time.Time `json:"start_time"`
	IntervalHours int       `json:"interval_hours"`
	// `measured`: started from a sensor reading at the spot.
	Basis string    `json:"basis"`
	P10   []float64 `json:"p10,omitempty"`
	P50   []float64 `json:"p50"`
	P90   []float64 `json:"p90,omitempty"`
}

// SnowpackPlumeResponse is a location's snow-depth plume.
type SnowpackPlumeResponse struct {
	ResponseMeta `json:"-"`

	// Always `true` while the product is a beta; fields may be added.
	Beta     bool        `json:"beta"`
	Location LocationRef `json:"location"`
	Units    Units       `json:"units"`
	// ModelRun is the run the plume is from: the 12Z blend cycle the members were read from.
	ModelRun *SnowpackRun `json:"model_run,omitempty"`
	// Stale is the plume's blend cycle is more than 24 h old.
	Stale bool `json:"stale"`
	// Depth is the members' depth at each step.
	Depth *SnowpackPlumeDepth `json:"depth,omitempty"`
	// Every member's own series, by model, each model to its own horizon.
	Models []SnowpackPlumeModel `json:"models,omitempty"`
	// Every member's depth as ONE unlabelled pool, for Free and anonymous callers (absent for Plus
	// and Pro, who are served `models`).
	MemberDepths [][]float64 `json:"member_depths,omitempty"`
	// What the bands are, in a sentence a client may show beside them.
	Spread string `json:"spread"`
}

// SnowpackPlumeDepth is the members' depth at each step.
type SnowpackPlumeDepth struct {
	// StartTime is the time of the first value of every series in the body.
	StartTime time.Time `json:"start_time"`
	// Hours between values (3).
	IntervalHours int `json:"interval_hours"`
	// Members that reach each step; it falls where a model's horizon ends.
	Members []int `json:"members"`
	// P50 is the members' median depth, in the response's units.
	P50 []float64 `json:"p50,omitempty"`
	// Served to every caller, as the bands are.
	Mean []float64 `json:"mean"`
	// Every caller (Plus and Pro only before).
	P10 []float64 `json:"p10"`
	// Every caller (Plus and Pro only before).
	P25 []float64 `json:"p25"`
	// Every caller (Plus and Pro only before).
	P75 []float64 `json:"p75"`
	// Every caller (Plus and Pro only before).
	P90 []float64 `json:"p90"`
}

// SnowpackPlumeModel is one model's members in the snow-depth plume.
type SnowpackPlumeModel struct {
	// `aifs`, `aigefs`, `gefs` or `ifs_ens`; a client must tolerate one it does not know.
	ID string `json:"id"`
	// MemberIDs is the model's own member numbers, in the order of `depth`.
	MemberIDs []int `json:"member_ids"`
	// One series per member, from `depth.start_time` every `depth.interval_hours`.
	Depth [][]float64 `json:"depth"`
	// DepthMean is the equal-weight mean of this model's members' `depth` at each of the model's own
	// steps (from `depth.start_time` every `depth.interval_hours`, to the model's horizon: the length
	// of its members' series), in the response's units.
	DepthMean []float64 `json:"depth_mean"`
	// Members not run because the ground starts bare and they bring no snow; their series is 0
	// throughout.
	MembersSkipped int `json:"members_skipped"`
}
