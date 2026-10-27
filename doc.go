// SPDX-License-Identifier: BSD-2-Clause

// Package ensembleweather is the official Go SDK for the Ourea probabilistic
// weather API (https://oureawx.com).
//
// Every forecast value is a distribution ([Quantiles]) computed by the server
// from the full ensemble. The SDK carries those distributions; it never sums,
// averages or interpolates them. Ask the server for what you need instead:
//
//   - a probability, with a [Chance] in [ForecastParams.Chances] (the chance
//     of at least 6 in of snow in the next 24 h, for example);
//   - a storm total, with a [Total] in [ForecastParams.Totals] (the
//     distribution of each member's total over a window).
//
// # Quick start
//
//	client := ensembleweather.NewClient() // anonymous unless OUREA_API_KEY is set
//	f, err := client.Forecast(ctx, "or/mt-hood-meadows", nil)
//
// # Authentication
//
// Pass a key with [WithAPIKey] or set OUREA_API_KEY. With no key, calls are
// anonymous (the Free tier). Forecasts at coordinates ([Client.ForecastAt])
// need a Pro key. Search, catalog details and attribution are by default sent
// without the key so their responses are shared at the edge cache; use
// [WithKeyOnOpenRoutes] to attribute every request to the key.
//
// # Errors and retries
//
// Every failed call returns an *[Error]. Test its category with errors.Is and a
// sentinel ([ErrNotFound], [ErrRateLimited], ...), and read its Status, Code,
// Message, RequestID and RetryAfter with errors.As. Connection errors, 429,
// 502, 503 and 504 are retried (2 retries by default) with full-jitter
// backoff, honouring Retry-After up to 60 s.
//
// # Caching
//
// The client keeps an in-memory ETag cache (256 entries, 32 MiB) and
// revalidates every repeated request with If-None-Match; a 304 returns the
// cached body with Revalidated set. Replace or disable it with [WithCache].
//
// # Paging
//
// List calls return one page and a NextCursor. On Go 1.23 and later,
// [Client.AllCatalogLocations] and its siblings return an iter.Seq2 that
// follows the cursor for you.
//
// Guides: https://docs.oureawx.com and https://docs.oureawx.com/go.
package ensembleweather
