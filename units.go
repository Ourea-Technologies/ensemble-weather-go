// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

// Units is a unit system for a request and its response. The server converts
// every value; the SDK never converts response values.
type Units string

// The unit systems.
const (
	// UnitsMetric: °C, mm, m/s, m. The API default.
	UnitsMetric Units = "metric"
	// UnitsImperial: °F, in, mph, ft.
	UnitsImperial Units = "imperial"
	// UnitsKelvin: metric, with temperatures in K.
	UnitsKelvin Units = "kelvin"
)

// IsKnown reports whether u is one of the unit systems this SDK knows.
// A value the API adds later parses as itself and reports false.
func (u Units) IsKnown() bool {
	return u == UnitsMetric || u == UnitsImperial || u == UnitsKelvin
}

// For returns the unit the server uses for variable v in this unit system,
// for example [UnitInch] for [Snowfall] in [UnitsImperial]. It returns "" for
// an unknown unit system or variable.
func (u Units) For(v Variable) Unit {
	imperial := u == UnitsImperial
	if !u.IsKnown() {
		return ""
	}
	switch v {
	case Temperature:
		switch u {
		case UnitsImperial:
			return UnitFahrenheit
		case UnitsKelvin:
			return UnitKelvin
		default:
			return UnitCelsius
		}
	case Precipitation, Snowfall:
		if imperial {
			return UnitInch
		}
		return UnitMillimeter
	case SnowLevel:
		if imperial {
			return UnitFoot
		}
		return UnitMeter
	case WindSpeed:
		if imperial {
			return UnitMilePerHour
		}
		return UnitMeterPerSecond
	}
	return ""
}

// Unit is the unit of a single quantity, such as a probability threshold.
type Unit string

// The units a threshold may be given in.
const (
	UnitMillimeter     Unit = "mm"
	UnitInch           Unit = "in"
	UnitCelsius        Unit = "degC"
	UnitFahrenheit     Unit = "degF"
	UnitKelvin         Unit = "K"
	UnitMeter          Unit = "m"
	UnitFoot           Unit = "ft"
	UnitMeterPerSecond Unit = "m/s"
	UnitMilePerHour    Unit = "mph"
)

// Symbol returns the unit's display symbol, such as "°C" or "in".
func (u Unit) Symbol() string {
	switch u {
	case UnitCelsius:
		return "°C"
	case UnitFahrenheit:
		return "°F"
	default:
		return string(u)
	}
}

// Exact conversion factors (BEHAVIOR 7). They convert the caller's threshold
// input only; response values are never converted.
const (
	mmPerInch    = 25.4
	metersPerFt  = 0.3048
	mpsPerMph    = 0.44704
	kelvinOffset = 273.15
)

// dimension reports which physical dimension a unit measures.
func (u Unit) dimension() string {
	switch u {
	case UnitMillimeter, UnitInch:
		return "depth"
	case UnitCelsius, UnitFahrenheit, UnitKelvin:
		return "temperature"
	case UnitMeter, UnitFoot:
		return "height"
	case UnitMeterPerSecond, UnitMilePerHour:
		return "speed"
	}
	return ""
}

// convert converts v from unit from to unit to, which must share a dimension.
// A value already in the target unit is returned unchanged.
func convert(v float64, from, to Unit) float64 {
	if from == to {
		return v
	}
	switch from.dimension() {
	case "depth":
		if from == UnitInch {
			return v * mmPerInch
		}
		return v / mmPerInch
	case "height":
		if from == UnitFoot {
			return v * metersPerFt
		}
		return v / metersPerFt
	case "speed":
		if from == UnitMilePerHour {
			return v * mpsPerMph
		}
		return v / mpsPerMph
	case "temperature":
		c := v
		switch from {
		case UnitFahrenheit:
			c = (v - 32) * 5 / 9
		case UnitKelvin:
			c = v - kelvinOffset
		}
		switch to {
		case UnitFahrenheit:
			return c*9/5 + 32
		case UnitKelvin:
			return c + kelvinOffset
		}
		return c
	}
	return v
}

// Threshold is a probability threshold. A zero Unit means the value is
// already in the request's units; otherwise the SDK converts it to the
// request's units with exact factors before sending.
type Threshold struct {
	Value float64
	Unit  Unit
}

// Millimeters returns a threshold of v mm.
func Millimeters(v float64) Threshold { return Threshold{v, UnitMillimeter} }

// Inches returns a threshold of v in.
func Inches(v float64) Threshold { return Threshold{v, UnitInch} }

// Celsius returns a threshold of v °C.
func Celsius(v float64) Threshold { return Threshold{v, UnitCelsius} }

// Fahrenheit returns a threshold of v °F.
func Fahrenheit(v float64) Threshold { return Threshold{v, UnitFahrenheit} }

// Kelvin returns a threshold of v K.
func Kelvin(v float64) Threshold { return Threshold{v, UnitKelvin} }

// Meters returns a threshold of v m.
func Meters(v float64) Threshold { return Threshold{v, UnitMeter} }

// Feet returns a threshold of v ft.
func Feet(v float64) Threshold { return Threshold{v, UnitFoot} }

// MetersPerSecond returns a threshold of v m/s.
func MetersPerSecond(v float64) Threshold { return Threshold{v, UnitMeterPerSecond} }

// MilesPerHour returns a threshold of v mph.
func MilesPerHour(v float64) Threshold { return Threshold{v, UnitMilePerHour} }
