// SPDX-License-Identifier: BSD-2-Clause

// Error handling: catch an out-of-coverage error by its code, tell error
// categories apart with errors.Is, and read the rate-limit state.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	ew "oureawx.com/go/ensemble-weather"
)

func main() {
	ctx := context.Background()
	client := ew.NewClient() // OUREA_API_KEY, if set, is used

	// Northern Italy is outside the forecast domain (CONUS).
	_, err := client.ForecastAt(ctx, ew.Point{Lat: 45.0, Lon: 10.0}, nil)
	var apiErr *ew.Error
	switch {
	case errors.As(err, &apiErr) && apiErr.Code == ew.CodeOutsideCoverage:
		fmt.Printf("outside coverage: %s (status %d, request %s; see %s)\n", apiErr.Message, apiErr.Status, apiErr.RequestID, apiErr.DocURL)
	case errors.Is(err, ew.ErrPermission):
		fmt.Println("coordinate forecasts need a Pro key:", err)
	case errors.Is(err, ew.ErrRateLimited):
		fmt.Printf("rate limited; retry after %s\n", apiErr.RetryAfter)
	case err != nil:
		log.Fatal(err)
	}

	f, err := client.Forecast(ctx, "or/mt-hood-meadows", nil)
	if err != nil {
		log.Fatal(err)
	}
	if rl := f.RateLimit; rl != nil {
		fmt.Printf("rate limit: %d of %d left, resets in %d s (policy %s)\n", rl.Remaining, rl.Limit, rl.Reset, rl.Policy)
	} else {
		fmt.Println("no rate-limit headers: this anonymous response is shared at the edge cache")
	}
}
