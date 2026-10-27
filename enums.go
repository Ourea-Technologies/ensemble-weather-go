// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

// Enumerations of the API's response values. Each is a string type, so a
// value the API adds later still parses (as itself) and keeps its raw string;
// its IsKnown method reports false. Treat unknown values as "other".

func oneOf[T ~string](v T, known ...T) bool {
	for _, k := range known {
		if v == k {
			return true
		}
	}
	return false
}

// LocationType is the kind of a named location.
type LocationType string

// The location types.
const (
	LocationAirport            LocationType = "airport"
	LocationBackcountry        LocationType = "backcountry"
	LocationBeachCoast         LocationType = "beach_coast"
	LocationCampground         LocationType = "campground"
	LocationCityTown           LocationType = "city_town"
	LocationClimbingArea       LocationType = "climbing_area"
	LocationCoastalPoint       LocationType = "coastal_point"
	LocationEventVenue         LocationType = "event_venue"
	LocationFireLookout        LocationType = "fire_lookout"
	LocationFishingRiverAccess LocationType = "fishing_river_access"
	LocationGolfCourse         LocationType = "golf_course"
	LocationHotSpring          LocationType = "hot_spring"
	LocationLake               LocationType = "lake"
	LocationMountainPass       LocationType = "mountain_pass"
	LocationNationalPark       LocationType = "national_park"
	LocationObservationStation LocationType = "observation_station"
	LocationOther              LocationType = "other"
	LocationPark               LocationType = "park"
	LocationRecreationArea     LocationType = "recreation_area"
	LocationReservoirLake      LocationType = "reservoir_lake"
	LocationScenicTourist      LocationType = "scenic_tourist"
	LocationSkiArea            LocationType = "ski_area"
	LocationStatePark          LocationType = "state_park"
	LocationSummit             LocationType = "summit"
	LocationTrailhead          LocationType = "trailhead"
	LocationVolcano            LocationType = "volcano"
	LocationWaterfall          LocationType = "waterfall"
	LocationWineAgritourism    LocationType = "wine_agritourism"
)

// IsKnown reports whether t is a location type this SDK knows.
func (t LocationType) IsKnown() bool {
	return oneOf(t, LocationAirport, LocationBackcountry, LocationBeachCoast, LocationCampground,
		LocationCityTown, LocationClimbingArea, LocationCoastalPoint, LocationEventVenue,
		LocationFireLookout, LocationFishingRiverAccess, LocationGolfCourse, LocationHotSpring,
		LocationLake, LocationMountainPass, LocationNationalPark, LocationObservationStation,
		LocationOther, LocationPark, LocationRecreationArea, LocationReservoirLake,
		LocationScenicTourist, LocationSkiArea, LocationStatePark, LocationSummit,
		LocationTrailhead, LocationVolcano, LocationWaterfall, LocationWineAgritourism)
}

// Condition is a stable machine code for an hour's or a day's weather, such
// as "partly_cloudy" or "rain_snow_mix_likely". Switch on it; show
// condition_text to people.
type Condition string

// The condition codes (hourly and daily together).
const (
	ConditionSunny                    Condition = "sunny"
	ConditionMostlySunny              Condition = "mostly_sunny"
	ConditionPartlyCloudy             Condition = "partly_cloudy"
	ConditionMostlyCloudy             Condition = "mostly_cloudy"
	ConditionCloudy                   Condition = "cloudy"
	ConditionClear                    Condition = "clear"
	ConditionMostlyClear              Condition = "mostly_clear"
	ConditionSlightChanceShowers      Condition = "slight_chance_showers"
	ConditionChanceShowers            Condition = "chance_showers"
	ConditionChanceRain               Condition = "chance_rain"
	ConditionShowersLikely            Condition = "showers_likely"
	ConditionRainLikely               Condition = "rain_likely"
	ConditionShowers                  Condition = "showers"
	ConditionRain                     Condition = "rain"
	ConditionLightRain                Condition = "light_rain"
	ConditionHeavyRain                Condition = "heavy_rain"
	ConditionSlightChanceRainSnowMix  Condition = "slight_chance_rain_snow_mix"
	ConditionChanceRainSnowMix        Condition = "chance_rain_snow_mix"
	ConditionRainSnowMixLikely        Condition = "rain_snow_mix_likely"
	ConditionRainSnowMix              Condition = "rain_snow_mix"
	ConditionSlightChanceSnow         Condition = "slight_chance_snow"
	ConditionChanceSnow               Condition = "chance_snow"
	ConditionSnowLikely               Condition = "snow_likely"
	ConditionSnow                     Condition = "snow"
	ConditionSlightChanceFreezingRain Condition = "slight_chance_freezing_rain"
	ConditionChanceFreezingRain       Condition = "chance_freezing_rain"
	ConditionFreezingRainLikely       Condition = "freezing_rain_likely"
	ConditionFreezingRain             Condition = "freezing_rain"
)

// IsKnown reports whether c is a condition code this SDK knows.
func (c Condition) IsKnown() bool {
	return oneOf(c, ConditionSunny, ConditionMostlySunny, ConditionPartlyCloudy, ConditionMostlyCloudy,
		ConditionCloudy, ConditionClear, ConditionMostlyClear, ConditionSlightChanceShowers,
		ConditionChanceShowers, ConditionChanceRain, ConditionShowersLikely, ConditionRainLikely,
		ConditionShowers, ConditionRain, ConditionLightRain, ConditionHeavyRain,
		ConditionSlightChanceRainSnowMix, ConditionChanceRainSnowMix, ConditionRainSnowMixLikely,
		ConditionRainSnowMix, ConditionSlightChanceSnow, ConditionChanceSnow, ConditionSnowLikely,
		ConditionSnow, ConditionSlightChanceFreezingRain, ConditionChanceFreezingRain,
		ConditionFreezingRainLikely, ConditionFreezingRain)
}

// ConditionIcon is an icon hint for a condition, such as "night_snow" or
// "sun_chance_showers". The API documents the full set; it is kept as a
// string because front ends map it to their own artwork.
type ConditionIcon string

// PrecipKind is a precipitation type.
type PrecipKind string

// The precipitation types.
const (
	PrecipNone         PrecipKind = "none"
	PrecipRain         PrecipKind = "rain"
	PrecipSnow         PrecipKind = "snow"
	PrecipRainSnowMix  PrecipKind = "rain_snow_mix"
	PrecipFreezingRain PrecipKind = "freezing_rain"
)

// IsKnown reports whether k is a precipitation type this SDK knows.
func (k PrecipKind) IsKnown() bool {
	return oneOf(k, PrecipNone, PrecipRain, PrecipSnow, PrecipRainSnowMix, PrecipFreezingRain)
}

// NowcastSource is a live input to the nowcast layer.
type NowcastSource string

// The nowcast inputs.
const (
	NowcastObservation    NowcastSource = "observation"
	NowcastRadar          NowcastSource = "radar"
	NowcastCloud          NowcastSource = "cloud"
	NowcastNearestSensors NowcastSource = "nearest_sensors"
)

// IsKnown reports whether s is a nowcast input this SDK knows.
func (s NowcastSource) IsKnown() bool {
	return oneOf(s, NowcastObservation, NowcastRadar, NowcastCloud, NowcastNearestSensors)
}

// MembersView says which per-member view a forecast carries.
type MembersView string

// The member views.
const (
	// MembersPlume (Free) withholds wind_speed, wind_direction and slr.
	MembersPlume MembersView = "plume"
	// MembersFull (Pro) carries every member variable.
	MembersFull MembersView = "full"
)

// IsKnown reports whether v is a member view this SDK knows.
func (v MembersView) IsKnown() bool { return oneOf(v, MembersPlume, MembersFull) }

// CurrentSource says what the current-conditions block was built from.
type CurrentSource string

// The current-conditions sources.
const (
	// CurrentNowcast: at least one live observation carried weight.
	CurrentNowcast CurrentSource = "nowcast"
	// CurrentModel: the calibrated forecast alone.
	CurrentModel CurrentSource = "model"
)

// IsKnown reports whether s is a source this SDK knows.
func (s CurrentSource) IsKnown() bool { return oneOf(s, CurrentNowcast, CurrentModel) }

// ElevationSource says where a point forecast's elevation came from.
type ElevationSource string

// The elevation sources.
const (
	ElevationFromRequest ElevationSource = "request"
	ElevationFromDEM     ElevationSource = "dem_3dep_10m"
)

// IsKnown reports whether s is an elevation source this SDK knows.
func (s ElevationSource) IsKnown() bool { return oneOf(s, ElevationFromRequest, ElevationFromDEM) }

// SourceKind is the kind of a data source in the attribution registry.
type SourceKind string

// The data-source kinds.
const (
	SourceModel       SourceKind = "model"
	SourcePeer        SourceKind = "peer"
	SourceObservation SourceKind = "observation"
	SourceAlerts      SourceKind = "alerts"
	SourceClimatology SourceKind = "climatology"
	SourceTerrain     SourceKind = "terrain"
	SourcePlaces      SourceKind = "places"
	SourceBasemap     SourceKind = "basemap"
)

// IsKnown reports whether k is a source kind this SDK knows.
func (k SourceKind) IsKnown() bool {
	return oneOf(k, SourceModel, SourcePeer, SourceObservation, SourceAlerts, SourceClimatology,
		SourceTerrain, SourcePlaces, SourceBasemap)
}

// InversionKind is the kind of a temperature inversion.
type InversionKind string

// The inversion kinds.
const (
	InversionSurfaceShallow InversionKind = "surface_shallow"
	InversionColdPool       InversionKind = "cold_pool"
	InversionElevated       InversionKind = "elevated"
)

// IsKnown reports whether k is an inversion kind this SDK knows.
func (k InversionKind) IsKnown() bool {
	return oneOf(k, InversionSurfaceShallow, InversionColdPool, InversionElevated)
}

// FreezingLevelKind says which shape a [FreezingLevel] has.
type FreezingLevelKind string

// The freezing-level kinds.
const (
	// FreezingBelowSurface: the whole column is below freezing.
	FreezingBelowSurface FreezingLevelKind = "below_surface"
	// FreezingHeight: the level is at Height.
	FreezingHeight FreezingLevelKind = "height"
	// FreezingAboveProfile: the level is above the profile's Top.
	FreezingAboveProfile FreezingLevelKind = "above_profile"
)

// IsKnown reports whether k is a freezing-level kind this SDK knows.
func (k FreezingLevelKind) IsKnown() bool {
	return oneOf(k, FreezingBelowSurface, FreezingHeight, FreezingAboveProfile)
}

// MoonPhase is a named phase of the moon.
type MoonPhase string

// The moon phases.
const (
	MoonNew            MoonPhase = "new_moon"
	MoonWaxingCrescent MoonPhase = "waxing_crescent"
	MoonFirstQuarter   MoonPhase = "first_quarter"
	MoonWaxingGibbous  MoonPhase = "waxing_gibbous"
	MoonFull           MoonPhase = "full_moon"
	MoonWaningGibbous  MoonPhase = "waning_gibbous"
	MoonLastQuarter    MoonPhase = "last_quarter"
	MoonWaningCrescent MoonPhase = "waning_crescent"
)

// IsKnown reports whether p is a moon phase this SDK knows.
func (p MoonPhase) IsKnown() bool {
	return oneOf(p, MoonNew, MoonWaxingCrescent, MoonFirstQuarter, MoonWaxingGibbous, MoonFull,
		MoonWaningGibbous, MoonLastQuarter, MoonWaningCrescent)
}

// Season is an astronomical season.
type Season string

// The seasons.
const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonAutumn Season = "autumn"
	SeasonWinter Season = "winter"
)

// IsKnown reports whether s is a season this SDK knows.
func (s Season) IsKnown() bool {
	return oneOf(s, SeasonSpring, SeasonSummer, SeasonAutumn, SeasonWinter)
}

// WindStatus classifies a day's wind.
type WindStatus string

// The wind classes.
const (
	WindCalm       WindStatus = "calm"
	WindLight      WindStatus = "light"
	WindModerate   WindStatus = "moderate"
	WindStrong     WindStatus = "strong"
	WindVeryStrong WindStatus = "very_strong"
	WindExtreme    WindStatus = "extreme"
)

// IsKnown reports whether s is a wind class this SDK knows.
func (s WindStatus) IsKnown() bool {
	return oneOf(s, WindCalm, WindLight, WindModerate, WindStrong, WindVeryStrong, WindExtreme)
}

// GapFlowStatus classifies Columbia Gorge gap flow.
type GapFlowStatus string

// The gap-flow classes.
const (
	GapFlowNone     GapFlowStatus = "none"
	GapFlowModerate GapFlowStatus = "moderate"
	GapFlowStrong   GapFlowStatus = "strong"
	GapFlowExtreme  GapFlowStatus = "extreme"
)

// IsKnown reports whether s is a gap-flow class this SDK knows.
func (s GapFlowStatus) IsKnown() bool {
	return oneOf(s, GapFlowNone, GapFlowModerate, GapFlowStrong, GapFlowExtreme)
}

// StormOrientation describes how the snow-to-liquid ratio changes over a storm.
type StormOrientation string

// The storm orientations.
const (
	// StormRightSideUp: dense to fluffy (stable, powder on top).
	StormRightSideUp StormOrientation = "right_side_up"
	// StormUpsideDown: fluffy to dense (weak base).
	StormUpsideDown StormOrientation = "upside_down"
	// StormMixed: no clear trend.
	StormMixed StormOrientation = "mixed"
)

// IsKnown reports whether o is a storm orientation this SDK knows.
func (o StormOrientation) IsKnown() bool {
	return oneOf(o, StormRightSideUp, StormUpsideDown, StormMixed)
}

// SectionKind is the kind of a catalog page template section.
type SectionKind string

// The template section kinds.
const (
	SectionWeatherPanel SectionKind = "weather_panel"
	SectionContentSlot  SectionKind = "content_slot"
	SectionAdSlot       SectionKind = "ad_slot"
	SectionMedia        SectionKind = "media"
)

// IsKnown reports whether k is a section kind this SDK knows.
func (k SectionKind) IsKnown() bool {
	return oneOf(k, SectionWeatherPanel, SectionContentSlot, SectionAdSlot, SectionMedia)
}

// AdSlotKind says how a catalog ad slot is filled.
type AdSlotKind string

// The ad slot kinds.
const (
	AdSlotHouse      AdSlotKind = "house"
	AdSlotThirdParty AdSlotKind = "third_party"
)

// IsKnown reports whether k is an ad slot kind this SDK knows.
func (k AdSlotKind) IsKnown() bool { return oneOf(k, AdSlotHouse, AdSlotThirdParty) }

// AdProvider is a third-party ad provider.
type AdProvider string

// The third-party ad providers.
const (
	AdProviderAdSense AdProvider = "adsense"
	AdProviderGAM     AdProvider = "gam"
	AdProviderCustom  AdProvider = "custom"
)

// IsKnown reports whether p is an ad provider this SDK knows.
func (p AdProvider) IsKnown() bool {
	return oneOf(p, AdProviderAdSense, AdProviderGAM, AdProviderCustom)
}
