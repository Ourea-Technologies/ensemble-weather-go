// SPDX-License-Identifier: BSD-2-Clause

// Chance of snow: several probability terms in one call. The server computes
// each probability from every ensemble member; each answer comes back in the
// order asked, with its window and the number of members that voted.
package main

import (
	"context"
	"fmt"
	"log"

	ew "oureawx.com/go/ensemble-weather"
)

func main() {
	client := ew.NewClient()

	chances := []ew.Chance{
		{Variable: ew.Snowfall, Op: ew.AtLeast, Threshold: ew.Inches(6), Window: ew.Hours(24)},
		{Variable: ew.Snowfall, Op: ew.AtLeast, Threshold: ew.Inches(12), Window: ew.Hours(72)},
		{Variable: ew.Snowfall, Op: ew.AtLeast, Threshold: ew.Inches(6), Window: ew.HoursBetween(24, 48)},
		// "all": every hour of the first 12 stays at or below freezing.
		{Variable: ew.Temperature, Op: ew.AtMost, Threshold: ew.Fahrenheit(32), Window: ew.Hours(12), Aggregation: ew.AggregationAll},
	}
	f, err := client.Forecast(context.Background(), "or/mt-hood-meadows", &ew.ForecastParams{
		Units:   ew.UnitsImperial,
		Chances: chances,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range f.Probabilities {
		chance := "not covered by this forecast" // absent, never 0
		if p.Probability != nil {
			chance = fmt.Sprintf("%.0f%%", *p.Probability*100)
		}
		fmt.Printf("%s %s %g %s, hours %d-%d (%s): %s, %d members\n",
			p.Variable, p.Op, p.Threshold, p.ThresholdUnit.Symbol(),
			p.WindowStartHours, p.WindowEndHours, p.Aggregation, chance, p.Members)
	}
}
