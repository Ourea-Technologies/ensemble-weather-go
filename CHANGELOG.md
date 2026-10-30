# Changelog

All notable changes to this SDK are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html). The SDK version is
independent of the API version; every 0.x release targets API contract v1.

## [Unreleased]

Moves to the v1 contract as deployed (ensemble-weather-spec 8b85fea, the
server's public spec byte for byte). Breaking, as 0.x allows.

### Changed

- `Client.ForecastAt` is renamed `Client.GetPointForecast`, matching the
  operation `getPointForecast`; there is no alias.
- `precip_*` fields are `precipitation_*`: `DailyForecast.TotalPrecipitation`,
  `TotalPrecipitationIfWet`, `PrecipitationProbability`,
  `PrecipitationOccurrenceProbability`, `PrecipitationType`,
  `PrecipitationDisplay`; `TimeSeriesPoint.PrecipitationType`;
  `CurrentConditions.PrecipitationProbability`;
  `SensorObservation.PrecipitationRate`; and
  `ProbabilityThresholds.DailyPrecipitationProbability` /
  `DailyPrecipitationOccurrenceProbability`.
- `EnsembleStats.WindDirection` and `ModelMean.WindDirection` are
  `CircularStats` (mean, stddev, 50 % and 90 % sectors), not `Quantiles`.
- Probabilities and window totals are matched to the caller's terms by the
  `term` the server echoes, never by position. `ProbabilityResult` and
  `WindowTotal` carry `Term` and `NullReason`; `WindowTotal` carries
  `Aggregation` (new `AggregationMin` / `AggregationMax` values).
- Error codes are the public spec's 16: `CodeAuthUnavailable` (unavailable)
  is new; `CodeForbidden`, `CodeModelRunNotFound`, `CodeGone`,
  `CodeInvalidMaxHours`, `CodeAmbiguous` and `CodeFeelsLikeUnavailable` are
  gone (an old code still keeps its string and maps by status).
  `OUTSIDE_COVERAGE` is a 404. `Retry-After` is honoured on 504 as well as
  429 and 503.
- `ActiveAlert.Severity` and `Urgency` are typed snake_case enums, and
  catalog enums are snake_case and typed: `StationNetwork`, `StationStatus`,
  `TemplateVariant`, `CatalogLocationsParams.Type` / `Template`,
  `CatalogStationsParams.Networks` / `Status`. `network` and `capability`
  are sent as one sorted, de-duplicated comma list (the server's canonical
  form); `capability` is no longer repeated.
- `LocationRef.Slug` and `Type` are empty for a coordinate forecast that
  matched no named location (the server sends null), and `Key()` is then "".

### Added

- `ActiveAlert.Certainty`, `CurrentStation.StationID`,
  `RunsResponse.NextCursor`, `ForecastParams.Include` (`mslp`,
  `skin_temperature`), `ConditionIcon` and `NullReason` constants, and
  `Error.DocURL`, `Error.Param` and `Error.NearestLocation`;
  `Error.RequestID` is read from the body's `request_id` first.
- A drift test that fails on any model field the spec does not have.

### Removed

- `StormReports`, `StormReport` and their models (not in the public spec).
- Catalog ads (`CatalogAdSlot` and its placements; `include=ads`).
- `EnsembleStats.FeelsLikeUTCI`, `Ptype` and `WindDirectionCircular`;
  `DailyForecast.FeelsLikeUTCIHigh` / `Low`; `ElevationForecast.Default`
  and `Timeseries` (`ElevationHour`).

## [0.1.0] - 2026-10-26

### Added

- `Client` for API v1 with functional options (`WithAPIKey`, `WithBaseURL`,
  `WithHTTPClient`, `WithUnits`, `WithTimeout`, `WithMaxRetries`,
  `WithCache`, `WithKeyOnOpenRoutes`, `WithUserAgent`, `WithLogger`) and
  `OUREA_API_KEY` / `OUREA_BASE_URL` from the environment.
- Every v1 operation except `getMapsToken`: `Forecast`, `ForecastAt`,
  `Current`, `CurrentAt`, `Alerts`, `AlertsAt`, `Nearby`, `NearbyAt`,
  `Search`, `Locations`, `CatalogLocations`, `CatalogLocation`,
  `CatalogStations`, `CatalogStation`, `Runs`, `StormReports`,
  `StormReport` and `Attribution`, with hand-written models for every
  schema they reach.
- Server-computed probabilities (`Chance`, `prob=`) and window totals
  (`Total`, `totals=`), answered in the caller's order; threshold unit
  conversion with the API's exact factors and rounding.
- Canonical URLs (sorted parameters, defaults omitted, RFC 3986 encoding),
  keyless open routes, same-host redirects (the key is dropped on any other
  host), retries with full-jitter backoff and `Retry-After`, an in-memory
  ETag cache, rate-limit state, deprecation warnings through `log/slog`, and
  `*Error` with sentinel errors per category.
- Page calls plus Go 1.23 iterators (`AllLocations`, `AllCatalogLocations`,
  `AllCatalogStations`).
- The shared test vectors (43/43), a spec-drift test module, and runnable
  examples.

[Unreleased]: https://github.com/Ourea-Technologies/ensemble-weather-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Ourea-Technologies/ensemble-weather-go/releases/tag/v0.1.0
