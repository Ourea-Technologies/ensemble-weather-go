// SPDX-License-Identifier: BSD-2-Clause

// Snow report: where is it snowing? Rank places by the next five days'
// snowfall, then read one place's season to date and its snow trend.
//
// Explore ranks places by a metric the server computes; Season is the
// observed season to date; SnowTrend (Plus or Pro) shows each recent run's
// daily snowfall, so you can see a storm firm up or fade. Anonymous and Free
// callers get the first two.
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
	client := ew.NewClient(ew.WithUnits(ew.UnitsImperial))

	ranked, err := client.Explore(ctx, ew.ExploreSnow5d, &ew.ExploreParams{Limit: 5})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Most snow in the next 5 days (%s):\n", ranked.Unit)
	for _, item := range ranked.Items {
		lowHigh := ""
		if item.P10 != nil && item.P90 != nil {
			lowHigh = fmt.Sprintf(" (%.1f to %.1f)", *item.P10, *item.P90)
		}
		fmt.Printf("  %d. %s, %s: %.1f%s\n", item.Rank, item.Location.Name, item.Location.State, item.Value, lowHigh)
	}

	season, err := client.Season(ctx, "or/mt-hood-meadows", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nWater year %d at %s, through %s:\n", season.WaterYear, season.Location.Name, season.Through)
	if total := season.Precipitation; total != nil {
		normal := ""
		if total.PercentOfNormal != nil {
			normal = fmt.Sprintf(", %d%% of normal", *total.PercentOfNormal)
		}
		fmt.Printf("  precipitation %g %s%s\n", total.Amount, total.Unit, normal)
	}

	trend, err := client.SnowTrend(ctx, "or/mt-hood-meadows", nil)
	switch {
	case errors.Is(err, ew.ErrPermission):
		fmt.Println("\nThe snow trend is part of Plus and Pro (set OUREA_API_KEY).")
		return
	case err != nil:
		log.Fatal(err)
	}
	grid := trend.Grids[0]
	if len(grid.Runs) == 0 {
		return
	}
	newest := grid.Runs[len(grid.Runs)-1]
	fmt.Printf("\nSnow trend (%s), newest run %s:\n", grid.Source, newest.InitializedAt.Format("Mon 15:04Z"))
	for _, day := range newest.Days {
		switch {
		case day == nil:
		case day.Amount == nil:
			fmt.Printf("  %s: none\n", day.Date)
		default:
			fmt.Printf("  %s: %g\n", day.Date, *day.Amount)
		}
	}
}
