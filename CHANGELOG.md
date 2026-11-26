# Changelog

All notable changes to this SDK are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html). The SDK version is
independent of the API version; every 0.x release targets API contract v1.

## [Unreleased]

## [0.1.0] - unreleased

The first tagged release, at `oureawx.com/go/ensemble-weather@v0.1.0`.
Targets the v1 contract at ensemble-weather-spec 7954268 (the server's
public spec at 256cd770, byte for byte). It folds in the October draft of
0.1.0, which was never tagged.

### Added

- `Client` for API v1 with functional options (`WithAPIKey`, `WithBaseURL`,
  `WithHTTPClient`, `WithUnits`, `WithTimeout`, `WithMaxRetries`,
  `WithCache`, `WithKeyOnOpenRoutes`, `WithUserAgent`, `WithLogger`) and
  `OUREA_API_KEY` / `OUREA_BASE_URL` from the environment.
- Every v1 operation except `getMapsToken`: `Forecast`, `GetPointForecast`,
  `Current`, `GetPointCurrent`, `Alerts`, `GetPointAlerts`, `Nearby`,
  `GetPointNearby`, `Search`, `Locations`, `CatalogLocations`,
  `CatalogLocation`, `CatalogStations`, `CatalogStation`, `Runs` and
  `Attribution`, and the routes added to the contract since October:
  - `Explore` and its iterator `AllExplore`: places ranked by an
    `ExploreMetric` (`ExploreSnow5d`, `ExploreBaseDepth`,
    `ExploreSeasonPrecip`, ...), with `Order`, `Type`, `State`, `Near`,
    `RadiusKm`, elevation bounds, `AllPoints` and cursor paging;
  - `Season` (observed season-to-date precipitation and snowfall),
    `SnowTrend` (each recent run's daily snowfall; Plus or Pro, else
    `PLUS_REQUIRED`), `Snowpack` and `SnowpackPlume` (beta);
  - `Composition` (an open route, sent without the key);
  - the GeoJSON routes `AlertOutlines` (filtered by `Classes` and a `BBox`)
    and `AvalancheZones`, passed through with the API's keys untouched.
- `Search` takes `Sort`, `Order`, `Near`, `RadiusKm` and `Units`; with a
  metric sort each hit carries `Explore`.
- `ForecastParams.Members` is a `MembersMode`: `MembersOn` (each hour's
  members) or `MembersAsPlume` (the plume view, columns in
  `ForecastData.Plume`); any other value is refused before sending.
- Models for every schema the operations reach, including the new forecast
  shapes (`ForecastLocation`, `MemberPlume`, `PrecipOutlook`,
  `SnowOnGround`, `AvalancheForecast`, `MembersSpread`, day normals,
  departures and percentiles, `snow_level_wet_bulb`, `snow_character`),
  `CurrentResponse.Summary` and `Pass` (a mountain pass's forecast), and
  the new `ActiveAlert` fields (`EventKey`, `VTECAction`, `Description`,
  ...). New enums `SnowCharacter`, `AlertClass`, `PassStatusLevel`,
  `ExploreMetric` and `Order`; new values in `Condition`, `ConditionIcon`,
  `MembersView` (`MembersPool`), `SourceKind` and `StationNetwork`; the
  error code `CodePlusRequired` (a permission error).
- Server-computed probabilities (`Chance`, `prob=`) and window totals
  (`Total`, `totals=`), matched to the caller's terms by the `term` the
  server echoes; threshold unit conversion with the API's exact factors and
  rounding.
- Canonical URLs (sorted parameters, defaults omitted, RFC 3986 encoding,
  comma lists sorted and de-duplicated with literal commas), keyless open
  routes, same-host redirects (the key is dropped on any other host),
  retries with full-jitter backoff and `Retry-After`, an in-memory ETag
  cache, rate-limit state, deprecation warnings through `log/slog`, and
  `*Error` with sentinel errors per category.
- Page calls plus Go 1.23 iterators (`AllLocations`, `AllCatalogLocations`,
  `AllCatalogStations`, `AllExplore`).
- The shared test vectors (58/58), a spec-drift test module that also
  checks every field name against the BEHAVIOR.md 3.0 rename table, and
  runnable examples (new: `examples/snow-report`).

### Changed since the October draft

- Stuttering JSON fields take the SDK names of BEHAVIOR.md 3.0 (the struct
  tag keeps the JSON name): `ForecastResponse.Series` (JSON `forecast`) and
  `Elevations` (`elevation_forecasts`), `CurrentResponse.Now` (`current`),
  `CurrentConditions.Code` / `Text` / `Icon`, `CurrentStation.ID`,
  `ProbabilityResult.Value`, `ProbabilityThresholds.DailyPrecipitation` /
  `DailyPrecipitationOccurrence` / `DailySnow`, `WindowTotal.StartHours` /
  `EndHours`, `ModelMembers.Items`, `MemberForecast.ID`, `Items` and `Count`
  on the alerts, locations, nearby and runs lists, `NearbyLocationsResponse.Origin`,
  an elevation's `Height` (`HeightM` in the catalog),
  `CatalogStation.SourceID` / `Status` / `URL`, `CatalogNearbyStation.Status`,
  `CatalogTemplate.Variant`, `CatalogContentEntry.Locale` and
  `SensorObservation.ObservedAt`. The README lists them all.
- `ForecastResponse.Location` is a `ForecastLocation` (the `LocationRef`
  plus the served `Elevation` and the named `Elevations`).
  `NamedElevation` is now `ElevationSummary`. `MembersCumulative` holds
  `MemberQuantiles`.
- Comma lists (`sources`, `exclude_sources`, `elevations`, `diagnostics`,
  `include`, `feels_like`, `network`, `capability`, `class`) are sent with
  literal commas, not `%2C`, and are always sorted and de-duplicated
  (BEHAVIOR.md 3.1 rule 5). A two-letter `state` is lowercased in a query
  too.
- The coordinate methods follow the operation ids: `ForecastAt` is
  `GetPointForecast`, `CurrentAt` is `GetPointCurrent`, `AlertsAt` is
  `GetPointAlerts` and `NearbyAt` is `GetPointNearby`. There are no aliases.

### Removed

- `StormReports`, `StormReport` and their models (not in the public spec),
  and catalog ads (`include=ads`).

[Unreleased]: https://github.com/Ourea-Technologies/ensemble-weather-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Ourea-Technologies/ensemble-weather-go/releases/tag/v0.1.0
