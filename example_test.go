// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	ensembleweather "oureawx.com/go/ensemble-weather"
)

// An anonymous forecast for a named location: no API key needed.
func Example() {
	client := ensembleweather.NewClient()
	f, err := client.Forecast(context.Background(), "or/mt-hood-meadows", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, hour := range f.Forecast.Timeseries {
		if r, ok := hour.Ensemble.Temperature.Range80(); ok {
			fmt.Println(hour.ValidTime, r, f.Units.For(ensembleweather.Temperature).Symbol())
		}
	}
}

func ExampleNewClient() {
	client := ensembleweather.NewClient(
		ensembleweather.WithAPIKey(os.Getenv("MY_OUREA_KEY")),
		ensembleweather.WithUnits(ensembleweather.UnitsImperial),
		ensembleweather.WithHTTPClient(&http.Client{Transport: http.DefaultTransport}),
		ensembleweather.WithTimeout(10*time.Second),
		ensembleweather.WithUserAgent("my-app/1.0"),
	)
	fmt.Println(client) // the key is never printed
}

// Ask the server for probabilities; each answer carries its reference class.
func ExampleChance() {
	client := ensembleweather.NewClient()
	f, err := client.Forecast(context.Background(), "or/mt-hood-meadows", &ensembleweather.ForecastParams{
		Chances: []ensembleweather.Chance{
			{Variable: ensembleweather.Snowfall, Op: ensembleweather.AtLeast, Threshold: ensembleweather.Inches(6), Window: ensembleweather.Hours(24)},
			{Variable: ensembleweather.WindSpeed, Op: ensembleweather.AtLeast, Threshold: ensembleweather.MilesPerHour(40), Window: ensembleweather.HoursBetween(12, 36)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range f.Probabilities { // in the order asked
		if p.Probability == nil {
			fmt.Println(p.Variable, "not covered")
			continue
		}
		fmt.Printf("%s %s %g %s over hours %d-%d: %.0f%% of %d members\n", p.Variable, p.Op, p.Threshold,
			p.ThresholdUnit.Symbol(), p.WindowStartHours, p.WindowEndHours, *p.Probability*100, p.Members)
	}
}

// Window totals: the distribution of each member's own total over a window.
func ExampleTotal() {
	client := ensembleweather.NewClient()
	f, err := client.Forecast(context.Background(), "or/mt-hood-meadows", &ensembleweather.ForecastParams{
		Totals: []ensembleweather.Total{{Variable: ensembleweather.Snowfall, Window: ensembleweather.Hours(72)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	if q := f.WindowTotals[0].Quantiles; q != nil {
		r, _ := q.Range80()
		fmt.Println("72 h snowfall, p10-p90:", r, f.WindowTotals[0].Unit.Symbol())
	}
}

func ExampleQuantiles_Range80() {
	q := ensembleweather.Quantiles{P10: ensembleweather.Ptr(-4.0), P50: ensembleweather.Ptr(-2.0), P90: ensembleweather.Ptr(0.5)}
	if r, ok := q.Range80(); ok {
		fmt.Println(r)
	}
	// Output: -4–0.5
}

func ExampleError() {
	client := ensembleweather.NewClient()
	_, err := client.GetPointForecast(context.Background(), ensembleweather.Point{Lat: 45.0, Lon: 10.0}, nil)

	var apiErr *ensembleweather.Error
	switch {
	case errors.As(err, &apiErr) && apiErr.Code == ensembleweather.CodeOutsideCoverage:
		fmt.Println("outside the forecast domain:", apiErr.Message)
	case errors.Is(err, ensembleweather.ErrPermission):
		fmt.Println("needs a Pro key")
	case errors.Is(err, ensembleweather.ErrRateLimited):
		fmt.Println("retry after", apiErr.RetryAfter)
	case err != nil:
		log.Fatal(err)
	}
}

func ExampleClient_With() {
	client := ensembleweather.NewClient()
	// A per-customer key, sharing the client's cache and connections.
	customer := client.With(ensembleweather.WithAPIKey("ow_customer_key"))
	_, _ = customer.Forecast(context.Background(), "co/vail", nil)
}
