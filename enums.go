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
// "sun_chance_showers". Front ends map it to their own artwork.
type ConditionIcon string

// The condition icons.
const (
	IconSunny                  ConditionIcon = "sunny"
	IconMostlySunny            ConditionIcon = "mostly_sunny"
	IconPartlyCloudy           ConditionIcon = "partly_cloudy"
	IconMostlyCloudy           ConditionIcon = "mostly_cloudy"
	IconCloudy                 ConditionIcon = "cloudy"
	IconSunChanceShowers       ConditionIcon = "sun_chance_showers"
	IconSunShowersLikely       ConditionIcon = "sun_showers_likely"
	IconSunShowers             ConditionIcon = "sun_showers"
	IconChanceShowers          ConditionIcon = "chance_showers"
	IconChanceRain             ConditionIcon = "chance_rain"
	IconShowersLikely          ConditionIcon = "showers_likely"
	IconRainLikely             ConditionIcon = "rain_likely"
	IconShowers                ConditionIcon = "showers"
	IconLightRain              ConditionIcon = "light_rain"
	IconRain                   ConditionIcon = "rain"
	IconHeavyRain              ConditionIcon = "heavy_rain"
	IconRainSnowMix            ConditionIcon = "rain_snow_mix"
	IconSnow                   ConditionIcon = "snow"
	IconFreezingRain           ConditionIcon = "freezing_rain"
	IconNightClear             ConditionIcon = "night_clear"
	IconNightMostlyClear       ConditionIcon = "night_mostly_clear"
	IconNightPartlyCloudy      ConditionIcon = "night_partly_cloudy"
	IconNightMostlyCloudy      ConditionIcon = "night_mostly_cloudy"
	IconNightCloudy            ConditionIcon = "night_cloudy"
	IconNightMoonChanceShowers ConditionIcon = "night_moon_chance_showers"
	IconNightMoonShowersLikely ConditionIcon = "night_moon_showers_likely"
	IconNightMoonShowers       ConditionIcon = "night_moon_showers"
	IconNightChanceShowers     ConditionIcon = "night_chance_showers"
	IconNightChanceRain        ConditionIcon = "night_chance_rain"
	IconNightShowersLikely     ConditionIcon = "night_showers_likely"
	IconNightRainLikely        ConditionIcon = "night_rain_likely"
	IconNightShowers           ConditionIcon = "night_showers"
	IconNightLightRain         ConditionIcon = "night_light_rain"
	IconNightRain              ConditionIcon = "night_rain"
	IconNightHeavyRain         ConditionIcon = "night_heavy_rain"
	IconNightRainSnowMix       ConditionIcon = "night_rain_snow_mix"
	IconNightSnow              ConditionIcon = "night_snow"
	IconNightFreezingRain      ConditionIcon = "night_freezing_rain"
)

// IsKnown reports whether v is an icon this SDK knows.
func (v ConditionIcon) IsKnown() bool {
	return oneOf(v, IconSunny, IconMostlySunny, IconPartlyCloudy, IconMostlyCloudy, IconCloudy,
		IconSunChanceShowers, IconSunShowersLikely, IconSunShowers, IconChanceShowers, IconChanceRain,
		IconShowersLikely, IconRainLikely, IconShowers, IconLightRain, IconRain, IconHeavyRain,
		IconRainSnowMix, IconSnow, IconFreezingRain, IconNightClear, IconNightMostlyClear,
		IconNightPartlyCloudy, IconNightMostlyCloudy, IconNightCloudy, IconNightMoonChanceShowers,
		IconNightMoonShowersLikely, IconNightMoonShowers, IconNightChanceShowers, IconNightChanceRain,
		IconNightShowersLikely, IconNightRainLikely, IconNightShowers, IconNightLightRain, IconNightRain,
		IconNightHeavyRain, IconNightRainSnowMix, IconNightSnow, IconNightFreezingRain)
}

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

// AlertSeverity is an alert's CAP severity.
type AlertSeverity string

// The alert severities.
const (
	SeverityExtreme  AlertSeverity = "extreme"
	SeveritySevere   AlertSeverity = "severe"
	SeverityModerate AlertSeverity = "moderate"
	SeverityMinor    AlertSeverity = "minor"
	SeverityUnknown  AlertSeverity = "unknown"
)

// IsKnown reports whether v is a severity this SDK knows.
func (v AlertSeverity) IsKnown() bool {
	return oneOf(v, SeverityExtreme, SeveritySevere, SeverityModerate, SeverityMinor, SeverityUnknown)
}

// AlertUrgency is an alert's CAP urgency.
type AlertUrgency string

// The alert urgencies.
const (
	UrgencyImmediate AlertUrgency = "immediate"
	UrgencyExpected  AlertUrgency = "expected"
	UrgencyFuture    AlertUrgency = "future"
	UrgencyPast      AlertUrgency = "past"
	UrgencyUnknown   AlertUrgency = "unknown"
)

// IsKnown reports whether v is an urgency this SDK knows.
func (v AlertUrgency) IsKnown() bool {
	return oneOf(v, UrgencyImmediate, UrgencyExpected, UrgencyFuture, UrgencyPast, UrgencyUnknown)
}

// AlertCertainty is an alert's CAP certainty.
type AlertCertainty string

// The alert certainties.
const (
	CertaintyObserved AlertCertainty = "observed"
	CertaintyLikely   AlertCertainty = "likely"
	CertaintyPossible AlertCertainty = "possible"
	CertaintyUnlikely AlertCertainty = "unlikely"
	CertaintyUnknown  AlertCertainty = "unknown"
)

// IsKnown reports whether v is a certainty this SDK knows.
func (v AlertCertainty) IsKnown() bool {
	return oneOf(v, CertaintyObserved, CertaintyLikely, CertaintyPossible, CertaintyUnlikely,
		CertaintyUnknown)
}

// StationNetwork is an observation station's network, such as "snotel".
type StationNetwork string

// The station networks.
const (
	NetworkSNOTEL          StationNetwork = "snotel"
	NetworkSnolite         StationNetwork = "snolite"
	NetworkRAWS            StationNetwork = "raws"
	NetworkASOS            StationNetwork = "asos"
	NetworkAWOS            StationNetwork = "awos"
	NetworkMETAR           StationNetwork = "metar"
	NetworkRWISWSDOT       StationNetwork = "rwis_wsdot"
	NetworkRWISODOT        StationNetwork = "rwis_odot"
	NetworkRWISCaltrans    StationNetwork = "rwis_caltrans"
	NetworkRWISIDOT        StationNetwork = "rwis_idot"
	NetworkRWISNDOT        StationNetwork = "rwis_ndot"
	NetworkRWISUDOT        StationNetwork = "rwis_udot"
	NetworkRWISCDOT        StationNetwork = "rwis_cdot"
	NetworkRWISMDT         StationNetwork = "rwis_mdt"
	NetworkRWISWYDOT       StationNetwork = "rwis_wydot"
	NetworkRWISOther       StationNetwork = "rwis_other"
	NetworkAgrimet         StationNetwork = "agrimet"
	NetworkAgweathernet    StationNetwork = "agweathernet"
	NetworkCoCoRaHS        StationNetwork = "cocorahs"
	NetworkCOOP            StationNetwork = "coop"
	NetworkUSGS            StationNetwork = "usgs"
	NetworkTempest         StationNetwork = "tempest"
	NetworkDavis           StationNetwork = "davis"
	NetworkWxflow          StationNetwork = "wxflow"
	NetworkCDEC            StationNetwork = "cdec"
	NetworkCIMIS           StationNetwork = "cimis"
	NetworkMesonetOregon   StationNetwork = "mesonet_oregon"
	NetworkMesonetSynoptic StationNetwork = "mesonet_synoptic"
	NetworkMesonetOther    StationNetwork = "mesonet_other"
	NetworkNDBCBuoy        StationNetwork = "ndbc_buoy"
	NetworkOther           StationNetwork = "other"
	NetworkHADS            StationNetwork = "hads"
	NetworkMADIS           StationNetwork = "madis"
	NetworkNWAC            StationNetwork = "nwac"
	NetworkMTMesonet       StationNetwork = "mt_mesonet"
)

// IsKnown reports whether v is a network this SDK knows.
func (v StationNetwork) IsKnown() bool {
	return oneOf(v, NetworkSNOTEL, NetworkSnolite, NetworkRAWS, NetworkASOS, NetworkAWOS, NetworkMETAR,
		NetworkRWISWSDOT, NetworkRWISODOT, NetworkRWISCaltrans, NetworkRWISIDOT, NetworkRWISNDOT,
		NetworkRWISUDOT, NetworkRWISCDOT, NetworkRWISMDT, NetworkRWISWYDOT, NetworkRWISOther,
		NetworkAgrimet, NetworkAgweathernet, NetworkCoCoRaHS, NetworkCOOP, NetworkUSGS, NetworkTempest,
		NetworkDavis, NetworkWxflow, NetworkCDEC, NetworkCIMIS, NetworkMesonetOregon,
		NetworkMesonetSynoptic, NetworkMesonetOther, NetworkNDBCBuoy, NetworkOther, NetworkHADS,
		NetworkMADIS, NetworkNWAC, NetworkMTMesonet)
}

// StationStatus is an observation station's live-feed status.
type StationStatus string

// The station statuses.
const (
	StationActive         StationStatus = "active"
	StationIntermittent   StationStatus = "intermittent"
	StationInactive       StationStatus = "inactive"
	StationDecommissioned StationStatus = "decommissioned"
	StationUnknown        StationStatus = "unknown"
)

// IsKnown reports whether v is a status this SDK knows.
func (v StationStatus) IsKnown() bool {
	return oneOf(v, StationActive, StationIntermittent, StationInactive, StationDecommissioned,
		StationUnknown)
}

// TemplateVariant is a catalog page template variant.
type TemplateVariant string

// The template variants.
const (
	TemplateMountain TemplateVariant = "mountain"
	TemplateCoastal  TemplateVariant = "coastal"
	TemplateTrail    TemplateVariant = "trail"
	TemplatePass     TemplateVariant = "pass"
	TemplateGeneric  TemplateVariant = "generic"
	TemplateStation  TemplateVariant = "station"
)

// IsKnown reports whether v is a variant this SDK knows.
func (v TemplateVariant) IsKnown() bool {
	return oneOf(v, TemplateMountain, TemplateCoastal, TemplateTrail, TemplatePass, TemplateGeneric,
		TemplateStation)
}

// NullReason says why the server could not answer a probability or window
// total.
type NullReason string

// The reasons.
const (
	// NullWindowNotCovered: the served series does not reach the window's end.
	NullWindowNotCovered NullReason = "window_not_covered"
	// NullTooFewMembers: fewer than two members cover the whole window.
	NullTooFewMembers NullReason = "too_few_members"
	// NullVariableUnavailable: no member carries the variable through the
	// window (for example with Hourly false, which serves no members).
	NullVariableUnavailable NullReason = "variable_unavailable"
)

// IsKnown reports whether v is a reason this SDK knows.
func (v NullReason) IsKnown() bool {
	return oneOf(v, NullWindowNotCovered, NullTooFewMembers, NullVariableUnavailable)
}
