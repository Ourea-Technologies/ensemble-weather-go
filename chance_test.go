// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"context"
	"errors"
	"math"
	"net/http"
	"testing"
)

func TestRoundThreshold(t *testing.T) {
	tests := []struct {
		v    float64
		d    int
		want string
	}{
		{0.0625, 3, "0.063"}, // banker's rounding would give 0.062
		{1.005, 2, "1.00"},   // 1.005 × 100 = 100.4999… in doubles
		{2.5, 0, "3"},
		{-2.5, 0, "-3"}, // half away from zero
		{-0.04, 1, "0.0"},
		{math.Copysign(0, -1), 2, "0.00"},
		{1524.4, 0, "1524"},
		{6 * 25.4, 2, "152.40"},
		{-5, 1, "-5.0"},
	}
	for _, tt := range tests {
		if got := formatThreshold(tt.v, tt.d); got != tt.want {
			t.Errorf("formatThreshold(%v, %d) = %q, want %q", tt.v, tt.d, got, tt.want)
		}
	}
}

func TestConvert(t *testing.T) {
	tests := []struct {
		v        float64
		from, to Unit
		want     float64
	}{
		{6, UnitInch, UnitMillimeter, 152.4},
		{25.4, UnitMillimeter, UnitInch, 1},
		{32, UnitFahrenheit, UnitCelsius, 0},
		{0, UnitCelsius, UnitFahrenheit, 32},
		{0, UnitCelsius, UnitKelvin, 273.15},
		{273.15, UnitKelvin, UnitFahrenheit, 32},
		{212, UnitFahrenheit, UnitKelvin, 373.15},
		{1000, UnitFoot, UnitMeter, 304.8},
		{10, UnitMilePerHour, UnitMeterPerSecond, 4.4704},
		{7, UnitMeter, UnitMeter, 7},
	}
	for _, tt := range tests {
		if got := convert(tt.v, tt.from, tt.to); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("convert(%v %s -> %s) = %v, want %v", tt.v, tt.from, tt.to, got, tt.want)
		}
	}
}

func TestChanceWire(t *testing.T) {
	tests := []struct {
		name    string
		units   Units
		chances []Chance
		want    string
	}{
		{"snow inches on metric", UnitsMetric, []Chance{{Variable: Snowfall, Op: AtLeast, Threshold: Inches(6)}}, "snowfall:ge:152.40:24h"},
		{"snow inches on imperial", UnitsImperial, []Chance{{Variable: Snowfall, Op: AtLeast, Threshold: Inches(6)}}, "snowfall:ge:6.000:24h"},
		{"kelvin temperature", UnitsKelvin, []Chance{{Variable: Temperature, Op: AtMost, Threshold: Celsius(0)}}, "temperature:le:273.2:24h"},
		{"unitless is request units", UnitsImperial, []Chance{{Variable: WindSpeed, Op: AtLeast, Threshold: Threshold{Value: 40}}}, "wind_speed:ge:40.0:24h"},
		{"feet on metric snow level", UnitsMetric, []Chance{{Variable: SnowLevel, Op: AtMost, Threshold: Feet(5000)}}, "snow_level:le:1524:24h"},
		{"all aggregation kept", UnitsMetric, []Chance{{Variable: Temperature, Op: AtMost, Threshold: Celsius(0), Window: HoursBetween(6, 18), Aggregation: AggregationAll}}, "temperature:le:0.0:6-18h:all"},
		{"explicit defaults omitted", UnitsMetric, []Chance{{Variable: Precipitation, Op: AtLeast, Threshold: Millimeters(10), Window: Hours(24), Aggregation: AggregationTotal}}, "precipitation:ge:10.00:24h"},
		{"sorted any before all, then threshold", UnitsMetric, []Chance{
			{Variable: Temperature, Op: AtLeast, Threshold: Celsius(5), Aggregation: AggregationAll},
			{Variable: Temperature, Op: AtLeast, Threshold: Celsius(10)},
			{Variable: Temperature, Op: AtLeast, Threshold: Celsius(-10)},
		}, "temperature:ge:-10.0:24h,temperature:ge:10.0:24h,temperature:ge:5.0:24h:all"},
		{"rounding dedups", UnitsMetric, []Chance{
			{Variable: SnowLevel, Op: AtMost, Threshold: Meters(1524.2)},
			{Variable: SnowLevel, Op: AtMost, Threshold: Meters(1523.9)},
		}, "snow_level:le:1524:24h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := buildChances("op", tt.chances, tt.units)
			if err != nil {
				t.Fatal(err)
			}
			if q.value != tt.want {
				t.Errorf("prob = %q, want %q", q.value, tt.want)
			}
			if len(q.keys) != len(tt.chances) {
				t.Errorf("%d keys for %d chances", len(q.keys), len(tt.chances))
			}
		})
	}
}

func TestChanceUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		c    Chance
	}{
		{"any on precipitation", Chance{Variable: Precipitation, Op: AtLeast, Threshold: Millimeters(1), Aggregation: AggregationAny}},
		{"all on snowfall", Chance{Variable: Snowfall, Op: AtLeast, Threshold: Inches(1), Aggregation: AggregationAll}},
		{"total on temperature", Chance{Variable: Temperature, Op: AtLeast, Threshold: Celsius(1), Aggregation: AggregationTotal}},
		{"unknown variable", Chance{Variable: "humidity", Op: AtLeast}},
		{"unknown op", Chance{Variable: Temperature, Op: "gt"}},
		{"unknown aggregation", Chance{Variable: Temperature, Op: AtLeast, Aggregation: "most"}},
		{"wrong unit dimension", Chance{Variable: Snowfall, Op: AtLeast, Threshold: Celsius(1)}},
		{"unknown unit", Chance{Variable: Snowfall, Op: AtLeast, Threshold: Threshold{1, "cm"}}},
		{"window past 360 h", Chance{Variable: Temperature, Op: AtLeast, Window: Hours(361)}},
		{"window backwards", Chance{Variable: Temperature, Op: AtLeast, Window: HoursBetween(24, 12)}},
		{"NaN threshold", Chance{Variable: Temperature, Op: AtLeast, Threshold: Celsius(math.NaN())}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rt, _ := testClient(t, okForecast)
			_, err := c.Forecast(context.Background(), "or/x", &ForecastParams{Chances: []Chance{tt.c}})
			var e *Error
			if !errors.As(err, &e) || e.Category != CategoryUsage || !errors.Is(err, ErrUsage) {
				t.Fatalf("err = %v, want a usage error", err)
			}
			if len(rt.reqs) != 0 {
				t.Errorf("sent %d requests", len(rt.reqs))
			}
		})
	}
	c, _, _ := testClient(t, okForecast)
	if _, err := c.Forecast(context.Background(), "or/x", &ForecastParams{Totals: []Total{{Variable: Temperature}}}); !errors.Is(err, ErrUsage) {
		t.Errorf("temperature total: err = %v, want usage", err)
	}
}

func TestChanceResultsMapBack(t *testing.T) {
	body := `{"units":"imperial","location":{},"model_run":{},"forecast":{},"attribution":{},"nowcast_applied":true,
	"probabilities":[
	 {"variable":"snowfall","op":"ge","threshold":6,"window_start_hours":0,"window_end_hours":24,"aggregation":"total","probability":0.4,"members":131},
	 {"variable":"temperature","op":"le","threshold":32,"window_start_hours":0,"window_end_hours":24,"aggregation":"any","probability":0.9,"members":131}]}`
	c, _, _ := testClient(t, func(int, *http.Request) (*http.Response, error) { return respond(200, body), nil })
	f, err := c.Forecast(context.Background(), "or/x", &ForecastParams{Units: UnitsImperial, Chances: []Chance{
		{Variable: Temperature, Op: AtMost, Threshold: Fahrenheit(32)},
		{Variable: Snowfall, Op: AtLeast, Threshold: Inches(6)},
		{Variable: Temperature, Op: AtMost, Threshold: Celsius(0)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Probabilities) != 3 || *f.Probabilities[0].Probability != 0.9 || *f.Probabilities[1].Probability != 0.4 || *f.Probabilities[2].Probability != 0.9 {
		t.Fatalf("probabilities = %+v", f.Probabilities)
	}
	if f.Probabilities[1].ThresholdUnit != UnitInch || f.Probabilities[0].ThresholdUnit != UnitFahrenheit {
		t.Errorf("threshold units = %s, %s", f.Probabilities[1].ThresholdUnit, f.Probabilities[0].ThresholdUnit)
	}
	// An answer missing from a non-empty list is an error, never a guess.
	_, err = c.Forecast(context.Background(), "or/x", &ForecastParams{Units: UnitsImperial, Chances: []Chance{
		{Variable: WindSpeed, Op: AtLeast, Threshold: MilesPerHour(40)},
	}})
	if !errors.Is(err, ErrUnavailable) {
		t.Errorf("unanswered term: err = %v, want unavailable", err)
	}
}

func TestUnitsFor(t *testing.T) {
	tests := []struct {
		u    Units
		v    Variable
		want Unit
	}{
		{UnitsMetric, Temperature, UnitCelsius}, {UnitsImperial, Temperature, UnitFahrenheit}, {UnitsKelvin, Temperature, UnitKelvin},
		{UnitsMetric, Snowfall, UnitMillimeter}, {UnitsImperial, Precipitation, UnitInch}, {UnitsKelvin, Snowfall, UnitMillimeter},
		{UnitsImperial, SnowLevel, UnitFoot}, {UnitsKelvin, SnowLevel, UnitMeter}, {UnitsImperial, WindSpeed, UnitMilePerHour},
		{"furlongs", WindSpeed, ""}, {UnitsMetric, "humidity", ""},
	}
	for _, tt := range tests {
		if got := tt.u.For(tt.v); got != tt.want {
			t.Errorf("%s.For(%s) = %q, want %q", tt.u, tt.v, got, tt.want)
		}
	}
	if UnitCelsius.Symbol() != "°C" || UnitInch.Symbol() != "in" {
		t.Error("Symbol")
	}
}

func TestQuantileViews(t *testing.T) {
	q := Quantiles{P10: Ptr(1.0), P25: Ptr(2.0), P50: Ptr(3.0), P90: Ptr(5.0)}
	if r, ok := q.Range80(); !ok || r != (Interval{1, 5}) || r.String() != "1–5" {
		t.Errorf("Range80 = %v, %v", r, ok)
	}
	if _, ok := q.Range50(); ok {
		t.Error("Range50 with p75 absent reported ok")
	}
	if m, ok := q.Median(); !ok || m != 3 {
		t.Errorf("Median = %v, %v", m, ok)
	}
	if _, ok := (Quantiles{}).Median(); ok {
		t.Error("Median of empty quantiles reported ok")
	}
}
