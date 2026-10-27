// SPDX-License-Identifier: BSD-2-Clause

// Units: the same forecast in metric and imperial. The server converts every
// value; the SDK only labels them.
package main

import (
	"context"
	"fmt"
	"log"

	ew "oureawx.com/go/ensemble-weather"
)

func main() {
	ctx := context.Background()
	client := ew.NewClient()

	for _, units := range []ew.Units{ew.UnitsMetric, ew.UnitsImperial} {
		f, err := client.Forecast(ctx, "or/mt-hood-meadows", &ew.ForecastParams{Units: units})
		if err != nil {
			log.Fatal(err)
		}
		if len(f.Forecast.Timeseries) == 0 {
			continue
		}
		first := f.Forecast.Timeseries[0].Ensemble
		temp, _ := first.Temperature.Median()
		precip, _ := first.Precipitation.Median()
		fmt.Printf("%-8s temperature median %.1f %s, precipitation median %.2f %s\n", f.Units,
			temp, f.Units.For(ew.Temperature).Symbol(), precip, f.Units.For(ew.Precipitation).Symbol())
	}
}
