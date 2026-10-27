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

// TestBodiesDecodeStrictly decodes every vector body into its model with
// unknown fields refused, so a field the models lack fails here. The two
// forward-compatibility bodies are decoded leniently, as the client does.
func TestBodiesDecodeStrictly(t *testing.T) {
	models := map[string]func() any{
		"forecast-":      func() any { return new(ForecastResponse) },
		"search":         func() any { return new(SearchLocationsResponse) },
		"catalog-page-":  func() any { return new(CatalogLocationPage) },
		"catalog-detail": func() any { return new(CatalogLocationDetail) },
		"attribution":    func() any { return new(AttributionRegistry) },
		"error-":         func() any { return new(errorBody) },
	}
	lenient := map[string]bool{"forecast-unknown-fields.json": true, "error-unknown-code.json": true}
	files, err := filepath.Glob(filepath.Join(vectorsDir, "bodies", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no bodies: %v", err)
	}
	for _, f := range files {
		name := filepath.Base(f)
		var mk func() any
		for prefix, m := range models {
			if strings.HasPrefix(name, prefix) {
				mk = m
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
