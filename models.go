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
}

// CurrentResponse is current conditions at a location or point.
type CurrentResponse struct {
	ResponseMeta `json:"-"`

	Location    LocationRef         `json:"location"`
	Alerts      []ActiveAlert       `json:"alerts"`
	AlertsAsOf  time.Time           `json:"alerts_as_of"`
	SensorCount int                 `json:"sensor_count"`
	Sensors     []SensorObservation `json:"sensors"`
	// Current is nil when no served hour is near the request time.
	Current               *CurrentConditions `json:"current"`
	AsOf                  time.Time          `json:"as_of"`
	ModelRunInitializedAt *time.Time         `json:"model_run_initialized_at"`
	// NowcastHours is present only with IncludeNowcastHours.
	NowcastHours       []TimeSeriesPoint   `json:"nowcast_hours"`
	Units              Units               `json:"units"`
	Attribution        ResponseAttribution `json:"attribution"`
	NowcastUnavailable []NowcastSource     `json:"nowcast_unavailable,omitempty"`
}

// CurrentConditions is "right now": a view of one served forecast hour.
type CurrentConditions struct {
	Time                     time.Time       `json:"time"`
	Temperature              float64         `json:"temperature"`
	FeelsLike                *float64        `json:"feels_like,omitempty"`
	Condition                Condition       `json:"condition"`
	ConditionText            string          `json:"condition_text"`
	ConditionIcon            ConditionIcon   `json:"condition_icon"`
	WindSpeed                float64         `json:"wind_speed"`
	WindGust                 *float64        `json:"wind_gust,omitempty"`
	WindDirection            float64         `json:"wind_direction"`
	RelativeHumidity         *float64        `json:"relative_humidity,omitempty"`
	Dewpoint                 *float64        `json:"dewpoint,omitempty"`
	PrecipitationProbability *float64        `json:"precipitation_probability,omitempty"`
	Source                   CurrentSource   `json:"source"`
	Station                  *CurrentStation `json:"station,omitempty"`
}

// CurrentStation is the surface station that anchored the current hour.
type CurrentStation struct {
	// StationID is the station's id, as in [SensorObservation.StationID].
	StationID       string    `json:"station_id"`
	Name            string    `json:"name"`
	ObservationTime time.Time `json:"observation_time"`
}

// SensorObservation is one sensor's latest reading, in the request's units.
type SensorObservation struct {
	StationID          string    `json:"station_id"`
	Name               string    `json:"name"`
	Source             string    `json:"source"`
	ObservationTime    time.Time `json:"observation_time"`
	Temperature        *float64  `json:"temperature,omitempty"`
	Dewpoint           *float64  `json:"dewpoint,omitempty"`
	RelativeHumidity   *float64  `json:"relative_humidity,omitempty"`
	WindSpeed          *float64  `json:"wind_speed,omitempty"`
	WindGust           *float64  `json:"wind_gust,omitempty"`
	WindDirection      *float64  `json:"wind_direction,omitempty"`
	Pressure           *float64  `json:"pressure,omitempty"`
	Precipitation      *float64  `json:"precipitation,omitempty"`
	PrecipitationRate  *float64  `json:"precipitation_rate,omitempty"`
	Reflectivity       *float64  `json:"reflectivity,omitempty"`
	PavementTemp       *float64  `json:"pavement_temp,omitempty"`
	SubsurfaceTemp     *float64  `json:"subsurface_temp,omitempty"`
	Visibility         *float64  `json:"visibility,omitempty"`
	CloudCover         *float64  `json:"cloud_cover,omitempty"`
	CeilingHeight      *float64  `json:"ceiling_height,omitempty"`
	SnowDepth          *float64  `json:"snow_depth,omitempty"`
	SWE                *float64  `json:"swe,omitempty"`
	SWEPercentOfMedian *float64  `json:"swe_percent_of_median,omitempty"`
	Snowfall           *float64  `json:"snowfall,omitempty"`
	SolarRadiation     *float64  `json:"solar_radiation,omitempty"`
	PavementCondition  *string   `json:"pavement_condition,omitempty"`
}

// AlertsResponse lists the alerts in effect at a location or point.
type AlertsResponse struct {
	ResponseMeta `json:"-"`

	Location    LocationRef         `json:"location"`
	AlertCount  int                 `json:"alert_count"`
	AsOf        time.Time           `json:"as_of"`
	Alerts      []ActiveAlert       `json:"alerts"`
	Attribution ResponseAttribution `json:"attribution"`
}

// LocationSummary is a named location in a list.
type LocationSummary struct {
	LocationRef
	County      string           `json:"county"`
	Description *string          `json:"description,omitempty"`
	Elevations  []NamedElevation `json:"elevations,omitempty"`
}

// NamedElevation is one of a location's forecast elevations, in the
// request's units (m or ft).
type NamedElevation struct {
	Name      string  `json:"name"`
	Elevation float64 `json:"elevation"`
	Default   bool    `json:"default,omitempty"`
}

// NearbyLocation is a location near another, with its distance.
type NearbyLocation struct {
	LocationSummary
	DistanceKm float64 `json:"distance_km"`
}

// NearbyLocationsResponse lists the locations near a named location.
type NearbyLocationsResponse struct {
	ResponseMeta `json:"-"`

	Location  LocationSummary  `json:"location"`
	RadiusKm  float64          `json:"radius_km"`
	Count     int              `json:"count"`
	Locations []NearbyLocation `json:"locations"`
}

// NearbyPointResponse lists the locations near a point.
type NearbyPointResponse struct {
	ResponseMeta `json:"-"`

	Lat       float64          `json:"lat"`
	Lon       float64          `json:"lon"`
	RadiusKm  float64          `json:"radius_km"`
	Count     int              `json:"count"`
	Locations []NearbyLocation `json:"locations"`
}

// LocationsResponse is one page of a state's (or every state's) locations.
type LocationsResponse struct {
	ResponseMeta `json:"-"`

	State     string            `json:"state"`
	Count     int               `json:"count"`
	Locations []LocationSummary `json:"locations"`
	// NextCursor is nil on the last page.
	NextCursor *string `json:"next_cursor"`
}

// SearchLocationsResponse holds location search hits.
type SearchLocationsResponse struct {
	ResponseMeta `json:"-"`

	Items []SearchHit `json:"items"`
	Count int         `json:"count"`
}

// SearchHit is one location search result. Key is its "state/slug".
type SearchHit struct {
	Key string `json:"key"`
	LocationRef
	County string `json:"county"`
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
	County                 *string         `json:"county,omitempty"`
	Country                string          `json:"country,omitempty"`
	Aliases                []string        `json:"aliases,omitempty"`
	Network                StationNetwork  `json:"network,omitempty"`
	SourceStationID        string          `json:"source_station_id,omitempty"`
	ElevationMslM          *float64        `json:"elevation_msl_m,omitempty"`
	SensorHeightAglM       *float64        `json:"sensor_height_agl_m,omitempty"`
	SensorCapabilities     []string        `json:"sensor_capabilities,omitempty"`
	StationStatus          StationStatus   `json:"station_status,omitempty"`
	LastReportAt           *time.Time      `json:"last_report_at,omitempty"`
	LastReportSummary      json.RawMessage `json:"last_report_summary,omitempty"`
	InstallDate            *string         `json:"install_date,omitempty"`
	DecommissionDate       *string         `json:"decommission_date,omitempty"`
	UpdateFrequencyMinutes *int            `json:"update_frequency_minutes,omitempty"`
	StationURL             *string         `json:"station_url,omitempty"`
	DataURL                *string         `json:"data_url,omitempty"`
	DescriptionShort       *string         `json:"description_short,omitempty"`
	DescriptionLong        *string         `json:"description_long,omitempty"`
	MetaTitle              *string         `json:"meta_title,omitempty"`
	MetaDescription        *string         `json:"meta_description,omitempty"`
	Tags                   []string        `json:"tags,omitempty"`
	IsFeatured             bool            `json:"is_featured,omitempty"`
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
	Label       string   `json:"label"`
	ElevationM  float64  `json:"elevation_m"`
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
	Slug           string         `json:"slug"`
	Name           string         `json:"name"`
	Network        StationNetwork `json:"network"`
	StationStatus  StationStatus  `json:"station_status"`
	DistanceKm     *float64       `json:"distance_km,omitempty"`
	ElevationDiffM *float64       `json:"elevation_diff_m,omitempty"`
	IsPrimary      bool           `json:"is_primary"`
}

// CatalogTemplate is the page template of a catalog record.
type CatalogTemplate struct {
	ID              string            `json:"id"`
	LocationType    *LocationType     `json:"location_type"`
	TemplateVariant *TemplateVariant  `json:"template_variant"`
	Spec            []TemplateSection `json:"spec"`
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
	ResolvedLocale string          `json:"resolved_locale"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// RunsResponse lists the available model runs.
type RunsResponse struct {
	ResponseMeta `json:"-"`

	// Served is the run forecasts are currently served from, or nil.
	Served *ServedRun     `json:"served"`
	Runs   []AvailableRun `json:"runs"`
	Count  int            `json:"count"`
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
