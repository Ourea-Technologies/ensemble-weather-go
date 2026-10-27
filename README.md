# ensemble-weather-go

Official Go SDK for the [Ourea](https://oureawx.com) probabilistic weather API.

[![CI](https://github.com/Ourea-Technologies/ensemble-weather-go/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Ourea-Technologies/ensemble-weather-go/actions/workflows/ci.yml?query=branch%3Amain)
[![Security](https://github.com/Ourea-Technologies/ensemble-weather-go/actions/workflows/security.yml/badge.svg?branch=main)](https://github.com/Ourea-Technologies/ensemble-weather-go/actions/workflows/security.yml?query=branch%3Amain)
[![Go Reference](https://pkg.go.dev/badge/oureawx.com/go/ensemble-weather.svg)](https://pkg.go.dev/oureawx.com/go/ensemble-weather)
[![Version](https://img.shields.io/github/v/tag/Ourea-Technologies/ensemble-weather-go?sort=semver&label=version)](https://pkg.go.dev/oureawx.com/go/ensemble-weather?tab=versions)
[![Go](https://img.shields.io/github/go-mod/go-version/Ourea-Technologies/ensemble-weather-go?label=go)](go.mod)
[![Licence: BSD-2-Clause](https://img.shields.io/badge/licence-BSD--2--Clause-blue.svg)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-docs.oureawx.com-informational)](https://docs.oureawx.com/go)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/Ourea-Technologies/ensemble-weather-go/badge)](https://scorecard.dev/viewer/?uri=github.com/Ourea-Technologies/ensemble-weather-go)

## Install

```sh
go get oureawx.com/go/ensemble-weather
```

## Quickstart

No signup and no key: an anonymous forecast for Mt. Hood Meadows, printing
the next 24 hours' temperature as a range (p10 to p90) and the chance of at
least 6 in (15 cm) of snow in the next 24 hours. Save it as `main.go` in a new
module (`go mod init example`, then the `go get` above) and `go run .`:

```go
package main

import (
	"context"
	"fmt"
	"log"

	ew "oureawx.com/go/ensemble-weather"
)

func main() {
	client := ew.NewClient()

	f, err := client.Forecast(context.Background(), "or/mt-hood-meadows", &ew.ForecastParams{
		Chances: []ew.Chance{
			{Variable: ew.Snowfall, Op: ew.AtLeast, Threshold: ew.Inches(6), Window: ew.Hours(24)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	unit := f.Units.For(ew.Temperature).Symbol()
	for _, hour := range f.Forecast.Timeseries {
		if r, ok := hour.Ensemble.Temperature.Range80(); ok {
			fmt.Printf("%s  %s %s\n", hour.ValidTime.Format("Mon 15:04 MST"), r, unit)
		}
	}
	for _, c := range f.Probabilities {
		if c.Probability != nil {
			fmt.Printf("Chance of >= %g %s of snow in %d-%d h: %.0f%% (%d members)\n",
				c.Threshold, c.ThresholdUnit.Symbol(), c.WindowStartHours, c.WindowEndHours, *c.Probability*100, c.Members)
		}
	}
}
```

This is [`examples/quickstart/main.go`](examples/quickstart/main.go) without
its header comment; CI runs it on every change.

## Authentication

Pass a key to the constructor, or set `OUREA_API_KEY` and call
`NewClient()` with no options:

```go
client := ew.NewClient(ew.WithAPIKey("ow_..."))
```

| | Anonymous / Free key | Pro key |
|---|---|---|
| Named-location forecasts, current conditions, alerts | yes, 10-day (240 h) horizon | yes, 15-day (360 h) horizon |
| Forecasts and current conditions at coordinates | no | yes |
| Per-model breakdowns, historical runs | no | yes |
| Ensemble members | plume view | every variable |

Plans and limits: https://oureawx.com/pricing. The key is sent only as
`X-API-Key`, only to the configured base URL, and never appears in errors,
logs or `fmt` output. Search, catalog details and attribution are sent
without the key so every caller shares one edge-cache entry;
`WithKeyOnOpenRoutes()` sends it there too.

## Probabilities, not single numbers

Every forecast value is a distribution, `Quantiles`, computed by the server
from the full ensemble (up to five weather models). Ask the server for the
number you need; the SDK never sums, averages or interpolates quantiles
itself, because only the server has every member. (Summing hourly medians,
for instance, does not give the median storm total.) More:
https://docs.oureawx.com/concepts/probabilities.

**A chance**, with its reference class (variable, threshold, window, members):

```go
f, err := client.Forecast(ctx, "or/mt-hood-meadows", &ew.ForecastParams{
	Chances: []ew.Chance{
		{Variable: ew.Snowfall, Op: ew.AtLeast, Threshold: ew.Inches(12), Window: ew.Hours(72)},
		{Variable: ew.Temperature, Op: ew.AtMost, Threshold: ew.Fahrenheit(32), Window: ew.HoursBetween(6, 18), Aggregation: ew.AggregationAll},
	},
})
// f.Probabilities[i] answers Chances[i]; Probability is nil (never 0) when the forecast does not reach the window.
```

**A window total**, the distribution of each member's own total:

```go
f, err := client.Forecast(ctx, "or/mt-hood-meadows", &ew.ForecastParams{
	Totals: []ew.Total{{Variable: ew.Snowfall, Window: ew.Hours(72)}},
})
if q := f.WindowTotals[0].Quantiles; q != nil {
	median, _ := q.Median()
	r, _ := q.Range80() // p10..p90
	fmt.Println(median, r, f.WindowTotals[0].Unit.Symbol())
}
```

**Quantiles**: the percentiles the server sent, as pointers (nil means
absent), with the served ranges as views:

```go
t := f.Forecast.Timeseries[0].Ensemble.Temperature
median, ok := t.Median() // p50: the median, not "the forecast"
r80, ok := t.Range80()   // p10..p90
r50, ok := t.Range50()   // p25..p75
```

Thresholds may carry any unit (`ew.Inches`, `ew.Celsius`, `ew.MilesPerHour`,
...); the SDK converts your threshold to the request's units with exact
factors and rounds it as the server does. Response values are never
converted: choose the unit system and the server converts.

| Units | Temperature | Precipitation, snowfall | Wind | Heights |
|---|---|---|---|---|
| `ew.UnitsMetric` (default) | °C | mm | m/s | m |
| `ew.UnitsImperial` | °F | in | mph | ft |
| `ew.UnitsKelvin` | K | mm | m/s | m |

## More examples

Each is a complete program; run it with `go run ./examples/<name>`.

| Example | Shows |
|---|---|
| [quickstart](examples/quickstart/main.go) | the program above, anonymous |
| [point-forecast](examples/point-forecast/main.go) | a forecast at coordinates with elevation (Pro key from `OUREA_API_KEY`) |
| [chance-of-snow](examples/chance-of-snow/main.go) | several `Chance` terms in one call, with windows and member counts |
| [storm-total](examples/storm-total/main.go) | `Total` window distributions for precipitation and snowfall |
| [search-and-catalog](examples/search-and-catalog/main.go) | search, then page through the catalog with the cursor |
| [current-and-alerts](examples/current-and-alerts/main.go) | current conditions and active alerts for a location |
| [error-handling](examples/error-handling/main.go) | catching an out-of-coverage error by its code; reading rate-limit state |
| [units](examples/units/main.go) | the same forecast in metric and imperial |

On Go 1.23+, list calls also come as iterators that follow the cursor:

```go
for loc, err := range client.AllCatalogLocations(ctx, nil) {
	if err != nil {
		return err
	}
	fmt.Println(loc.Key())
}
```

## Errors, retries and rate limits

Every call returns `error`; failures are `*ew.Error` values with `Status`,
`Code` (the API's code, kept verbatim), `Message`, `RequestID`, `RetryAfter`
and `RateLimit`. Test the category with `errors.Is` and a sentinel, and read
the details with `errors.As`:

```go
_, err := client.ForecastAt(ctx, ew.Point{Lat: 45.0, Lon: 10.0}, nil)
var apiErr *ew.Error
switch {
case errors.As(err, &apiErr) && apiErr.Code == ew.CodeOutsideCoverage:
	// outside the forecast domain
case errors.Is(err, ew.ErrRateLimited):
	time.Sleep(apiErr.RetryAfter)
}
```

Sentinels: `ErrAuthentication`, `ErrPermission`, `ErrNotFound`,
`ErrInvalidRequest`, `ErrOutsideCoverage`, `ErrRateLimited`,
`ErrUnavailable`, `ErrConnection`, and `ErrUsage` (refused before sending,
such as an invalid `Chance`).

Connection errors (except TLS verification failures), 429, 502, 503 and 504
are retried twice by default (`WithMaxRetries`), with full-jitter backoff; a
429 or 503 `Retry-After` is honoured exactly when it is 60 s or less, and
returned in `RetryAfter` when it is longer. Nothing else is retried. Each
attempt times out after 30 s (`WithTimeout`), and `ctx` bounds the whole call.

Every result embeds `ResponseMeta`: `f.RateLimit` holds the response's
`RateLimit-*` headers (nil on edge-cacheable responses, which carry none),
and `f.Revalidated` reports a `304` served from the client's ETag cache
(256 entries / 32 MiB in memory by default; `WithCache` replaces or, with
nil, disables it).

## Documentation

- Guides: https://docs.oureawx.com
- This SDK's guide: https://docs.oureawx.com/go
- API reference for this package: https://pkg.go.dev/oureawx.com/go/ensemble-weather
- The OpenAPI contract (API v1): [ensemble-weather-spec](https://github.com/Ourea-Technologies/ensemble-weather-spec), pinned here in [`spec/`](spec/)

This SDK targets API contract v1. Its own version (0.x) is independent of the
API version.

## Supported versions

Go 1.22 and the two latest Go releases, all tested in CI. The iterators
(`AllLocations`, `AllCatalogLocations`, `AllCatalogStations`) need Go 1.23;
the page calls work on every supported version. The SDK uses only the
standard library.

## Contributing, security, licence

- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security: report vulnerabilities privately, as [SECURITY.md](SECURITY.md) describes
- Licence: [BSD-2-Clause](LICENSE)
