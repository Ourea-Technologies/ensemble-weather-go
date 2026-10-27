// SPDX-License-Identifier: BSD-2-Clause

// Quickstart: an anonymous forecast for Mt. Hood Meadows. It prints the next
// 24 hours' temperature as a range (p10 to p90) and the chance of at least
// 6 in (15 cm) of snow in the next 24 hours. No API key is needed.
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
