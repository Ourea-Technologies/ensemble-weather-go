// SPDX-License-Identifier: BSD-2-Clause

package spectest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	ew "oureawx.com/go/ensemble-weather"
)

const specPath = "../spec/openapi/v1/openapi.yaml"

// outOfScope are the operations the SDK deliberately does not expose
// (BEHAVIOR 3: getMapsToken is a web-map credential, out of scope for 0.x).
var outOfScope = map[string]bool{"getMapsToken": true}

// staleExamples are spec response examples that do not match their own
// schema in the pinned spec (scripts/validate.py in ensemble-weather-spec
// reports the same ones). They are still round-tripped leniently; the strict
// check expects them to fail and fails when they are fixed. The pinned spec
// has none.
var staleExamples = map[string]bool{}

type operationCase struct {
	method string // the *Client method that exposes it
	result func() any
	call   func(ctx context.Context, c *ew.Client) error
}

func ignore[T any](_ T, err error) error { return err }

// operations maps every in-scope operationId to the SDK method exposing it.
var operations = map[string]operationCase{
	"getLocationForecast": {"Forecast", func() any { return new(ew.ForecastResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Forecast(ctx, "or/mt-hood-meadows", nil))
	}},
	"getPointForecast": {"GetPointForecast", func() any { return new(ew.ForecastResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.GetPointForecast(ctx, ew.Point{Lat: 45.33, Lon: -121.71}, nil))
	}},
	"getCurrentConditions": {"Current", func() any { return new(ew.CurrentResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Current(ctx, "or/mt-hood-meadows", nil))
	}},
	"getCurrentByCoordinates": {"CurrentAt", func() any { return new(ew.CurrentResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.CurrentAt(ctx, ew.Point{Lat: 45.33, Lon: -121.71}, nil))
	}},
	"getLocationAlerts": {"Alerts", func() any { return new(ew.AlertsResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Alerts(ctx, "or/mt-hood-meadows"))
	}},
	"getAlertsByCoordinates": {"AlertsAt", func() any { return new(ew.AlertsResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.AlertsAt(ctx, ew.Point{Lat: 45.33, Lon: -121.71}))
	}},
	"getNearbyLocations": {"Nearby", func() any { return new(ew.NearbyLocationsResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Nearby(ctx, "or/mt-hood-meadows", nil))
	}},
	"getNearbyByCoordinates": {"NearbyAt", func() any { return new(ew.NearbyPointResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.NearbyAt(ctx, ew.Point{Lat: 45.33, Lon: -121.71}, nil))
	}},
	"searchLocations": {"Search", func() any { return new(ew.SearchLocationsResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Search(ctx, "hood", nil))
	}},
	"getAllLocations": {"Locations", func() any { return new(ew.LocationsResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Locations(ctx, nil))
	}},
	"getLocations": {"Locations", func() any { return new(ew.LocationsResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Locations(ctx, &ew.LocationsParams{State: "or"}))
	}},
	"listCatalogLocations": {"CatalogLocations", func() any { return new(ew.CatalogLocationPage) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.CatalogLocations(ctx, nil))
	}},
	"getCatalogLocation": {"CatalogLocation", func() any { return new(ew.CatalogLocationDetail) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.CatalogLocation(ctx, "mt-hood-meadows", nil))
	}},
	"listCatalogStations": {"CatalogStations", func() any { return new(ew.CatalogStationPage) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.CatalogStations(ctx, nil))
	}},
	"getCatalogStation": {"CatalogStation", func() any { return new(ew.CatalogStationDetail) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.CatalogStation(ctx, "x", nil))
	}},
	"getAvailableRuns": {"Runs", func() any { return new(ew.RunsResponse) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Runs(ctx, 0))
	}},
	"getAttribution": {"Attribution", func() any { return new(ew.AttributionRegistry) }, func(ctx context.Context, c *ew.Client) error {
		return ignore(c.Attribution(ctx))
	}},
}

// schemas maps every component schema the operations reach to its model.
var schemas = map[string]reflect.Type{
	"ActiveAlert": reflect.TypeOf(ew.ActiveAlert{}), "LocationRef": reflect.TypeOf(ew.LocationRef{}),
	"CurrentResponse": reflect.TypeOf(ew.CurrentResponse{}), "AlertsResponse": reflect.TypeOf(ew.AlertsResponse{}),
	"SearchLocationsResponse": reflect.TypeOf(ew.SearchLocationsResponse{}), "PublicCatalogLocationPage": reflect.TypeOf(ew.CatalogLocationPage{}),
	"PublicCatalogStationPage": reflect.TypeOf(ew.CatalogStationPage{}), "RunsResponse": reflect.TypeOf(ew.RunsResponse{}),
	"NearbyPointResponse": reflect.TypeOf(ew.NearbyPointResponse{}),
	"LocationsResponse":   reflect.TypeOf(ew.LocationsResponse{}), "NearbyLocationsResponse": reflect.TypeOf(ew.NearbyLocationsResponse{}),
	"NearbyLocation": reflect.TypeOf(ew.NearbyLocation{}), "LocationSummary": reflect.TypeOf(ew.LocationSummary{}),
	"ProbabilityResult": reflect.TypeOf(ew.ProbabilityResult{}), "WindowTotal": reflect.TypeOf(ew.WindowTotal{}),
	"Quantiles": reflect.TypeOf(ew.Quantiles{}), "PrecipDisplay": reflect.TypeOf(ew.PrecipDisplay{}),
	"ProbabilityThresholds": reflect.TypeOf(ew.ProbabilityThresholds{}), "AppliedOptions": reflect.TypeOf(ew.AppliedOptions{}),
	"ForecastResponse": reflect.TypeOf(ew.ForecastResponse{}), "ResponseAttribution": reflect.TypeOf(ew.ResponseAttribution{}),
	"AttributionRegistry": reflect.TypeOf(ew.AttributionRegistry{}), "AttributionProvider": reflect.TypeOf(ew.AttributionProvider{}),
	"AttributionSource": reflect.TypeOf(ew.AttributionSource{}), "SensorObservation": reflect.TypeOf(ew.SensorObservation{}),
	"CurrentConditions": reflect.TypeOf(ew.CurrentConditions{}), "PointInfo": reflect.TypeOf(ew.PointInfo{}),
	"NearestSensor": reflect.TypeOf(ew.NearestSensor{}), "ModelRunInfo": reflect.TypeOf(ew.ModelRunInfo{}),
	"ModelEntry": reflect.TypeOf(ew.ModelEntry{}), "ForecastData": reflect.TypeOf(ew.ForecastData{}),
	"SnowOutlook": reflect.TypeOf(ew.SnowOutlook{}), "SnowOutlookQuantiles": reflect.TypeOf(ew.SnowOutlookQuantiles{}),
	"TimeSeriesPoint": reflect.TypeOf(ew.TimeSeriesPoint{}), "HourProvenance": reflect.TypeOf(ew.HourProvenance{}),
	"Sector": reflect.TypeOf(ew.Sector{}), "CircularStats": reflect.TypeOf(ew.CircularStats{}),
	"MemberForecast": reflect.TypeOf(ew.MemberForecast{}), "MembersCumulative": reflect.TypeOf(ew.MembersCumulative{}),
	"InversionInfo": reflect.TypeOf(ew.InversionInfo{}), "PrecipType": reflect.TypeOf(ew.PrecipType{}),
	"EnsembleStats": reflect.TypeOf(ew.EnsembleStats{}), "FreezingLevel": reflect.TypeOf(ew.FreezingLevel{}),
	"ModelMeans": reflect.TypeOf([]ew.ModelMean{}), "ModelStats": reflect.TypeOf(ew.ModelMean{}),
	"DailyForecast": reflect.TypeOf(ew.DailyForecast{}), "DailyAstro": reflect.TypeOf(ew.DailyAstro{}),
	"DayClimatology": reflect.TypeOf(ew.DayClimatology{}), "ElevationForecast": reflect.TypeOf(ew.ElevationForecast{}),
	"SearchHit":             reflect.TypeOf(ew.SearchHit{}),
	"PublicCatalogLocation": reflect.TypeOf(ew.CatalogLocation{}), "PublicCatalogLocationDetail": reflect.TypeOf(ew.CatalogLocationDetail{}),
	"PublicCatalogStation": reflect.TypeOf(ew.CatalogStation{}), "PublicCatalogStationDetail": reflect.TypeOf(ew.CatalogStationDetail{}),
	"PublicCatalogPrimaryLocation": reflect.TypeOf(ew.CatalogPrimaryLocation{}), "PublicCatalogElevation": reflect.TypeOf(ew.CatalogElevation{}),
	"PublicCatalogPhoto": reflect.TypeOf(ew.CatalogPhoto{}), "PublicCatalogWebcam": reflect.TypeOf(ew.CatalogWebcam{}),
	"PublicCatalogNearbyStation": reflect.TypeOf(ew.CatalogNearbyStation{}), "CmsTemplateSection": reflect.TypeOf(ew.TemplateSection{}),
	"CatalogTemplateSummary": reflect.TypeOf(ew.CatalogTemplate{}), "CatalogResolvedEntry": reflect.TypeOf(ew.CatalogContentEntry{}),
	"GapFlow": reflect.TypeOf(ew.GapFlow{}), "DailyWind": reflect.TypeOf(ew.DailyWind{}),
	"NearestLocation": reflect.TypeOf(ew.NearestLocation{}), "ConditionIcon": reflect.TypeOf(ew.ConditionIcon("")),
	"NullReason": reflect.TypeOf(ew.NullReason("")),
}

// notModelled are component schemas with no public model of their own:
// the error body becomes *ew.Error, the unit system and freezing-level
// variants are expressed by ew.Units and ew.FreezingLevel, and a next
// cursor is a *string.
var notModelled = map[string]bool{
	"ErrorResponse": true, "UnitSystem": true, "NextCursor": true,
	"FreezingLevelBelowSurface": true, "FreezingLevelHeight": true, "FreezingLevelAboveProfile": true,
}

type spec struct {
	Components struct {
		Schemas map[string]map[string]any `yaml:"schemas"`
	} `yaml:"components"`
}

func loadSpec(t *testing.T) (*spec, map[string]any) {
	t.Helper()
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	var s spec
	if err := yaml.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	var generic map[string]any
	if err := yaml.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	return &s, generic
}

type specOp struct {
	id        string
	responses map[string]any
}

func specOperations(t *testing.T, generic map[string]any) []specOp {
	t.Helper()
	var ops []specOp
	for _, item := range generic["paths"].(map[string]any) {
		for _, raw := range item.(map[string]any) {
			op, ok := raw.(map[string]any)
			if !ok || op["operationId"] == nil {
				continue
			}
			resp, _ := op["responses"].(map[string]any)
			ops = append(ops, specOp{id: op["operationId"].(string), responses: resp})
		}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].id < ops[j].id })
	return ops
}

func TestEveryOperationExposed(t *testing.T) {
	_, generic := loadSpec(t)
	client := reflect.TypeOf(&ew.Client{})
	seen := map[string]bool{}
	for _, op := range specOperations(t, generic) {
		seen[op.id] = true
		if outOfScope[op.id] {
			continue
		}
		oc, ok := operations[op.id]
		if !ok {
			t.Errorf("the spec has operation %s, which the SDK does not expose", op.id)
			continue
		}
		if _, ok := client.MethodByName(oc.method); !ok {
			t.Errorf("%s: *Client has no method %s", op.id, oc.method)
		}
	}
	for id := range operations {
		if !seen[id] {
			t.Errorf("the SDK maps %s, which the spec no longer has", id)
		}
	}
}

// resolve follows $ref and merges allOf/oneOf/anyOf branches into one set of
// properties.
func resolve(all map[string]map[string]any, s map[string]any) (props map[string]any, items map[string]any) {
	props = map[string]any{}
	var walk func(s map[string]any)
	walk = func(s map[string]any) {
		if s == nil {
			return
		}
		if ref, ok := s["$ref"].(string); ok {
			walk(all[ref[strings.LastIndex(ref, "/")+1:]])
			return
		}
		for _, k := range []string{"allOf", "oneOf", "anyOf"} {
			if branches, ok := s[k].([]any); ok {
				for _, b := range branches {
					walk(b.(map[string]any))
				}
			}
		}
		if p, ok := s["properties"].(map[string]any); ok {
			for k, v := range p {
				props[k] = v
			}
		}
		if it, ok := s["items"].(map[string]any); ok {
			items = it
		}
	}
	walk(s)
	return props, items
}

func jsonFields(t reflect.Type) map[string]reflect.Type {
	out := map[string]reflect.Type{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if f.Anonymous {
			if tag == "-" {
				continue
			}
			for k, v := range jsonFields(f.Type) {
				out[k] = v
			}
			continue
		}
		if tag == "" || tag == "-" || !f.IsExported() {
			continue
		}
		out[tag] = f.Type
	}
	return out
}

var rawMessage = reflect.TypeOf(json.RawMessage{})

func elem(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer || (t.Kind() == reflect.Slice && t != rawMessage) {
		t = t.Elem()
	}
	return t
}

// checkModel asserts that every property of schema s is a JSON field of t,
// recursing into nested objects.
func checkModel(t *testing.T, all map[string]map[string]any, path string, s map[string]any, gt reflect.Type, seen map[string]bool) {
	t.Helper()
	gt = elem(gt)
	key := path + "|" + gt.String()
	if seen[key] || gt == rawMessage || gt.Kind() != reflect.Struct || gt.PkgPath() == "time" {
		return
	}
	seen[key] = true
	props, items := resolve(all, s)
	if len(props) == 0 && items != nil {
		props, _ = resolve(all, items)
	}
	fields := jsonFields(gt)
	for name, ps := range props {
		ft, ok := fields[name]
		if !ok {
			t.Errorf("%s.%s: the spec property has no field in %s", path, name, gt)
			continue
		}
		checkModel(t, all, path+"."+name, ps.(map[string]any), ft, seen)
	}
}

// checkNoExtraFields asserts that every JSON field of t is a property of
// schema s (a field the spec no longer has is a renamed or withdrawn one),
// recursing into nested objects.
func checkNoExtraFields(t *testing.T, all map[string]map[string]any, path string, s map[string]any, gt reflect.Type, seen map[string]bool) {
	t.Helper()
	gt = elem(gt)
	key := path + "|" + gt.String()
	if seen[key] || gt == rawMessage || gt.Kind() != reflect.Struct || gt.PkgPath() == "time" {
		return
	}
	seen[key] = true
	props, items := resolve(all, s)
	if len(props) == 0 && items != nil {
		props, _ = resolve(all, items)
	}
	for name, ft := range jsonFields(gt) {
		ps, ok := props[name]
		if !ok {
			t.Errorf("%s.%s: %s has a field the spec does not", path, name, gt)
			continue
		}
		checkNoExtraFields(t, all, path+"."+name, ps.(map[string]any), ft, seen)
	}
}

func TestNoModelFieldOutsideSpec(t *testing.T) {
	s, _ := loadSpec(t)
	all := s.Components.Schemas
	seen := map[string]bool{}
	for name, gt := range schemas {
		if name == "ModelStats" {
			continue // ew.ModelMean carries ModelMeans' id too; checked there
		}
		schema, ok := all[name]
		if !ok {
			t.Errorf("the SDK models %s, which the spec no longer has", name)
			continue
		}
		checkNoExtraFields(t, all, name, schema, gt, seen)
	}
}

func TestEverySchemaPropertyModelled(t *testing.T) {
	s, _ := loadSpec(t)
	all := s.Components.Schemas
	seen := map[string]bool{}
	for name, schema := range all {
		gt, ok := schemas[name]
		switch {
		case notModelled[name]:
			continue
		case !ok:
			t.Errorf("component schema %s has no model", name)
			continue
		}
		checkModel(t, all, name, schema, gt, seen)
	}
}

// subset reports where got (the model's re-encoding) differs from want (the
// example): every value the model kept must equal the example's.
func subset(path string, got, want any) []string {
	switch g := got.(type) {
	case map[string]any:
		w, ok := want.(map[string]any)
		if !ok {
			return []string{path + ": object vs " + fmt.Sprintf("%T", want)}
		}
		var diffs []string
		for k, v := range g {
			wv, ok := w[k]
			if !ok {
				if isZero(v) {
					continue
				}
				diffs = append(diffs, path+"."+k+": added by the model")
				continue
			}
			diffs = append(diffs, subset(path+"."+k, v, wv)...)
		}
		return diffs
	case []any:
		w, ok := want.([]any)
		if !ok || len(w) != len(g) {
			return []string{path + ": array length differs"}
		}
		var diffs []string
		for i := range g {
			diffs = append(diffs, subset(fmt.Sprintf("%s[%d]", path, i), g[i], w[i])...)
		}
		return diffs
	default:
		if !reflect.DeepEqual(got, want) {
			return []string{fmt.Sprintf("%s: %v, want %v", path, got, want)}
		}
	}
	return nil
}

// isZero reports a value the model writes for a required field the example
// left out (a zero number, empty string, false, or an empty object).
func isZero(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case float64:
		return x == 0
	case string:
		return x == "" || x == "0001-01-01T00:00:00Z"
	case bool:
		return !x
	case []any:
		return len(x) == 0
	case map[string]any:
		for _, e := range x {
			if !isZero(e) {
				return false
			}
		}
		return true
	}
	return false
}

func toJSON(v any) any {
	b, _ := json.Marshal(v)
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}

type example struct {
	op, status, name string
	value            any
}

func examples(t *testing.T, generic map[string]any) []example {
	var out []example
	for _, op := range specOperations(t, generic) {
		for status, raw := range op.responses {
			r, _ := raw.(map[string]any)
			content, _ := r["content"].(map[string]any)
			js, _ := content["application/json"].(map[string]any)
			if js == nil {
				continue
			}
			if v, ok := js["example"]; ok {
				out = append(out, example{op.id, status, "example", v})
			}
			exs, _ := js["examples"].(map[string]any)
			for name, e := range exs {
				out = append(out, example{op.id, status, name, e.(map[string]any)["value"]})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].op+out[i].status+out[i].name < out[j].op+out[j].status+out[j].name
	})
	return out
}

func TestResponseExamplesRoundTrip(t *testing.T) {
	_, generic := loadSpec(t)
	n := 0
	for _, ex := range examples(t, generic) {
		if outOfScope[ex.op] || !strings.HasPrefix(ex.status, "2") {
			continue
		}
		n++
		label := ex.op + " " + ex.status + " " + ex.name
		t.Run(label, func(t *testing.T) {
			raw, err := json.Marshal(ex.value)
			if err != nil {
				t.Fatal(err)
			}
			model := operations[ex.op].result()
			if err := json.Unmarshal(raw, model); err != nil {
				t.Fatalf("decode: %v", err)
			}
			for _, d := range subset("$", toJSON(model), toJSON(ex.value)) {
				t.Errorf("round trip: %s", d)
			}
			dec := json.NewDecoder(bytes.NewReader(raw))
			dec.DisallowUnknownFields()
			strictErr := dec.Decode(operations[ex.op].result())
			switch {
			case staleExamples[label] && strictErr == nil:
				t.Errorf("this example now matches its schema: remove it from staleExamples")
			case staleExamples[label]:
				t.Logf("known stale spec example (field outside its schema): %v", strictErr)
			case strictErr != nil:
				t.Errorf("strict decode: %v", strictErr)
			}
		})
	}
	if n == 0 {
		t.Fatal("no response examples found")
	}
}

type fixed struct {
	status int
	body   []byte
}

func (f fixed) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: f.status, Header: http.Header{"Content-Type": {"application/json"}},
		Body: io.NopCloser(bytes.NewReader(f.body))}, nil
}

// TestErrorExamplesBecomeTypedErrors sends every error example through the
// client and checks that it becomes an *ew.Error with the example's code.
func TestErrorExamplesBecomeTypedErrors(t *testing.T) {
	_, generic := loadSpec(t)
	for _, ex := range examples(t, generic) {
		if outOfScope[ex.op] || strings.HasPrefix(ex.status, "2") {
			continue
		}
		t.Run(ex.op+" "+ex.status+" "+ex.name, func(t *testing.T) {
			var status int
			if _, err := fmt.Sscan(ex.status, &status); err != nil {
				t.Skipf("status %q is not a number", ex.status)
			}
			body, _ := json.Marshal(ex.value)
			c := ew.NewClient(ew.WithAPIKey("ow_test"), ew.WithMaxRetries(0), ew.WithCache(nil),
				ew.WithHTTPClient(&http.Client{Transport: fixed{status, body}}))
			err := operations[ex.op].call(context.Background(), c)
			var e *ew.Error
			if !errors.As(err, &e) {
				t.Fatalf("err = %v, want *ew.Error", err)
			}
			want, _ := ex.value.(map[string]any)["error"].(string)
			if string(e.Code) != want || e.Status != status || e.Category == ew.CategoryUsage {
				t.Errorf("error = %+v, want code %s status %d", e, want, status)
			}
		})
	}
}
