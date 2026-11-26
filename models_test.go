// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBodiesDecodeStrictly decodes every vector body into its model (the
// longest matching file-name prefix) with
// unknown fields refused, so a field the models lack fails here. The two
// forward-compatibility bodies are decoded leniently, as the client does.
func TestBodiesDecodeStrictly(t *testing.T) {
	models := map[string]func() any{
		"forecast-":              func() any { return new(ForecastResponse) },
		"search":                 func() any { return new(SearchLocationsResponse) },
		"catalog-page-":          func() any { return new(CatalogLocationPage) },
		"catalog-detail":         func() any { return new(CatalogLocationDetail) },
		"attribution":            func() any { return new(AttributionRegistry) },
		"error-":                 func() any { return new(errorBody) },
		"alert-outlines":         func() any { return new(AlertOutlineCollection) },
		"alerts":                 func() any { return new(AlertsResponse) },
		"avalanche-zones":        func() any { return new(AvalancheZoneCollection) },
		"catalog-station-detail": func() any { return new(CatalogStationDetail) },
		"composition":            func() any { return new(CompositionResponse) },
		"current-":               func() any { return new(CurrentResponse) },
		"explore":                func() any { return new(ExploreResponse) },
		"locations-page":         func() any { return new(LocationsResponse) },
		"nearby-point":           func() any { return new(NearbyPointResponse) },
		"nearby":                 func() any { return new(NearbyLocationsResponse) },
		"runs":                   func() any { return new(RunsResponse) },
		"season":                 func() any { return new(SeasonResponse) },
		"snow-trend":             func() any { return new(SnowTrendResponse) },
		"snowpack-plume":         func() any { return new(SnowpackPlumeResponse) },
		"snowpack":               func() any { return new(SnowpackResponse) },
	}
	lenient := map[string]bool{"forecast-unknown-fields.json": true, "error-unknown-code.json": true}
	files, err := filepath.Glob(filepath.Join(vectorsDir, "bodies", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no bodies: %v", err)
	}
	for _, f := range files {
		name := filepath.Base(f)
		var mk func() any
		longest := 0
		for prefix, m := range models {
			if strings.HasPrefix(name, prefix) && len(prefix) > longest {
				mk, longest = m, len(prefix)
			}
		}
		if mk == nil {
			t.Errorf("%s: no model mapped", name)
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		dec := json.NewDecoder(bytes.NewReader(raw))
		if !lenient[name] {
			dec.DisallowUnknownFields()
		}
		if err := dec.Decode(mk()); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
