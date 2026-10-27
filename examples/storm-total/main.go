// SPDX-License-Identifier: BSD-2-Clause

// Storm total: the distribution of each ensemble member's total precipitation
// and snowfall over a window, computed by the server. Never add up hourly
// quantiles yourself: the sum of hourly medians is not the median total.
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
		MaxHours: 72,
		Totals: []ew.Total{
			{Variable: ew.Snowfall, Window: ew.Hours(72)},
			{Variable: ew.Precipitation, Window: ew.Hours(72)},
			{Variable: ew.Snowfall, Window: ew.HoursBetween(24, 48)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range f.WindowTotals {
		fmt.Printf("%s, hours %d-%d: ", t.Variable, t.WindowStartHours, t.WindowEndHours)
		if t.Quantiles == nil {
			fmt.Println("not covered by this forecast")
			continue
		}
		median, _ := t.Quantiles.Median()
		r, _ := t.Quantiles.Range80()
		fmt.Printf("median %g %s, 80%% range %s %s (%d members)\n", median, t.Unit.Symbol(), r, t.Unit.Symbol(), t.Members)
	}
}
