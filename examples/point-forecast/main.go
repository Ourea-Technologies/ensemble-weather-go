// SPDX-License-Identifier: BSD-2-Clause

// Point forecast: a forecast at coordinates with an explicit elevation (the
// top of a ski lift). Forecasts at coordinates need a Pro key, read here from
// OUREA_API_KEY; without one the example says so and exits.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ew "oureawx.com/go/ensemble-weather"
)

func main() {
	if os.Getenv("OUREA_API_KEY") == "" {
		fmt.Println("point-forecast: set OUREA_API_KEY to a Pro key to run this example")
		return
	}
	client := ew.NewClient() // reads OUREA_API_KEY

	at := ew.Point{Lat: 45.3311, Lon: -121.6649, ElevationM: ew.Ptr(2050.0)}
	f, err := client.ForecastAt(context.Background(), at, &ew.ForecastParams{MaxHours: 48})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s, %d members from %d models (run %s)\n",
		f.Location.Name, f.ModelRun.TotalMembers, len(f.ModelRun.Models), f.ModelRun.InitializedAt.Format("2006-01-02 15Z"))
	wind := f.Units.For(ew.WindSpeed).Symbol()
	for _, hour := range f.Forecast.Timeseries {
		median, ok := hour.Ensemble.WindSpeed.Median()
		if !ok {
			continue
		}
		fmt.Printf("%s  wind median %.1f %s", hour.ValidTime.Format("Mon 15:04"), median, wind)
		if r, ok := hour.Ensemble.WindSpeed.Range80(); ok {
			fmt.Printf(" (p10-p90 %s)", r)
		}
		fmt.Println()
	}
	if f.RateLimit != nil {
		fmt.Printf("rate limit: %d of %d left\n", f.RateLimit.Remaining, f.RateLimit.Limit)
	}
}
