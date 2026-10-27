# Changelog

All notable changes to this SDK are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html). The SDK version is
independent of the API version; every 0.x release targets API contract v1.

## [Unreleased]

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
