// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"encoding/json"
	"time"
)

// LocationRef identifies a named location, or the point of a coordinate
// request.
type LocationRef struct {
	// Slug is empty for a coordinate request that matched no named location
	// (the server sends null).
	Slug  string  `json:"slug,omitempty"`
	State string  `json:"state"`
	Name  string  `json:"name"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	// ElevationM is in metres whatever the request's units.
	ElevationM float64 `json:"elevation_m"`
	// Type is empty when Slug is (the server sends null).
	Type LocationType `json:"type,omitempty"`
	// Timezone is the location's IANA timezone, such as "America/Los_Angeles".
	Timezone string `json:"timezone"`
}

// Key returns the location's "state/slug" form, as the location methods take
// it, or "" when the location has no slug (a coordinate request's point).
func (l LocationRef) Key() string {
	if l.Slug == "" {
		return ""
	}
	return l.State + "/" + l.Slug
}

// ActiveAlert is a weather alert in effect.
type ActiveAlert struct {
	ID        string         `json:"id"`
	Event     string         `json:"event"`
	Severity  AlertSeverity  `json:"severity"`
	Urgency   AlertUrgency   `json:"urgency"`
	Certainty AlertCertainty `json:"certainty"`
	Headline  string         `json:"headline"`
	Onset     *time.Time     `json:"onset,omitempty"`
	Expires   time.Time      `json:"expires"`
	Ends      *time.Time     `json:"ends,omitempty"`
	// Sent is when NWS issued this message.
	Sent *time.Time `json:"sent,omitempty"`
	// MessageType is the CAP message type ("alert", "update" or "cancel"). It
	// is not a sign of a new event: dedupe on EventKey.
	MessageType string `json:"message_type,omitempty"`
	// References are the ids of the messages this one replaces.
	References []string `json:"references,omitempty"`
	SenderName string   `json:"sender_name,omitempty"`
	// AreaDesc is the places this message covers, as NWS words them.
	AreaDesc string `json:"area_desc,omitempty"`
	// Description is the full NWS text, unedited.
	Description string `json:"description,omitempty"`
	// Instruction is what NWS says to do, unedited; empty when NWS gives none.
	Instruction string `json:"instruction,omitempty"`
	// SourceURL is the NWS forecast page for this location.
	SourceURL string `json:"source_url,omitempty"`
	// EventKey is the stable identity of this event: every message in one
	// event has the same key, so to notify once per event, dedupe on it.
	// Treat it as opaque.
	EventKey string `json:"event_key,omitempty"`
	// VTECAction is the message's VTEC action, lower case ("new", "con",
	// "ext", "exa", "exb", "upg", "can", "exp", "cor", "rou"); "can", "exp"
	// and "upg" end the event. Empty when NWS sent no VTEC.
	VTECAction string `json:"vtec_action,omitempty"`
}

// CurrentResponse is current conditions at a location or point.
type CurrentResponse struct {
	ResponseMeta `json:"-"`

	Location    LocationRef         `json:"location"`
	Alerts      []ActiveAlert       `json:"alerts"`
	AlertsAsOf  time.Time           `json:"alerts_as_of"`
	SensorCount int                 `json:"sensor_count"`
	Sensors     []SensorObservation `json:"sensors"`
	// Now (JSON `current`) is nil when no served hour is near the request time.
	Now                   *CurrentConditions `json:"current"`
	AsOf                  time.Time          `json:"as_of"`
	ModelRunInitializedAt *time.Time         `json:"model_run_initialized_at"`
	// NowcastHours is present only with IncludeNowcastHours.
	NowcastHours       []TimeSeriesPoint   `json:"nowcast_hours"`
	Units              Units               `json:"units"`
	Attribution        ResponseAttribution `json:"attribution"`
	NowcastUnavailable []NowcastSource     `json:"nowcast_unavailable,omitempty"`
	// Summary is the place summary, written by the server from served values
	// only, so every client shows the same words; nil when no cycle is servable.
	Summary *PlaceSummary `json:"summary,omitempty"`
	// Pass is the mountain pass forecast at the pass's own point and
	// elevation, present only for a [LocationMountainPass] location.
	Pass *PassForecast `json:"pass,omitempty"`
}

// CurrentConditions is "right now": a view of one served forecast hour.
type CurrentConditions struct {
	Time        time.Time `json:"time"`
	Temperature float64   `json:"temperature"`
	FeelsLike   *float64  `json:"feels_like,omitempty"`
	// Code (JSON `condition`) is the hour's condition code; switch on it.
	Code Condition `json:"condition"`
	// Text (JSON `condition_text`) is Code's display text; show it, never parse it.
	Text string `json:"condition_text"`
	// Icon (JSON `condition_icon`) is the icon hint for Code.
	Icon ConditionIcon `json:"condition_icon"`
	// Measured reports that the radar measured precipitation in the hour in
	// progress: show "Measured" in place of the chance.
	Measured                 bool     `json:"measured"`
	WindSpeed                float64  `json:"wind_speed"`
	WindGust                 *float64 `json:"wind_gust,omitempty"`
	WindDirection            float64  `json:"wind_direction"`
	RelativeHumidity         *float64 `json:"relative_humidity,omitempty"`
	Dewpoint                 *float64 `json:"dewpoint,omitempty"`
	PrecipitationProbability *float64 `json:"precipitation_probability,omitempty"`
	// CloudCover is the hour's mean cloud cover, percent (0-100); nil where
	// the hour serves none.
	CloudCover *float64        `json:"cloud_cover,omitempty"`
	Source     CurrentSource   `json:"source"`
	Station    *CurrentStation `json:"station,omitempty"`
}

// CurrentStation is the surface station that anchored the current hour.
type CurrentStation struct {
	// ID (JSON `station_id`) is the station's id in its network, as in
	// [SensorObservation.StationID].
	ID              string    `json:"station_id"`
	Name            string    `json:"name"`
	ObservationTime time.Time `json:"observation_time"`
	// Calm reports a calm reading (0 wind speed). A calm reading has no
	// bearing, so the served wind direction is then the forecast's, not the
	// station's.
	Calm bool `json:"calm"`
}

// SensorObservation is one sensor's latest reading, in the request's units.
type SensorObservation struct {
	StationID string `json:"station_id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	// ObservedAt (JSON `observation_time`) is when the reading was taken.
	ObservedAt time.Time `json:"observation_time"`
	// DistanceKm is the straight-line distance from the location to the
	// sensor (at most 25 km).
	DistanceKm float64 `json:"distance_km"`
	// ElevationDiff is the sensor's elevation minus the location's, in the
	// request's length units; nil when the station has no registry elevation.
	ElevationDiff      *float64 `json:"elevation_diff,omitempty"`
	Temperature        *float64 `json:"temperature,omitempty"`
	Dewpoint           *float64 `json:"dewpoint,omitempty"`
	RelativeHumidity   *float64 `json:"relative_humidity,omitempty"`
	WindSpeed          *float64 `json:"wind_speed,omitempty"`
	WindGust           *float64 `json:"wind_gust,omitempty"`
	WindDirection      *float64 `json:"wind_direction,omitempty"`
	Pressure           *float64 `json:"pressure,omitempty"`
	Precipitation      *float64 `json:"precipitation,omitempty"`
	PrecipitationRate  *float64 `json:"precipitation_rate,omitempty"`
	Reflectivity       *float64 `json:"reflectivity,omitempty"`
	PavementTemp       *float64 `json:"pavement_temp,omitempty"`
	SubsurfaceTemp     *float64 `json:"subsurface_temp,omitempty"`
	Visibility         *float64 `json:"visibility,omitempty"`
	CloudCover         *float64 `json:"cloud_cover,omitempty"`
	CeilingHeight      *float64 `json:"ceiling_height,omitempty"`
	SnowDepth          *float64 `json:"snow_depth,omitempty"`
	SWE                *float64 `json:"swe,omitempty"`
	SWEPercentOfMedian *float64 `json:"swe_percent_of_median,omitempty"`
	Snowfall           *float64 `json:"snowfall,omitempty"`
	SolarRadiation     *float64 `json:"solar_radiation,omitempty"`
	PavementCondition  *string  `json:"pavement_condition,omitempty"`
}

// AlertsResponse lists the alerts in effect at a location or point.
type AlertsResponse struct {
	ResponseMeta `json:"-"`

	Location LocationRef `json:"location"`
	// Count (JSON `alert_count`) is the number of alerts in Items.
	Count int       `json:"alert_count"`
	AsOf  time.Time `json:"as_of"`
	// Items (JSON `alerts`) are the alerts in effect.
	Items       []ActiveAlert       `json:"alerts"`
	Attribution ResponseAttribution `json:"attribution"`
}

// LocationSummary is a named location in a list.
type LocationSummary struct {
	LocationRef
	County      string  `json:"county"`
	Description *string `json:"description,omitempty"`
	// Elevations are the location's named elevation points, in the request's
	// length units, the primary (Default) first.
	Elevations []ElevationSummary `json:"elevations,omitempty"`
}

// NearbyLocation is a location near another, with its distance.
type NearbyLocation struct {
	LocationSummary
	DistanceKm float64 `json:"distance_km"`
}

// NearbyLocationsResponse lists the locations near a named location.
type NearbyLocationsResponse struct {
	ResponseMeta `json:"-"`

	// Origin (JSON `location`) is the location the search started from.
	Origin   LocationSummary `json:"location"`
	RadiusKm float64         `json:"radius_km"`
	Count    int             `json:"count"`
	// Items (JSON `locations`) are the neighbours, nearest first.
	Items []NearbyLocation `json:"locations"`
}

// NearbyPointResponse lists the locations near a point.
type NearbyPointResponse struct {
	ResponseMeta `json:"-"`

	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	RadiusKm float64 `json:"radius_km"`
	Count    int     `json:"count"`
	// Items (JSON `locations`) are the neighbours, nearest first.
	Items []NearbyLocation `json:"locations"`
}

// LocationsResponse is one page of a state's (or every state's) locations.
type LocationsResponse struct {
	ResponseMeta `json:"-"`

	State string `json:"state"`
	Count int    `json:"count"`
	// Items (JSON `locations`) are this page's locations.
	Items []LocationSummary `json:"locations"`
	// NextCursor is nil on the last page.
	NextCursor *string `json:"next_cursor"`
}

// SearchLocationsResponse holds location search hits.
type SearchLocationsResponse struct {
	ResponseMeta `json:"-"`

	Items []SearchHit `json:"items"`
	Count int         `json:"count"`
}

// SearchHit is one location search result.
type SearchHit struct {
	// Key is the hit's state and slug joined by a dash, such as
	// "or-mt-hood-meadows". It is not the "state/slug" form the location
	// methods take: use [LocationRef.Key] for that.
	Key string `json:"key"`
	LocationRef
	County string `json:"county"`
	// Explore is the place's value as [Client.Explore] ranks it, present only
	// with a metric [SearchParams.Sort]; nil where the place has none.
	Explore *ExploreValue `json:"explore,omitempty"`
}

// CatalogLocationPage is one page of the location catalog.
type CatalogLocationPage struct {
	ResponseMeta `json:"-"`

	Items []CatalogLocation `json:"items"`
	// NextCursor is nil on the last page.
	NextCursor *string `json:"next_cursor"`
}

// CatalogStationPage is one page of the station catalog.
type CatalogStationPage struct {
	ResponseMeta `json:"-"`

	Items []CatalogStation `json:"items"`
	// NextCursor is nil on the last page.
	NextCursor *string `json:"next_cursor"`
}

// CatalogLocation is a location's catalog record.
type CatalogLocation struct {
	LocationRef
	County            *string         `json:"county,omitempty"`
	Country           string          `json:"country,omitempty"`
	Aliases           []string        `json:"aliases,omitempty"`
	DescriptionShort  *string         `json:"description_short,omitempty"`
	DescriptionLong   *string         `json:"description_long,omitempty"`
	MicroclimateNotes *string         `json:"microclimate_notes,omitempty"`
	MetaTitle         *string         `json:"meta_title,omitempty"`
	MetaDescription   *string         `json:"meta_description,omitempty"`
	BestSeasons       []string        `json:"best_seasons,omitempty"`
	KeyFactors        []string        `json:"key_factors,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	ExternalLinks     json.RawMessage `json:"external_links,omitempty"`
	IsFeatured        bool            `json:"is_featured,omitempty"`
}

// CatalogLocationDetail is a location's full catalog record.
type CatalogLocationDetail struct {
	ResponseMeta `json:"-"`
	CatalogLocation

	Elevations     []CatalogElevation     `json:"elevations"`
	Photos         []CatalogPhoto         `json:"photos"`
	Webcams        []CatalogWebcam        `json:"webcams"`
	NearbyStations []CatalogNearbyStation `json:"nearby_stations"`
	// Template and Content are present when asked for with Include.
	Template *CatalogTemplate      `json:"template,omitempty"`
	Content  []CatalogContentEntry `json:"content,omitempty"`
}

// CatalogStation is an observation station's catalog record.
type CatalogStation struct {
	LocationRef
	County  *string        `json:"county,omitempty"`
	Country string         `json:"country,omitempty"`
	Aliases []string       `json:"aliases,omitempty"`
	Network StationNetwork `json:"network,omitempty"`
	// SourceID (JSON `source_station_id`) is the station's id in its network.
	SourceID           string   `json:"source_station_id,omitempty"`
	ElevationMslM      *float64 `json:"elevation_msl_m,omitempty"`
	SensorHeightAglM   *float64 `json:"sensor_height_agl_m,omitempty"`
	SensorCapabilities []string `json:"sensor_capabilities,omitempty"`
	// Status (JSON `station_status`) is the station's status.
	Status                 StationStatus   `json:"station_status,omitempty"`
	LastReportAt           *time.Time      `json:"last_report_at,omitempty"`
	LastReportSummary      json.RawMessage `json:"last_report_summary,omitempty"`
	InstallDate            *string         `json:"install_date,omitempty"`
	DecommissionDate       *string         `json:"decommission_date,omitempty"`
	UpdateFrequencyMinutes *int            `json:"update_frequency_minutes,omitempty"`
	// URL (JSON `station_url`) is the station's page at its network.
	URL              *string  `json:"station_url,omitempty"`
	DataURL          *string  `json:"data_url,omitempty"`
	DescriptionShort *string  `json:"description_short,omitempty"`
	DescriptionLong  *string  `json:"description_long,omitempty"`
	MetaTitle        *string  `json:"meta_title,omitempty"`
	MetaDescription  *string  `json:"meta_description,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	IsFeatured       bool     `json:"is_featured,omitempty"`
}

// CatalogStationDetail is a station's full catalog record.
type CatalogStationDetail struct {
	ResponseMeta `json:"-"`
	CatalogStation

	Photos          []CatalogPhoto          `json:"photos"`
	Webcams         []CatalogWebcam         `json:"webcams"`
	PrimaryLocation *CatalogPrimaryLocation `json:"primary_location,omitempty"`
	Template        *CatalogTemplate        `json:"template,omitempty"`
	Content         []CatalogContentEntry   `json:"content,omitempty"`
	// Indexable reports whether search engines should index the station's
	// page; serve a page that is not with noindex.
	Indexable bool `json:"indexable"`
}

// CatalogPrimaryLocation is the location a station primarily serves.
type CatalogPrimaryLocation struct {
	Slug       string       `json:"slug"`
	Name       string       `json:"name"`
	Lat        float64      `json:"lat"`
	Lon        float64      `json:"lon"`
	ElevationM float64      `json:"elevation_m"`
	Type       LocationType `json:"type"`
}

// CatalogElevation is one of a catalog location's forecast elevations.
type CatalogElevation struct {
	Label string `json:"label"`
	// DisplayName is what the place is called, for display; empty where the
	// catalog has no name for the point (show Label).
	DisplayName *string `json:"display_name,omitempty"`
	// HeightM (JSON `elevation_m`) is the elevation in metres.
	HeightM     float64  `json:"elevation_m"`
	LatOverride *float64 `json:"lat_override,omitempty"`
	LonOverride *float64 `json:"lon_override,omitempty"`
	IsDefault   bool     `json:"is_default"`
}

// CatalogPhoto is a catalog photo with its credit and licence.
type CatalogPhoto struct {
	URL        string  `json:"url"`
	Caption    *string `json:"caption,omitempty"`
	Credit     *string `json:"credit,omitempty"`
	AltText    *string `json:"alt_text,omitempty"`
	License    *string `json:"license,omitempty"`
	LicenseURL *string `json:"license_url,omitempty"`
}

// CatalogWebcam is a webcam at a catalog location or station.
type CatalogWebcam struct {
	Name                   string  `json:"name"`
	URL                    string  `json:"url"`
	EmbedType              *string `json:"embed_type,omitempty"`
	Provider               *string `json:"provider,omitempty"`
	RefreshIntervalSeconds *int    `json:"refresh_interval_seconds,omitempty"`
}

// CatalogNearbyStation is an observation station near a catalog location.
type CatalogNearbyStation struct {
	Slug    string         `json:"slug"`
	Name    string         `json:"name"`
	Network StationNetwork `json:"network"`
	// Status (JSON `station_status`) is the station's status.
	Status         StationStatus `json:"station_status"`
	DistanceKm     *float64      `json:"distance_km,omitempty"`
	ElevationDiffM *float64      `json:"elevation_diff_m,omitempty"`
	IsPrimary      bool          `json:"is_primary"`
}

// CatalogTemplate is the page template of a catalog record.
type CatalogTemplate struct {
	ID           string        `json:"id"`
	LocationType *LocationType `json:"location_type"`
	// Variant (JSON `template_variant`) is the template's variant.
	Variant *TemplateVariant  `json:"template_variant"`
	Spec    []TemplateSection `json:"spec"`
}

// TemplateSection is one section of a catalog page template.
type TemplateSection struct {
	Kind   SectionKind     `json:"kind"`
	Key    string          `json:"key"`
	Config json.RawMessage `json:"config,omitempty"`
}

// CatalogContentEntry is one resolved content entry of a catalog record.
type CatalogContentEntry struct {
	ID             string          `json:"id"`
	ContentTypeKey string          `json:"content_type_key"`
	Title          string          `json:"title"`
	Body           json.RawMessage `json:"body"`
	Slot           string          `json:"slot"`
	DisplayOrder   int             `json:"display_order"`
	// Locale (JSON `resolved_locale`) is the locale the entry was resolved in.
	Locale    string    `json:"resolved_locale"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RunsResponse lists the available model runs.
type RunsResponse struct {
	ResponseMeta `json:"-"`

	// Served is the run forecasts are currently served from, or nil.
	Served *ServedRun `json:"served"`
	// Items (JSON `runs`) are the runs, newest first.
	Items []AvailableRun `json:"runs"`
	Count int            `json:"count"`
	// NextCursor is nil on the last page.
	NextCursor *string `json:"next_cursor"`
}

// ServedRun is the run currently served, with its models.
type ServedRun struct {
	InitializedAt time.Time    `json:"initialized_at"`
	Models        []ModelEntry `json:"models"`
}

// AvailableRun is one available model run.
type AvailableRun struct {
	InitializedAt time.Time `json:"initialized_at"`
}

// AttributionRegistry lists every data source with its licence and terms.
type AttributionRegistry struct {
	ResponseMeta `json:"-"`

	Providers []AttributionProvider `json:"providers"`
}

// AttributionProvider is a data provider with its licence.
type AttributionProvider struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Credit     string              `json:"credit"`
	Licence    string              `json:"licence"`
	LicenceURL string              `json:"licence_url"`
	TermsURL   string              `json:"terms_url"`
	Modified   bool                `json:"modified"`
	Notices    []string            `json:"notices"`
	Sources    []AttributionSource `json:"sources"`
}

// AttributionSource is one data source of a provider.
type AttributionSource struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Kind     SourceKind `json:"kind"`
	Provider string     `json:"provider"`
}
