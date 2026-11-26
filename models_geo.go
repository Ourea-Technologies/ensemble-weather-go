// SPDX-License-Identifier: BSD-2-Clause

// GeoJSON for maps: active alert outlines and avalanche zones. Keys are the wire's, untouched.

package ensembleweather

import (
	"encoding/json"
	"time"
)

// AlertOutlineCollection is a GeoJSON `FeatureCollection` (RFC 7946) of the active winter and wind
// alerts' outlines, with this response's `attribution` as a foreign member.
type AlertOutlineCollection struct {
	ResponseMeta `json:"-"`

	Type string `json:"type"`
	// One Feature per alert, ordered by `properties.id`.
	Features    []AlertOutlineFeature `json:"features"`
	Attribution ResponseAttribution   `json:"attribution"`
}

// AlertOutlineFeature is one alert's outline, a GeoJSON `Feature`.
type AlertOutlineFeature struct {
	Type       string                 `json:"type"`
	Geometry   AlertOutlineGeometry   `json:"geometry"`
	Properties AlertOutlineProperties `json:"properties"`
}

// AlertOutlineGeometry is a GeoJSON `Polygon` or `MultiPolygon`.
type AlertOutlineGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// AlertOutlineProperties is the alert, under the names `ActiveAlert` uses for the same fields, and
// its `class`.
type AlertOutlineProperties struct {
	// NWS alert identifier (urn:oid:…), as `ActiveAlert.id`.
	ID    string     `json:"id"`
	Event string     `json:"event"`
	Class AlertClass `json:"class"`
	// As `ActiveAlert.severity`.
	Severity AlertSeverity `json:"severity"`
	// As `ActiveAlert.urgency`.
	Urgency AlertUrgency `json:"urgency"`
	// As `ActiveAlert.certainty`.
	Certainty AlertCertainty `json:"certainty"`
	Headline  string         `json:"headline"`
	// When the hazard begins.
	Onset *time.Time `json:"onset,omitempty"`
	// When NWS is due to re-issue this message; not the alert's end (see `ActiveAlert.expires`).
	Expires time.Time `json:"expires"`
	// When the hazard ends.
	Ends *time.Time `json:"ends,omitempty"`
	// SenderName is the NWS office.
	SenderName *string `json:"sender_name,omitempty"`
	// SourceURL is the NWS forecast page of a point inside the outline, which lists the alert and
	// links to its text: where "view on NWS" goes.
	SourceURL string `json:"source_url"`
}

// AvalancheZoneCollection is a GeoJSON `FeatureCollection` (RFC 7946) of the avalanche centers'
// forecast zones, each with its published rating, and this response's `attribution` as a foreign
// member.
type AvalancheZoneCollection struct {
	ResponseMeta `json:"-"`

	Type string `json:"type"`
	// One Feature per zone, the less specific first: where zones overlap, the one drawn last is the
	// one a point there is served.
	Features    []AvalancheZoneFeature `json:"features"`
	Attribution ResponseAttribution    `json:"attribution"`
}

// AvalancheZoneFeature is one forecast zone, a GeoJSON `Feature`.
type AvalancheZoneFeature struct {
	Type string `json:"type"`
	// ID is the zone's id, as `properties.zone_id`.
	ID         int64                   `json:"id"`
	Geometry   AvalancheZoneGeometry   `json:"geometry"`
	Properties AvalancheZoneProperties `json:"properties"`
}

// AvalancheZoneGeometry is a GeoJSON `MultiPolygon`: an array of polygons, each an array of rings
// (the outer one first, then any holes), each ring closed.
type AvalancheZoneGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// AvalancheZoneProperties is the avalanche center's danger rating for the forecast zone the
// location's point is in, from avalanche.org's public map layer.
type AvalancheZoneProperties struct {
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
	// ForecastURL is the center's full forecast for the zone.
	ForecastURL string `json:"forecast_url"`
	CenterURL   string `json:"center_url"`
	// Attribution is the credit to show with the block: "Avalanche forecast by <center> via
	// avalanche.org", or "Avalanche forecast by Colorado Avalanche Information Center via
	// avalanche.state.co.us" for a CAIC area.
	Attribution string `json:"attribution"`
}
